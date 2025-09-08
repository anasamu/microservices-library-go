package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/sirupsen/logrus"
)

// Elasticsearch represents Elasticsearch client
type Elasticsearch struct {
	client *elasticsearch.Client
	config *ElasticsearchConfig
	logger *logrus.Logger
}

// ElasticsearchConfig holds Elasticsearch configuration
type ElasticsearchConfig struct {
	URL      string
	Username string
	Password string
	Index    string
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query        interface{}              `json:"query,omitempty"`
	Sort         []map[string]interface{} `json:"sort,omitempty"`
	From         int                      `json:"from,omitempty"`
	Size         int                      `json:"size,omitempty"`
	Source       interface{}              `json:"_source,omitempty"`
	Aggregations map[string]interface{}   `json:"aggs,omitempty"`
	Highlight    map[string]interface{}   `json:"highlight,omitempty"`
}

// SearchResponse represents a search response
type SearchResponse struct {
	Took     int64 `json:"took"`
	TimedOut bool  `json:"timed_out"`
	Hits     struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []struct {
			Index  string                 `json:"_index"`
			Type   string                 `json:"_type"`
			ID     string                 `json:"_id"`
			Score  float64                `json:"_score"`
			Source map[string]interface{} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations map[string]interface{} `json:"aggregations,omitempty"`
}

// Document represents a document to be indexed
type Document struct {
	ID     string                 `json:"id,omitempty"`
	Index  string                 `json:"index,omitempty"`
	Type   string                 `json:"type,omitempty"`
	Source map[string]interface{} `json:"source"`
}

// NewElasticsearch creates a new Elasticsearch client
func NewElasticsearch(config *ElasticsearchConfig, logger *logrus.Logger) (*Elasticsearch, error) {
	esConfig := elasticsearch.Config{
		Addresses: []string{config.URL},
	}

	if config.Username != "" && config.Password != "" {
		esConfig.Username = config.Username
		esConfig.Password = config.Password
	}

	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	es := &Elasticsearch{
		client: client,
		config: config,
		logger: logger,
	}

	// Test connection
	if err := es.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping Elasticsearch: %w", err)
	}

	es.logger.Info("Elasticsearch client initialized successfully")
	return es, nil
}

// Ping tests the connection to Elasticsearch
func (es *Elasticsearch) Ping() error {
	res, err := es.client.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Elasticsearch ping failed: %s", res.String())
	}

	return nil
}

// CreateIndex creates an index with mapping
func (es *Elasticsearch) CreateIndex(indexName string, mapping map[string]interface{}) error {
	exists, err := es.IndexExists(indexName)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	if exists {
		es.logger.Debugf("Index %s already exists", indexName)
		return nil
	}

	var body bytes.Buffer
	if mapping != nil {
		if err := json.NewEncoder(&body).Encode(mapping); err != nil {
			return fmt.Errorf("failed to encode mapping: %w", err)
		}
	}

	res, err := es.client.Indices.Create(
		indexName,
		es.client.Indices.Create.WithBody(&body),
	)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to create index: %s", res.String())
	}

	es.logger.Infof("Created index: %s", indexName)
	return nil
}

// DeleteIndex deletes an index
func (es *Elasticsearch) DeleteIndex(indexName string) error {
	res, err := es.client.Indices.Delete([]string{indexName})
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to delete index: %s", res.String())
	}

	es.logger.Infof("Deleted index: %s", indexName)
	return nil
}

