package database

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB represents MongoDB database connection
type MongoDB struct {
	Client *mongo.Client
	DB     *mongo.Database
	config *MongoDBConfig
	logger *logrus.Logger
}

// MongoDBConfig holds MongoDB configuration
type MongoDBConfig struct {
	URI      string
	Database string
	MaxPool  int
	MinPool  int
	Timeout  time.Duration
}

// NewMongoDB creates new MongoDB connection
func NewMongoDB(config *MongoDBConfig, logger *logrus.Logger) (*MongoDB, error) {
	// Set default timeout if not provided
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	// Set default pool sizes if not provided
	if config.MaxPool == 0 {
		config.MaxPool = 100
	}
	if config.MinPool == 0 {
		config.MinPool = 10
	}

	// Create client options
	clientOptions := options.Client().
		ApplyURI(config.URI).
		SetMaxPoolSize(uint64(config.MaxPool)).
		SetMinPoolSize(uint64(config.MinPool)).
		SetServerSelectionTimeout(config.Timeout).
		SetConnectTimeout(config.Timeout).
		SetSocketTimeout(config.Timeout)

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Get database
	db := client.Database(config.Database)

	mongoDB := &MongoDB{
		Client: client,
		DB:     db,
		config: config,
		logger: logger,
	}

	mongoDB.logger.Info("MongoDB connection established successfully")
	return mongoDB, nil
}

// Close closes the MongoDB connection
func (m *MongoDB) Close() error {
	if m.Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		m.logger.Info("Closing MongoDB connection")
		return m.Client.Disconnect(ctx)
	}
	return nil
}

// Ping checks MongoDB connection
func (m *MongoDB) Ping(ctx context.Context) error {
	return m.Client.Ping(ctx, nil)
}

// GetDatabase returns the database instance
func (m *MongoDB) GetDatabase() *mongo.Database {
	return m.DB
}

// GetCollection returns a collection
func (m *MongoDB) GetCollection(name string) *mongo.Collection {
	return m.DB.Collection(name)
}

// InsertOne inserts a single document
func (m *MongoDB) InsertOne(ctx context.Context, collection string, document interface{}) (*mongo.InsertOneResult, error) {
	coll := m.GetCollection(collection)
	result, err := coll.InsertOne(ctx, document)
	if err != nil {
		return nil, fmt.Errorf("failed to insert document: %w", err)
	}
	return result, nil
}

// InsertMany inserts multiple documents
func (m *MongoDB) InsertMany(ctx context.Context, collection string, documents []interface{}) (*mongo.InsertManyResult, error) {
	coll := m.GetCollection(collection)
	result, err := coll.InsertMany(ctx, documents)
	if err != nil {
		return nil, fmt.Errorf("failed to insert documents: %w", err)
	}
	return result, nil
}

// FindOne finds a single document
func (m *MongoDB) FindOne(ctx context.Context, collection string, filter interface{}, result interface{}) error {
	coll := m.GetCollection(collection)
	err := coll.FindOne(ctx, filter).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("document not found")
		}
		return fmt.Errorf("failed to find document: %w", err)
	}
	return nil
}

// Find finds multiple documents
func (m *MongoDB) Find(ctx context.Context, collection string, filter interface{}, results interface{}, opts ...*options.FindOptions) error {
	coll := m.GetCollection(collection)
	cursor, err := coll.Find(ctx, filter, opts...)
	if err != nil {
		return fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, results); err != nil {
		return fmt.Errorf("failed to decode documents: %w", err)
	}
	return nil
}

// FindOneAndUpdate finds and updates a single document
func (m *MongoDB) FindOneAndUpdate(ctx context.Context, collection string, filter, update interface{}, result interface{}, opts ...*options.FindOneAndUpdateOptions) error {
	coll := m.GetCollection(collection)
	err := coll.FindOneAndUpdate(ctx, filter, update, opts...).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("document not found")
		}
		return fmt.Errorf("failed to find and update document: %w", err)
	}
	return nil
}

// UpdateOne updates a single document
func (m *MongoDB) UpdateOne(ctx context.Context, collection string, filter, update interface{}) (*mongo.UpdateResult, error) {
	coll := m.GetCollection(collection)
	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}
	return result, nil
}

// UpdateMany updates multiple documents
func (m *MongoDB) UpdateMany(ctx context.Context, collection string, filter, update interface{}) (*mongo.UpdateResult, error) {
	coll := m.GetCollection(collection)
	result, err := coll.UpdateMany(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update documents: %w", err)
	}
	return result, nil
}

