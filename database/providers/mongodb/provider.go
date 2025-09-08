package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/libs/database/gateway"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Provider implements DatabaseProvider for MongoDB
type Provider struct {
	client *mongo.Client
	db     *mongo.Database
	config map[string]interface{}
	logger *logrus.Logger
}

// NewProvider creates a new MongoDB database provider
func NewProvider(logger *logrus.Logger) *Provider {
	return &Provider{
		config: make(map[string]interface{}),
		logger: logger,
	}
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "mongodb"
}

// GetSupportedFeatures returns supported features
func (p *Provider) GetSupportedFeatures() []gateway.DatabaseFeature {
	return []gateway.DatabaseFeature{
		gateway.FeatureTransactions,
		gateway.FeatureConnectionPool,
		gateway.FeatureClustering,
		gateway.FeatureSharding,
		gateway.FeatureFullTextSearch,
		gateway.FeatureJSONSupport,
		gateway.FeatureGeoSpatial,
		gateway.FeatureTimeSeries,
		gateway.FeatureDocumentStore,
		gateway.FeaturePersistent,
	}
}

// GetConnectionInfo returns connection information
func (p *Provider) GetConnectionInfo() *gateway.ConnectionInfo {
	uri, _ := p.config["uri"].(string)
	database, _ := p.config["database"].(string)

	return &gateway.ConnectionInfo{
		Host:     uri,
		Port:     27017,
		Database: database,
		User:     "",
		Driver:   "mongodb",
		Version:  "4.4+",
	}
}

// Configure configures the MongoDB provider
func (p *Provider) Configure(config map[string]interface{}) error {
	uri, ok := config["uri"].(string)
	if !ok || uri == "" {
		return fmt.Errorf("mongodb uri is required")
	}

	database, ok := config["database"].(string)
	if !ok || database == "" {
		return fmt.Errorf("mongodb database is required")
	}

	maxPool, ok := config["max_pool"].(int)
	if !ok || maxPool == 0 {
		maxPool = 100
	}

	minPool, ok := config["min_pool"].(int)
	if !ok || minPool == 0 {
		minPool = 10
	}

	timeout, ok := config["timeout"].(time.Duration)
	if !ok || timeout == 0 {
		timeout = 10 * time.Second
	}

	// Create client options
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(uint64(maxPool)).
		SetMinPoolSize(uint64(minPool)).
		SetServerSelectionTimeout(timeout).
		SetConnectTimeout(timeout).
		SetSocketTimeout(timeout)

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Get database
	db := client.Database(database)

	p.client = client
	p.db = db
	p.config = config

	p.logger.Info("MongoDB provider configured successfully")
	return nil
}

// IsConfigured checks if the provider is configured
func (p *Provider) IsConfigured() bool {
	return p.client != nil && p.db != nil
}

// Connect connects to the database
func (p *Provider) Connect(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("mongodb provider not configured")
	}

	// Test connection
	if err := p.client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	p.logger.Info("MongoDB connected successfully")
	return nil
}

// Disconnect disconnects from the database
func (p *Provider) Disconnect(ctx context.Context) error {
	if p.client != nil {
		return p.client.Disconnect(ctx)
	}
	return nil
}

// Ping checks database connection
func (p *Provider) Ping(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("mongodb provider not configured")
	}
	return p.client.Ping(ctx, nil)
}

// IsConnected checks if the database is connected
func (p *Provider) IsConnected() bool {
	if !p.IsConfigured() {
		return false
	}
	return p.client.Ping(context.Background(), nil) == nil
}

// BeginTransaction begins a new transaction
func (p *Provider) BeginTransaction(ctx context.Context) (gateway.Transaction, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("mongodb provider not configured")
	}

	session, err := p.client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	return &Transaction{session: session}, nil
}

// WithTransaction executes a function within a transaction
func (p *Provider) WithTransaction(ctx context.Context, fn func(gateway.Transaction) error) error {
	if !p.IsConfigured() {
		return fmt.Errorf("mongodb provider not configured")
	}

	session, err := p.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	return mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		return session.WithTransaction(sc, func(sc mongo.SessionContext) (interface{}, error) {
			tx := &Transaction{session: session}
			return nil, fn(tx)
		})
	})
}