// IndexExists checks if an index exists
func (es *Elasticsearch) IndexExists(indexName string) (bool, error) {
	res, err := es.client.Indices.Exists([]string{indexName})
	if err != nil {
		return false, fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

// IndexDocument indexes a document
func (es *Elasticsearch) IndexDocument(ctx context.Context, doc *Document) error {
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(doc.Source); err != nil {
		return fmt.Errorf("failed to encode document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      doc.Index,
		DocumentID: doc.ID,
		Body:       &body,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index document: %s", res.String())
	}

	es.logger.Debugf("Indexed document: %s/%s", doc.Index, doc.ID)
	return nil
}

// GetDocument retrieves a document by ID
func (es *Elasticsearch) GetDocument(ctx context.Context, indexName, docID string) (map[string]interface{}, error) {
	req := esapi.GetRequest{
		Index:      indexName,
		DocumentID: docID,
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, fmt.Errorf("document not found")
		}
		return nil, fmt.Errorf("failed to get document: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// UpdateDocument updates a document
func (es *Elasticsearch) UpdateDocument(ctx context.Context, indexName, docID string, doc map[string]interface{}) error {
	var body bytes.Buffer
	updateDoc := map[string]interface{}{
		"doc": doc,
	}
	if err := json.NewEncoder(&body).Encode(updateDoc); err != nil {
		return fmt.Errorf("failed to encode document: %w", err)
	}

	req := esapi.UpdateRequest{
		Index:      indexName,
		DocumentID: docID,
		Body:       &body,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to update document: %s", res.String())
	}

	es.logger.Debugf("Updated document: %s/%s", indexName, docID)
	return nil
}

// DeleteDocument deletes a document
func (es *Elasticsearch) DeleteDocument(ctx context.Context, indexName, docID string) error {
	req := esapi.DeleteRequest{
		Index:      indexName,
		DocumentID: docID,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return fmt.Errorf("document not found")
		}
		return fmt.Errorf("failed to delete document: %s", res.String())
	}

	es.logger.Debugf("Deleted document: %s/%s", indexName, docID)
	return nil
}

// Search performs a search query
func (es *Elasticsearch) Search(ctx context.Context, indexName string, req *SearchRequest) (*SearchResponse, error) {
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(req); err != nil {
		return nil, fmt.Errorf("failed to encode search request: %w", err)
	}

	searchReq := esapi.SearchRequest{
		Index: []string{indexName},
		Body:  &body,
	}

	res, err := searchReq.Do(ctx, es.client)
	if err != nil {
		return nil, fmt.Errorf("failed to perform search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search failed: %s", res.String())
	}

	var result SearchResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return &result, nil
}

// BulkIndex performs bulk indexing
func (es *Elasticsearch) BulkIndex(ctx context.Context, docs []*Document) error {
	var body bytes.Buffer

	for _, doc := range docs {
		// Index action
		action := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": doc.Index,
				"_id":    doc.ID,
			},
		}

		actionJSON, err := json.Marshal(action)
		if err != nil {
			return fmt.Errorf("failed to marshal action: %w", err)
		}

		body.Write(actionJSON)
		body.WriteString("\n")

		// Document source
		docJSON, err := json.Marshal(doc.Source)
		if err != nil {
			return fmt.Errorf("failed to marshal document: %w", err)
		}

		body.Write(docJSON)
		body.WriteString("\n")
	}

	req := esapi.BulkRequest{
		Body:    &body,
		Refresh: "true",
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to perform bulk index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk index failed: %s", res.String())
	}

	es.logger.Debugf("Bulk indexed %d documents", len(docs))
	return nil
}

// CreateMatchQuery creates a match query
func CreateMatchQuery(field, value string) map[string]interface{} {
	return map[string]interface{}{
		"match": map[string]interface{}{
			field: value,
		},
	}
}

// CreateTermQuery creates a term query
func CreateTermQuery(field, value string) map[string]interface{} {
	return map[string]interface{}{
		"term": map[string]interface{}{
			field: value,
		},
	}
}

// CreateRangeQuery creates a range query
func CreateRangeQuery(field string, gte, lte interface{}) map[string]interface{} {
	rangeQuery := map[string]interface{}{}
	if gte != nil {
		rangeQuery["gte"] = gte
	}
	if lte != nil {
		rangeQuery["lte"] = lte
	}

	return map[string]interface{}{
		"range": map[string]interface{}{
			field: rangeQuery,
		},
	}
}

// CreateBoolQuery creates a bool query
func CreateBoolQuery(must, should, mustNot []interface{}, filter []interface{}) map[string]interface{} {
	boolQuery := map[string]interface{}{}

	if len(must) > 0 {
		boolQuery["must"] = must
	}
	if len(should) > 0 {
		boolQuery["should"] = should
	}
	if len(mustNot) > 0 {
		boolQuery["must_not"] = mustNot
	}
	if len(filter) > 0 {
		boolQuery["filter"] = filter
	}

	return map[string]interface{}{
		"bool": boolQuery,
	}
}

// CreateMultiMatchQuery creates a multi-match query
func CreateMultiMatchQuery(query string, fields []string, queryType string) map[string]interface{} {
	multiMatch := map[string]interface{}{
		"query":  query,
		"fields": fields,
	}

	if queryType != "" {
		multiMatch["type"] = queryType
	}

	return map[string]interface{}{
		"multi_match": multiMatch,
	}
}

// CreateWildcardQuery creates a wildcard query
func CreateWildcardQuery(field, value string) map[string]interface{} {
	return map[string]interface{}{
		"wildcard": map[string]interface{}{
			field: value,
		},
	}
}

// CreateFuzzyQuery creates a fuzzy query
func CreateFuzzyQuery(field, value string, fuzziness string) map[string]interface{} {
	fuzzy := map[string]interface{}{
		"value": value,
	}

	if fuzziness != "" {
		fuzzy["fuzziness"] = fuzziness
	}

	return map[string]interface{}{
		"fuzzy": map[string]interface{}{
			field: fuzzy,
		},
	}
}

// CreateSort creates a sort clause
func CreateSort(field string, order string) map[string]interface{} {
	return map[string]interface{}{
		field: map[string]interface{}{
			"order": order,
		},
	}
}

// CreateAggregation creates an aggregation
func CreateAggregation(name string, aggType string, field string, size int) map[string]interface{} {
	agg := map[string]interface{}{
		aggType: map[string]interface{}{
			"field": field,
		},
	}

	if size > 0 {
		agg[aggType].(map[string]interface{})["size"] = size
	}

	return map[string]interface{}{
		name: agg,
	}
}

// HealthCheck performs a health check on Elasticsearch
func (es *Elasticsearch) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Test ping
	if err := es.Ping(); err != nil {
		return fmt.Errorf("elasticsearch health check failed: %w", err)
	}

	// Test basic operations
	testIndex := "health_check_test"
	testDoc := &Document{
		Index: testIndex,
		ID:    "test_doc",
		Source: map[string]interface{}{
			"test_field": "test_value",
			"timestamp":  time.Now(),
		},
	}

	// Create test index
	if err := es.CreateIndex(testIndex, nil); err != nil {
		return fmt.Errorf("failed to create test index: %w", err)
	}

	// Index test document
	if err := es.IndexDocument(ctx, testDoc); err != nil {
		return fmt.Errorf("failed to index test document: %w", err)
	}

	// Search test document
	searchReq := &SearchRequest{
		Query: CreateTermQuery("test_field", "test_value"),
		Size:  1,
	}

	if _, err := es.Search(ctx, testIndex, searchReq); err != nil {
		return fmt.Errorf("failed to search test document: %w", err)
	}

	// Delete test document
	if err := es.DeleteDocument(ctx, testIndex, testDoc.ID); err != nil {
		return fmt.Errorf("failed to delete test document: %w", err)
	}

	// Delete test index
	if err := es.DeleteIndex(testIndex); err != nil {
		return fmt.Errorf("failed to delete test index: %w", err)
	}

	return nil
}

// Close closes the Elasticsearch client
func (es *Elasticsearch) Close() error {
	// Elasticsearch client doesn't need explicit closing
	return nil
}
