package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anasamu/microservices-library-go/event/types"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// PostgreSQLProvider implements EventSourcingProvider for PostgreSQL
type PostgreSQLProvider struct {
	db     *sql.DB
	config *PostgreSQLConfig
	logger *logrus.Logger
}

// PostgreSQLConfig holds PostgreSQL-specific configuration
type PostgreSQLConfig struct {
	Host            string            `json:"host"`
	Port            int               `json:"port"`
	Database        string            `json:"database"`
	Username        string            `json:"username"`
	Password        string            `json:"password"`
	SSLMode         string            `json:"ssl_mode"`
	MaxConnections  int               `json:"max_connections"`
	MaxIdleConns    int               `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration     `json:"conn_max_lifetime"`
	Metadata        map[string]string `json:"metadata"`
}

// NewPostgreSQLProvider creates a new PostgreSQL event sourcing provider
func NewPostgreSQLProvider(config *PostgreSQLConfig, logger *logrus.Logger) (*PostgreSQLProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if logger == nil {
		logger = logrus.New()
	}

	provider := &PostgreSQLProvider{
		config: config,
		logger: logger,
	}

	return provider, nil
}

// GetName returns the provider name
func (p *PostgreSQLProvider) GetName() string {
	return "postgresql"
}

// GetSupportedFeatures returns the features supported by this provider
func (p *PostgreSQLProvider) GetSupportedFeatures() []types.EventSourcingFeature {
	return []types.EventSourcingFeature{
		types.FeatureEventAppend,
		types.FeatureEventRetrieval,
		types.FeatureStreamManagement,
		types.FeatureEventQuery,
		types.FeatureEventFiltering,
		types.FeatureSnapshots,
		types.FeatureEventReplay,
		types.FeatureEventVersioning,
		types.FeatureEventOrdering,
		types.FeatureEventCorrelation,
		types.FeatureEventStreaming,
	}
}

// GetConnectionInfo returns connection information
func (p *PostgreSQLProvider) GetConnectionInfo() *types.ConnectionInfo {
	return &types.ConnectionInfo{
		Host:     p.config.Host,
		Port:     p.config.Port,
		Database: p.config.Database,
		Username: p.config.Username,
		Status:   p.getConnectionStatus(),
		Metadata: p.config.Metadata,
	}
}

// Connect establishes connection to PostgreSQL
func (p *PostgreSQLProvider) Connect(ctx context.Context) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.config.Host, p.config.Port, p.config.Username, p.config.Password,
		p.config.Database, p.config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(p.config.MaxConnections)
	db.SetMaxIdleConns(p.config.MaxIdleConns)
	db.SetConnMaxLifetime(p.config.ConnMaxLifetime)

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	p.db = db

	// Create tables if they don't exist
	if err := p.createTables(ctx); err != nil {
		p.db.Close()
		return fmt.Errorf("failed to create tables: %w", err)
	}

	p.logger.Info("Connected to PostgreSQL event sourcing provider")
	return nil
}

// Disconnect closes the database connection
func (p *PostgreSQLProvider) Disconnect(ctx context.Context) error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// Ping checks if the database is reachable
func (p *PostgreSQLProvider) Ping(ctx context.Context) error {
	if p.db == nil {
		return fmt.Errorf("not connected to database")
	}
	return p.db.PingContext(ctx)
}

// IsConnected returns true if connected to the database
func (p *PostgreSQLProvider) IsConnected() bool {
	if p.db == nil {
		return false
	}
	return p.db.Ping() == nil
}

// AppendEvent appends an event to a stream
func (p *PostgreSQLProvider) AppendEvent(ctx context.Context, request *types.AppendEventRequest) (*types.AppendEventResponse, error) {
	if !p.IsConnected() {
		return nil, fmt.Errorf("not connected to database")
	}

	eventID := uuid.New().String()
	now := time.Now()

	// Serialize event data and metadata
	eventDataJSON, err := json.Marshal(request.EventData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event data: %w", err)
	}

	eventMetadataJSON, err := json.Marshal(request.EventMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event metadata: %w", err)
	}

	// Get current version for the stream
	var version int64
	err = p.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM events WHERE stream_id = $1", request.StreamID).Scan(&version)
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	// Check expected version if provided
	if request.ExpectedVersion != nil && *request.ExpectedVersion != version {
		return nil, &types.EventSourcingError{
			Code:     types.ErrCodeVersionMismatch,
			Message:  fmt.Sprintf("expected version %d, but current version is %d", *request.ExpectedVersion, version),
			StreamID: request.StreamID,
		}
	}

	newVersion := version + 1

	// Insert event
	query := `
		INSERT INTO events (
			id, stream_id, event_type, event_data, event_metadata, version, timestamp,
			correlation_id, causation_id, user_id, tenant_id, aggregate_id, aggregate_type
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err = p.db.ExecContext(ctx, query,
		eventID, request.StreamID, request.EventType, eventDataJSON, eventMetadataJSON,
		newVersion, now, request.CorrelationID, request.CausationID,
		request.UserID, request.TenantID, request.AggregateID, request.AggregateType)

	if err != nil {
		return nil, fmt.Errorf("failed to insert event: %w", err)
	}

	// Update stream version
	_, err = p.db.ExecContext(ctx, `
		INSERT INTO streams (id, name, aggregate_id, aggregate_type, version, created_at, updated_at)
		VALUES ($1, $1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			version = $4,
			updated_at = $6
	`, request.StreamID, request.AggregateID, request.AggregateType, newVersion, now, now)

	if err != nil {
		return nil, fmt.Errorf("failed to update stream: %w", err)
	}

	return &types.AppendEventResponse{
		EventID:   eventID,
		StreamID:  request.StreamID,
		Version:   newVersion,
		Timestamp: now,
	}, nil
}

// GetEvents retrieves events based on criteria
func (p *PostgreSQLProvider) GetEvents(ctx context.Context, request *types.GetEventsRequest) (*types.GetEventsResponse, error) {
	if !p.IsConnected() {
		return nil, fmt.Errorf("not connected to database")
	}

	query := "SELECT * FROM events WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	// Add filters
	if request.StreamID != "" {
		query += fmt.Sprintf(" AND stream_id = $%d", argIndex)
		args = append(args, request.StreamID)
		argIndex++
	}

	if len(request.EventTypes) > 0 {
		query += fmt.Sprintf(" AND event_type = ANY($%d)", argIndex)
		args = append(args, pq.Array(request.EventTypes))
		argIndex++
	}

	if request.StartVersion != nil {
		query += fmt.Sprintf(" AND version >= $%d", argIndex)
		args = append(args, *request.StartVersion)
		argIndex++
	}

	if request.EndVersion != nil {
		query += fmt.Sprintf(" AND version <= $%d", argIndex)
		args = append(args, *request.EndVersion)
		argIndex++
	}

	if request.StartTime != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, *request.StartTime)
		argIndex++
	}

	if request.EndTime != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, *request.EndTime)
		argIndex++
	}

	// Add ordering
	sortOrder := "ASC"
	if request.SortOrder == "desc" {
		sortOrder = "DESC"
	}
	query += fmt.Sprintf(" ORDER BY version %s", sortOrder)

	// Add limit and offset
	if request.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, request.Limit)
		argIndex++
	}

	if request.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, request.Offset)
		argIndex++
	}

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []*types.Event
	for rows.Next() {
		event, err := p.scanEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	// Get total count
	var totalCount int64
	countQuery := "SELECT COUNT(*) FROM events WHERE 1=1"
	countArgs := args[:len(args)-2] // Remove limit and offset
	if request.Limit > 0 || request.Offset > 0 {
		countArgs = args[:len(args)-2]
	}

	if len(countArgs) > 0 {
		err = p.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalCount)
	} else {
		err = p.db.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	hasMore := false
	nextOffset := 0
	if request.Limit > 0 && int64(request.Offset+request.Limit) < totalCount {
		hasMore = true
		nextOffset = request.Offset + request.Limit
	}

	return &types.GetEventsResponse{
		Events:     events,
		TotalCount: totalCount,
		HasMore:    hasMore,
		NextOffset: nextOffset,
	}, nil
}