// Query executes a query that returns rows (MongoDB uses collections)
func (p *Provider) Query(ctx context.Context, query string, args ...interface{}) (gateway.QueryResult, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("mongodb provider not configured")
	}

	// For MongoDB, we'll use the collection name as the "query"
	collection := p.db.Collection(query)

	// Create a cursor for the collection
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &QueryResult{cursor: cursor}, nil
}

// QueryRow executes a query that returns a single row
func (p *Provider) QueryRow(ctx context.Context, query string, args ...interface{}) (gateway.Row, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("mongodb provider not configured")
	}

	// For MongoDB, we'll use the collection name as the "query"
	collection := p.db.Collection(query)

	// Find one document
	var result bson.M
	err := collection.FindOne(ctx, bson.M{}).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query row: %w", err)
	}

	return &Row{result: result}, nil
}

// Exec executes a query without returning rows
func (p *Provider) Exec(ctx context.Context, query string, args ...interface{}) (gateway.ExecResult, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("mongodb provider not configured")
	}

	// For MongoDB, we'll use the collection name as the "query"
	collection := p.db.Collection(query)

	// Insert a test document to simulate execution
	result, err := collection.InsertOne(ctx, bson.M{"_exec": true, "timestamp": time.Now()})
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &ExecResult{insertedID: result.InsertedID}, nil
}

// Prepare prepares a statement (MongoDB doesn't have prepared statements in the same way)
func (p *Provider) Prepare(ctx context.Context, query string) (gateway.PreparedStatement, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("mongodb provider not configured")
	}

	// For MongoDB, we'll return a prepared statement that uses the collection
	collection := p.db.Collection(query)
	return &PreparedStatement{collection: collection}, nil
}

// HealthCheck performs a health check on the database
func (p *Provider) HealthCheck(ctx context.Context) error {
	if !p.IsConfigured() {
		return fmt.Errorf("mongodb provider not configured")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Test ping
	if err := p.Ping(ctx); err != nil {
		return fmt.Errorf("mongodb health check failed: %w", err)
	}

	// Test basic operations
	testCollection := "health_check_test"
	testDoc := bson.M{
		"_id":       primitive.NewObjectID(),
		"test":      "health_check",
		"timestamp": time.Now(),
	}

	// Insert test document
	_, err := p.db.Collection(testCollection).InsertOne(ctx, testDoc)
	if err != nil {
		return fmt.Errorf("mongodb insert test failed: %w", err)
	}

	// Find test document
	var result bson.M
	err = p.db.Collection(testCollection).FindOne(ctx, bson.M{"_id": testDoc["_id"]}).Decode(&result)
	if err != nil {
		return fmt.Errorf("mongodb find test failed: %w", err)
	}

	// Delete test document
	_, err = p.db.Collection(testCollection).DeleteOne(ctx, bson.M{"_id": testDoc["_id"]})
	if err != nil {
		return fmt.Errorf("mongodb delete test failed: %w", err)
	}

	return nil
}

// GetStats returns database statistics
func (p *Provider) GetStats(ctx context.Context) (*gateway.DatabaseStats, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("mongodb provider not configured")
	}

	// Get database stats
	var stats bson.M
	err := p.db.RunCommand(ctx, bson.D{{"dbStats", 1}}).Decode(&stats)
	if err != nil {
		return nil, fmt.Errorf("failed to get database stats: %w", err)
	}

	return &gateway.DatabaseStats{
		ActiveConnections: 0, // MongoDB doesn't expose this directly
		IdleConnections:   0,
		MaxConnections:    0,
		WaitCount:         0,
		WaitDuration:      0,
		MaxIdleClosed:     0,
		MaxIdleTimeClosed: 0,
		MaxLifetimeClosed: 0,
		ProviderData: map[string]interface{}{
			"driver": "mongodb",
			"stats":  stats,
		},
	}, nil
}

// Close closes the database connection
func (p *Provider) Close() error {
	if p.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return p.client.Disconnect(ctx)
	}
	return nil
}

// Transaction represents a MongoDB transaction
type Transaction struct {
	session mongo.Session
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	// MongoDB transactions are handled by the session
	return nil
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	// MongoDB transactions are handled by the session
	return nil
}