// DeleteOne deletes a single document
func (m *MongoDB) DeleteOne(ctx context.Context, collection string, filter interface{}) (*mongo.DeleteResult, error) {
	coll := m.GetCollection(collection)
	result, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to delete document: %w", err)
	}
	return result, nil
}

// DeleteMany deletes multiple documents
func (m *MongoDB) DeleteMany(ctx context.Context, collection string, filter interface{}) (*mongo.DeleteResult, error) {
	coll := m.GetCollection(collection)
	result, err := coll.DeleteMany(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to delete documents: %w", err)
	}
	return result, nil
}

// CountDocuments counts documents matching the filter
func (m *MongoDB) CountDocuments(ctx context.Context, collection string, filter interface{}) (int64, error) {
	coll := m.GetCollection(collection)
	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}
	return count, nil
}

// Aggregate performs aggregation pipeline
func (m *MongoDB) Aggregate(ctx context.Context, collection string, pipeline interface{}, results interface{}) error {
	coll := m.GetCollection(collection)
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("failed to aggregate documents: %w", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, results); err != nil {
		return fmt.Errorf("failed to decode aggregation results: %w", err)
	}
	return nil
}

// CreateIndex creates an index on a collection
func (m *MongoDB) CreateIndex(ctx context.Context, collection string, model mongo.IndexModel) (string, error) {
	coll := m.GetCollection(collection)
	indexName, err := coll.Indexes().CreateOne(ctx, model)
	if err != nil {
		return "", fmt.Errorf("failed to create index: %w", err)
	}
	return indexName, nil
}

// CreateIndexes creates multiple indexes on a collection
func (m *MongoDB) CreateIndexes(ctx context.Context, collection string, models []mongo.IndexModel) ([]string, error) {
	coll := m.GetCollection(collection)
	indexNames, err := coll.Indexes().CreateMany(ctx, models)
	if err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}
	return indexNames, nil
}

// ListIndexes lists all indexes on a collection
func (m *MongoDB) ListIndexes(ctx context.Context, collection string) ([]bson.M, error) {
	coll := m.GetCollection(collection)
	cursor, err := coll.Indexes().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list indexes: %w", err)
	}
	defer cursor.Close(ctx)

	var indexes []bson.M
	if err := cursor.All(ctx, &indexes); err != nil {
		return nil, fmt.Errorf("failed to decode indexes: %w", err)
	}
	return indexes, nil
}

// DropIndex drops an index
func (m *MongoDB) DropIndex(ctx context.Context, collection, indexName string) error {
	coll := m.GetCollection(collection)
	_, err := coll.Indexes().DropOne(ctx, indexName)
	if err != nil {
		return fmt.Errorf("failed to drop index: %w", err)
	}
	return nil
}

// DropIndexes drops all indexes except the default _id index
func (m *MongoDB) DropIndexes(ctx context.Context, collection string) error {
	coll := m.GetCollection(collection)
	_, err := coll.Indexes().DropAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to drop indexes: %w", err)
	}
	return nil
}

// CreateCollection creates a new collection
func (m *MongoDB) CreateCollection(ctx context.Context, name string, opts ...*options.CreateCollectionOptions) error {
	err := m.DB.CreateCollection(ctx, name, opts...)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}
	return nil
}

// DropCollection drops a collection
func (m *MongoDB) DropCollection(ctx context.Context, name string) error {
	err := m.DB.Collection(name).Drop(ctx)
	if err != nil {
		return fmt.Errorf("failed to drop collection: %w", err)
	}
	return nil
}

// ListCollections lists all collections
func (m *MongoDB) ListCollections(ctx context.Context) ([]string, error) {
	cursor, err := m.DB.ListCollections(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	defer cursor.Close(ctx)

	var collections []string
	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode collection: %w", err)
		}
		if name, ok := result["name"].(string); ok {
			collections = append(collections, name)
		}
	}

	return collections, nil
}

// StartSession starts a new session
func (m *MongoDB) StartSession() (mongo.Session, error) {
	session, err := m.Client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}
	return session, nil
}

// WithTransaction executes a function within a transaction
func (m *MongoDB) WithTransaction(ctx context.Context, fn func(mongo.SessionContext) (interface{}, error)) (interface{}, error) {
	session, err := m.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	return mongo.WithSession(ctx, session, func(sc mongo.SessionContext) (interface{}, error) {
		return session.WithTransaction(sc, func(sc mongo.SessionContext) (interface{}, error) {
			return fn(sc)
		})
	})
}