// GetEventByID retrieves an event by its ID
func (p *PostgreSQLProvider) GetEventByID(ctx context.Context, request *types.GetEventByIDRequest) (*types.Event, error) {
	if !p.IsConnected() {
		return nil, fmt.Errorf("not connected to database")
	}

	query := "SELECT * FROM events WHERE id = $1"
	row := p.db.QueryRowContext(ctx, query, request.EventID)

	event, err := p.scanEvent(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &types.EventSourcingError{
				Code:    types.ErrCodeEventNotFound,
				Message: fmt.Sprintf("event with ID %s not found", request.EventID),
				EventID: request.EventID,
			}
		}
		return nil, fmt.Errorf("failed to get event by ID: %w", err)
	}

	return event, nil
}

// GetEventsByStream retrieves events for a specific stream
func (p *PostgreSQLProvider) GetEventsByStream(ctx context.Context, request *types.GetEventsByStreamRequest) (*types.GetEventsResponse, error) {
	getEventsRequest := &types.GetEventsRequest{
		StreamID:     request.StreamID,
		StartVersion: request.StartVersion,
		EndVersion:   request.EndVersion,
		Limit:        request.Limit,
		SortOrder:    "asc",
		Options:      request.Options,
	}

	return p.GetEvents(ctx, getEventsRequest)
}