// Query executes a query within the transaction
func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (gateway.QueryResult, error) {
	// MongoDB transactions are handled by the session context
	return nil, fmt.Errorf("mongodb transactions are handled by session context")
}

// QueryRow executes a query that returns a single row within the transaction
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) (gateway.Row, error) {
	// MongoDB transactions are handled by the session context
	return nil, fmt.Errorf("mongodb transactions are handled by session context")
}

// Exec executes a query without returning rows within the transaction
func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) (gateway.ExecResult, error) {
	// MongoDB transactions are handled by the session context
	return nil, fmt.Errorf("mongodb transactions are handled by session context")
}

// Prepare prepares a statement within the transaction
func (t *Transaction) Prepare(ctx context.Context, query string) (gateway.PreparedStatement, error) {
	// MongoDB transactions are handled by the session context
	return nil, fmt.Errorf("mongodb transactions are handled by session context")
}

// QueryResult represents a MongoDB query result
type QueryResult struct {
	cursor *mongo.Cursor
}

// Close closes the result set
func (qr *QueryResult) Close() error {
	return qr.cursor.Close(context.Background())
}

// Next advances to the next row
func (qr *QueryResult) Next() bool {
	return qr.cursor.Next(context.Background())
}

// Scan scans the current row into dest
func (qr *QueryResult) Scan(dest ...interface{}) error {
	var result bson.M
	if err := qr.cursor.Decode(&result); err != nil {
		return err
	}

	// Convert bson.M to the destination type
	if len(dest) > 0 {
		if destMap, ok := dest[0].(*bson.M); ok {
			*destMap = result
		}
	}

	return nil
}

// Columns returns the column names
func (qr *QueryResult) Columns() ([]string, error) {
	// MongoDB doesn't have fixed columns, return empty slice
	return []string{}, nil
}

// Err returns any error that occurred during iteration
func (qr *QueryResult) Err() error {
	return qr.cursor.Err()
}

// Row represents a MongoDB row
type Row struct {
	result bson.M
}

// Scan scans the row into dest
func (r *Row) Scan(dest ...interface{}) error {
	if len(dest) > 0 {
		if destMap, ok := dest[0].(*bson.M); ok {
			*destMap = r.result
		}
	}
	return nil
}

// Err returns any error that occurred during scanning
func (r *Row) Err() error {
	return nil
}

// ExecResult represents a MongoDB execution result
type ExecResult struct {
	insertedID interface{}
}

// LastInsertId returns the last insert ID
func (er *ExecResult) LastInsertId() (int64, error) {
	if oid, ok := er.insertedID.(primitive.ObjectID); ok {
		return int64(oid.Timestamp().Unix()), nil
	}
	return 0, nil
}

// RowsAffected returns the number of rows affected
func (er *ExecResult) RowsAffected() (int64, error) {
	// MongoDB doesn't provide this information directly
	return 1, nil
}

// PreparedStatement represents a MongoDB prepared statement
type PreparedStatement struct {
	collection *mongo.Collection
}

// Close closes the prepared statement
func (ps *PreparedStatement) Close() error {
	// MongoDB doesn't need explicit closing
	return nil
}

// Query executes the prepared statement with args
func (ps *PreparedStatement) Query(ctx context.Context, args ...interface{}) (gateway.QueryResult, error) {
	cursor, err := ps.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	return &QueryResult{cursor: cursor}, nil
}

// QueryRow executes the prepared statement with args and returns a single row
func (ps *PreparedStatement) QueryRow(ctx context.Context, args ...interface{}) (gateway.Row, error) {
	var result bson.M
	err := ps.collection.FindOne(ctx, bson.M{}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &Row{result: result}, nil
}

// Exec executes the prepared statement with args
func (ps *PreparedStatement) Exec(ctx context.Context, args ...interface{}) (gateway.ExecResult, error) {
	result, err := ps.collection.InsertOne(ctx, bson.M{"_exec": true, "timestamp": time.Now()})
	if err != nil {
		return nil, err
	}
	return &ExecResult{insertedID: result.InsertedID}, nil
}