// BulkWrite performs bulk write operations
func (m *MongoDB) BulkWrite(ctx context.Context, collection string, operations []mongo.WriteModel) (*mongo.BulkWriteResult, error) {
	coll := m.GetCollection(collection)
	result, err := coll.BulkWrite(ctx, operations)
	if err != nil {
		return nil, fmt.Errorf("failed to perform bulk write: %w", err)
	}
	return result, nil
}

// Watch watches for changes on a collection
func (m *MongoDB) Watch(ctx context.Context, collection string, pipeline interface{}) (*mongo.ChangeStream, error) {
	coll := m.GetCollection(collection)
	stream, err := coll.Watch(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to watch collection: %w", err)
	}
	return stream, nil
}

// HealthCheck performs a health check on MongoDB
func (m *MongoDB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Test ping
	if err := m.Ping(ctx); err != nil {
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
	_, err := m.InsertOne(ctx, testCollection, testDoc)
	if err != nil {
		return fmt.Errorf("mongodb insert test failed: %w", err)
	}

	// Find test document
	var result bson.M
	err = m.FindOne(ctx, testCollection, bson.M{"_id": testDoc["_id"]}, &result)
	if err != nil {
		return fmt.Errorf("mongodb find test failed: %w", err)
	}

	// Delete test document
	_, err = m.DeleteOne(ctx, testCollection, bson.M{"_id": testDoc["_id"]})
	if err != nil {
		return fmt.Errorf("mongodb delete test failed: %w", err)
	}

	return nil
}

// GetStats returns database statistics
func (m *MongoDB) GetStats(ctx context.Context) (bson.M, error) {
	var stats bson.M
	err := m.DB.RunCommand(ctx, bson.D{{"dbStats", 1}}).Decode(&stats)
	if err != nil {
		return nil, fmt.Errorf("failed to get database stats: %w", err)
	}
	return stats, nil
}

// GetCollectionStats returns collection statistics
func (m *MongoDB) GetCollectionStats(ctx context.Context, collection string) (bson.M, error) {
	coll := m.GetCollection(collection)
	var stats bson.M
	err := coll.Database().RunCommand(ctx, bson.D{
		{"collStats", collection},
	}).Decode(&stats)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection stats: %w", err)
	}
	return stats, nil
}

// Helper functions for common operations

// FindByID finds a document by ObjectID
func (m *MongoDB) FindByID(ctx context.Context, collection string, id primitive.ObjectID, result interface{}) error {
	return m.FindOne(ctx, collection, bson.M{"_id": id}, result)
}

// FindByStringID finds a document by string ID
func (m *MongoDB) FindByStringID(ctx context.Context, collection string, id string, result interface{}) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid object ID: %w", err)
	}
	return m.FindByID(ctx, collection, objectID, result)
}

// UpdateByID updates a document by ObjectID
func (m *MongoDB) UpdateByID(ctx context.Context, collection string, id primitive.ObjectID, update interface{}) (*mongo.UpdateResult, error) {
	return m.UpdateOne(ctx, collection, bson.M{"_id": id}, update)
}

// UpdateByStringID updates a document by string ID
func (m *MongoDB) UpdateByStringID(ctx context.Context, collection string, id string, update interface{}) (*mongo.UpdateResult, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid object ID: %w", err)
	}
	return m.UpdateByID(ctx, collection, objectID, update)
}

// DeleteByID deletes a document by ObjectID
func (m *MongoDB) DeleteByID(ctx context.Context, collection string, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	return m.DeleteOne(ctx, collection, bson.M{"_id": id})
}

// DeleteByStringID deletes a document by string ID
func (m *MongoDB) DeleteByStringID(ctx context.Context, collection string, id string) (*mongo.DeleteResult, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid object ID: %w", err)
	}
	return m.DeleteByID(ctx, collection, objectID)
}

// Exists checks if a document exists
func (m *MongoDB) Exists(ctx context.Context, collection string, filter interface{}) (bool, error) {
	count, err := m.CountDocuments(ctx, collection, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByID checks if a document exists by ObjectID
func (m *MongoDB) ExistsByID(ctx context.Context, collection string, id primitive.ObjectID) (bool, error) {
	return m.Exists(ctx, collection, bson.M{"_id": id})
}

// ExistsByStringID checks if a document exists by string ID
func (m *MongoDB) ExistsByStringID(ctx context.Context, collection string, id string) (bool, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, fmt.Errorf("invalid object ID: %w", err)
	}
	return m.ExistsByID(ctx, collection, objectID)
}