// CreateStream creates a new event stream
func (p *PostgreSQLProvider) CreateStream(ctx context.Context, request *types.CreateStreamRequest) error {
	if !p.IsConnected() {
		return fmt.Errorf("not connected to database")
	}

	now := time.Now()
	metadataJSON, err := json.Marshal(request.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO streams (id, name, aggregate_id, aggregate_type, version, created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, 0, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING
	`

	_, err = p.db.ExecContext(ctx, query,
		request.StreamID, request.Name, request.AggregateID, request.AggregateType,
		now, now, metadataJSON)

	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	return nil
}

// DeleteStream deletes an event stream
func (p *PostgreSQLProvider) DeleteStream(ctx context.Context, request *types.DeleteStreamRequest) error {
	if !p.IsConnected() {
		return fmt.Errorf("not connected to database")
	}

	// Delete events first
	_, err := p.db.ExecContext(ctx, "DELETE FROM events WHERE stream_id = $1", request.StreamID)
	if err != nil {
		return fmt.Errorf("failed to delete events: %w", err)
	}

	// Delete stream
	_, err = p.db.ExecContext(ctx, "DELETE FROM streams WHERE id = $1", request.StreamID)
	if err != nil {
		return fmt.Errorf("failed to delete stream: %w", err)
	}

	return nil
}

// StreamExists checks if a stream exists
func (p *PostgreSQLProvider) StreamExists(ctx context.Context, request *types.StreamExistsRequest) (bool, error) {
	if !p.IsConnected() {
		return false, fmt.Errorf("not connected to database")
	}

	var exists bool
	err := p.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM streams WHERE id = $1)", request.StreamID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check stream existence: %w", err)
	}

	return exists, nil
}

// ListStreams lists all streams
func (p *PostgreSQLProvider) ListStreams(ctx context.Context) ([]types.StreamInfo, error) {
	if !p.IsConnected() {
		return nil, fmt.Errorf("not connected to database")
	}

	query := `
		SELECT s.id, s.name, s.aggregate_id, s.aggregate_type, s.version, s.created_at, s.updated_at, s.metadata,
		       COUNT(e.id) as event_count, COALESCE(MAX(e.timestamp), s.created_at) as last_event_at
		FROM streams s
		LEFT JOIN events e ON s.id = e.stream_id
		GROUP BY s.id, s.name, s.aggregate_id, s.aggregate_type, s.version, s.created_at, s.updated_at, s.metadata
		ORDER BY s.created_at DESC
	`

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list streams: %w", err)
	}
	defer rows.Close()

	var streams []types.StreamInfo
	for rows.Next() {
		var stream types.StreamInfo
		var metadataJSON []byte
		var lastEventAt sql.NullTime

		err := rows.Scan(
			&stream.ID, &stream.Name, &stream.AggregateID, &stream.AggregateType,
			&stream.Version, &stream.CreatedAt, &stream.UpdatedAt, &metadataJSON,
			&stream.EventCount, &lastEventAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stream: %w", err)
		}

		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &stream.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		if lastEventAt.Valid {
			stream.LastEventAt = &lastEventAt.Time
		}

		streams = append(streams, stream)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating streams: %w", err)
	}

	return streams, nil
}

// CreateSnapshot creates a snapshot of an aggregate state
func (p *PostgreSQLProvider) CreateSnapshot(ctx context.Context, request *types.CreateSnapshotRequest) (*types.CreateSnapshotResponse, error) {
	if !p.IsConnected() {
		return nil, fmt.Errorf("not connected to database")
	}

	snapshotID := uuid.New().String()
	now := time.Now()

	dataJSON, err := json.Marshal(request.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal snapshot data: %w", err)
	}

	metadataJSON, err := json.Marshal(request.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal snapshot metadata: %w", err)
	}

	query := `
		INSERT INTO snapshots (id, stream_id, aggregate_id, aggregate_type, version, data, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = p.db.ExecContext(ctx, query,
		snapshotID, request.StreamID, request.AggregateID, request.AggregateType,
		request.Version, dataJSON, metadataJSON, now)

	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	return &types.CreateSnapshotResponse{
		SnapshotID: snapshotID,
		StreamID:   request.StreamID,
		Version:    request.Version,
		Timestamp:  now,
	}, nil
}

// GetSnapshot retrieves a snapshot
func (p *PostgreSQLProvider) GetSnapshot(ctx context.Context, request *types.GetSnapshotRequest) (*types.Snapshot, error) {
	if !p.IsConnected() {
		return nil, fmt.Errorf("not connected to database")
	}

	query := "SELECT * FROM snapshots WHERE stream_id = $1"
	args := []interface{}{request.StreamID}
	argIndex := 2

	if request.Version != nil {
		query += fmt.Sprintf(" AND version <= $%d", argIndex)
		args = append(args, *request.Version)
		argIndex++
	}

	query += " ORDER BY version DESC LIMIT 1"

	row := p.db.QueryRowContext(ctx, query, args...)

	var snapshot types.Snapshot
	var dataJSON, metadataJSON []byte

	err := row.Scan(
		&snapshot.ID, &snapshot.StreamID, &snapshot.AggregateID, &snapshot.AggregateType,
		&snapshot.Version, &dataJSON, &metadataJSON, &snapshot.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &types.EventSourcingError{
				Code:     types.ErrCodeSnapshotNotFound,
				Message:  fmt.Sprintf("snapshot not found for stream %s", request.StreamID),
				StreamID: request.StreamID,
			}
		}
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	if err := json.Unmarshal(dataJSON, &snapshot.Data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot data: %w", err)
	}

	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &snapshot.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal snapshot metadata: %w", err)
		}
	}

	return &snapshot, nil
}

// DeleteSnapshot deletes a snapshot
func (p *PostgreSQLProvider) DeleteSnapshot(ctx context.Context, request *types.DeleteSnapshotRequest) error {
	if !p.IsConnected() {
		return fmt.Errorf("not connected to database")
	}

	_, err := p.db.ExecContext(ctx, "DELETE FROM snapshots WHERE id = $1", request.SnapshotID)
	if err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	return nil
}

// AppendEventsBatch appends multiple events in a batch
func (p *PostgreSQLProvider) AppendEventsBatch(ctx context.Context, request *types.AppendEventsBatchRequest) (*types.AppendEventsBatchResponse, error) {
	if !p.IsConnected() {
		return nil, fmt.Errorf("not connected to database")
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current version
	var version int64
	err = tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM events WHERE stream_id = $1", request.StreamID).Scan(&version)
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	// Check expected version if provided
	if request.ExpectedVersion != nil && *request.ExpectedVersion != version {
		return nil, &types.EventSourcingError{
			Code:     types.ErrCodeVersionMismatch,
			Message:  fmt.Sprintf("expected version %d, but current version is %d", *request.ExpectedVersion, version),
			StreamID: request.StreamID,
		}
	}

	startVersion := version
	appendedCount := 0
	var failedEvents []*types.Event

	for _, event := range request.Events {
		eventID := uuid.New().String()
		now := time.Now()
		version++

		// Serialize event data and metadata
		eventDataJSON, err := json.Marshal(event.EventData)
		if err != nil {
			failedEvents = append(failedEvents, event)
			continue
		}

		eventMetadataJSON, err := json.Marshal(event.EventMetadata)
		if err != nil {
			failedEvents = append(failedEvents, event)
			continue
		}

		// Insert event
		_, err = tx.ExecContext(ctx, `
			INSERT INTO events (
				id, stream_id, event_type, event_data, event_metadata, version, timestamp,
				correlation_id, causation_id, user_id, tenant_id, aggregate_id, aggregate_type
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		`, eventID, request.StreamID, event.EventType, eventDataJSON, eventMetadataJSON,
			version, now, event.CorrelationID, event.CausationID,
			event.UserID, event.TenantID, event.AggregateID, event.AggregateType)

		if err != nil {
			failedEvents = append(failedEvents, event)
			continue
		}

		appendedCount++
	}

	// Update stream version
	_, err = tx.ExecContext(ctx, `
		INSERT INTO streams (id, name, aggregate_id, aggregate_type, version, created_at, updated_at)
		VALUES ($1, $1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			version = $4,
			updated_at = $6
	`, request.StreamID, request.Events[0].AggregateID, request.Events[0].AggregateType, version, time.Now(), time.Now())

	if err != nil {
		return nil, fmt.Errorf("failed to update stream: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &types.AppendEventsBatchResponse{
		AppendedCount: appendedCount,
		FailedCount:   len(failedEvents),
		FailedEvents:  failedEvents,
		StartVersion:  startVersion + 1,
		EndVersion:    version,
	}, nil
}

// GetStreamInfo gets information about a stream
func (p *PostgreSQLProvider) GetStreamInfo(ctx context.Context, request *types.GetStreamInfoRequest) (*types.StreamInfo, error) {
	if !p.IsConnected() {
		return nil, fmt.Errorf("not connected to database")
	}

	query := `
		SELECT s.id, s.name, s.aggregate_id, s.aggregate_type, s.version, s.created_at, s.updated_at, s.metadata,
		       COUNT(e.id) as event_count, COALESCE(MAX(e.timestamp), s.created_at) as last_event_at
		FROM streams s
		LEFT JOIN events e ON s.id = e.stream_id
		WHERE s.id = $1
		GROUP BY s.id, s.name, s.aggregate_id, s.aggregate_type, s.version, s.created_at, s.updated_at, s.metadata
	`

	row := p.db.QueryRowContext(ctx, query, request.StreamID)

	var stream types.StreamInfo
	var metadataJSON []byte
	var lastEventAt sql.NullTime

	err := row.Scan(
		&stream.ID, &stream.Name, &stream.AggregateID, &stream.AggregateType,
		&stream.Version, &stream.CreatedAt, &stream.UpdatedAt, &metadataJSON,
		&stream.EventCount, &lastEventAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &types.EventSourcingError{
				Code:     types.ErrCodeStreamNotFound,
				Message:  fmt.Sprintf("stream %s not found", request.StreamID),
				StreamID: request.StreamID,
			}
		}
		return nil, fmt.Errorf("failed to get stream info: %w", err)
	}

	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &stream.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	if lastEventAt.Valid {
		stream.LastEventAt = &lastEventAt.Time
	}

	return &stream, nil
}

// GetStats returns event sourcing statistics
func (p *PostgreSQLProvider) GetStats(ctx context.Context) (*types.EventSourcingStats, error) {
	if !p.IsConnected() {
		return nil, fmt.Errorf("not connected to database")
	}

	stats := &types.EventSourcingStats{
		EventsByType:   make(map[string]int64),
		EventsByStream: make(map[string]int64),
		LastUpdate:     time.Now(),
		Provider:       p.GetName(),
	}

	// Get total events count
	err := p.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM events").Scan(&stats.TotalEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to get total events count: %w", err)
	}

	// Get total streams count
	err = p.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM streams").Scan(&stats.TotalStreams)
	if err != nil {
		return nil, fmt.Errorf("failed to get total streams count: %w", err)
	}

	// Get total snapshots count
	err = p.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM snapshots").Scan(&stats.TotalSnapshots)
	if err != nil {
		return nil, fmt.Errorf("failed to get total snapshots count: %w", err)
	}

	// Get events by type
	rows, err := p.db.QueryContext(ctx, "SELECT event_type, COUNT(*) FROM events GROUP BY event_type")
	if err != nil {
		return nil, fmt.Errorf("failed to get events by type: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var eventType string
		var count int64
		if err := rows.Scan(&eventType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan events by type: %w", err)
		}
		stats.EventsByType[eventType] = count
	}

	// Get events by stream
	rows, err = p.db.QueryContext(ctx, "SELECT stream_id, COUNT(*) FROM events GROUP BY stream_id")
	if err != nil {
		return nil, fmt.Errorf("failed to get events by stream: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var streamID string
		var count int64
		if err := rows.Scan(&streamID, &count); err != nil {
			return nil, fmt.Errorf("failed to scan events by stream: %w", err)
		}
		stats.EventsByStream[streamID] = count
	}

	return stats, nil
}

// HealthCheck performs a health check
func (p *PostgreSQLProvider) HealthCheck(ctx context.Context) error {
	if !p.IsConnected() {
		return fmt.Errorf("not connected to database")
	}

	return p.Ping(ctx)
}

// Configure configures the provider
func (p *PostgreSQLProvider) Configure(config map[string]interface{}) error {
	// This method can be used to update configuration at runtime
	// For now, we'll just log the configuration update
	p.logger.WithField("config", config).Info("PostgreSQL provider configuration updated")
	return nil
}

// IsConfigured returns true if the provider is configured
func (p *PostgreSQLProvider) IsConfigured() bool {
	return p.config != nil && p.config.Host != "" && p.config.Database != ""
}

// Close closes the provider
func (p *PostgreSQLProvider) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// Helper methods

func (p *PostgreSQLProvider) getConnectionStatus() types.ConnectionStatus {
	if p.IsConnected() {
		return types.StatusConnected
	}
	return types.StatusDisconnected
}

func (p *PostgreSQLProvider) createTables(ctx context.Context) error {
	// Create events table
	eventsTable := `
		CREATE TABLE IF NOT EXISTS events (
			id VARCHAR(36) PRIMARY KEY,
			stream_id VARCHAR(255) NOT NULL,
			event_type VARCHAR(255) NOT NULL,
			event_data JSONB NOT NULL,
			event_metadata JSONB,
			version BIGINT NOT NULL,
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
			correlation_id VARCHAR(255),
			causation_id VARCHAR(255),
			user_id VARCHAR(255),
			tenant_id VARCHAR(255),
			aggregate_id VARCHAR(255),
			aggregate_type VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`

	// Create streams table
	streamsTable := `
		CREATE TABLE IF NOT EXISTS streams (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			aggregate_id VARCHAR(255) NOT NULL,
			aggregate_type VARCHAR(255) NOT NULL,
			version BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			metadata JSONB
		)
	`

	// Create snapshots table
	snapshotsTable := `
		CREATE TABLE IF NOT EXISTS snapshots (
			id VARCHAR(36) PRIMARY KEY,
			stream_id VARCHAR(255) NOT NULL,
			aggregate_id VARCHAR(255) NOT NULL,
			aggregate_type VARCHAR(255) NOT NULL,
			version BIGINT NOT NULL,
			data JSONB NOT NULL,
			metadata JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_events_stream_id ON events(stream_id)",
		"CREATE INDEX IF NOT EXISTS idx_events_stream_version ON events(stream_id, version)",
		"CREATE INDEX IF NOT EXISTS idx_events_event_type ON events(event_type)",
		"CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_events_correlation_id ON events(correlation_id)",
		"CREATE INDEX IF NOT EXISTS idx_events_aggregate_id ON events(aggregate_id)",
		"CREATE INDEX IF NOT EXISTS idx_snapshots_stream_id ON snapshots(stream_id)",
		"CREATE INDEX IF NOT EXISTS idx_snapshots_stream_version ON snapshots(stream_id, version)",
	}

	// Execute table creation
	if _, err := p.db.ExecContext(ctx, eventsTable); err != nil {
		return fmt.Errorf("failed to create events table: %w", err)
	}

	if _, err := p.db.ExecContext(ctx, streamsTable); err != nil {
		return fmt.Errorf("failed to create streams table: %w", err)
	}

	if _, err := p.db.ExecContext(ctx, snapshotsTable); err != nil {
		return fmt.Errorf("failed to create snapshots table: %w", err)
	}

	// Execute index creation
	for _, index := range indexes {
		if _, err := p.db.ExecContext(ctx, index); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

func (p *PostgreSQLProvider) scanEvent(scanner interface{}) (*types.Event, error) {
	var event types.Event
	var eventDataJSON, eventMetadataJSON []byte
	var correlationID, causationID, userID, tenantID, aggregateID, aggregateType sql.NullString

	var err error
	switch s := scanner.(type) {
	case *sql.Row:
		err = s.Scan(
			&event.ID, &event.StreamID, &event.EventType, &eventDataJSON, &eventMetadataJSON,
			&event.Version, &event.Timestamp, &correlationID, &causationID,
			&userID, &tenantID, &aggregateID, &aggregateType,
		)
	case *sql.Rows:
		err = s.Scan(
			&event.ID, &event.StreamID, &event.EventType, &eventDataJSON, &eventMetadataJSON,
			&event.Version, &event.Timestamp, &correlationID, &causationID,
			&userID, &tenantID, &aggregateID, &aggregateType,
		)
	default:
		return nil, fmt.Errorf("unsupported scanner type")
	}

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(eventDataJSON, &event.EventData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	if eventMetadataJSON != nil {
		if err := json.Unmarshal(eventMetadataJSON, &event.EventMetadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event metadata: %w", err)
		}
	}

	// Set nullable fields
	if correlationID.Valid {
		event.CorrelationID = correlationID.String
	}
	if causationID.Valid {
		event.CausationID = causationID.String
	}
	if userID.Valid {
		event.UserID = userID.String
	}
	if tenantID.Valid {
		event.TenantID = tenantID.String
	}
	if aggregateID.Valid {
		event.AggregateID = aggregateID.String
	}
	if aggregateType.Valid {
		event.AggregateType = aggregateType.String
	}

	return &event, nil
}
