package communication

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// GRPCServer represents a gRPC server
type GRPCServer struct {
	server *grpc.Server
	config *GRPCConfig
	logger *logrus.Logger
}

// GRPCClient represents a gRPC client
type GRPCClient struct {
	conn   *grpc.ClientConn
	config *GRPCConfig
	logger *logrus.Logger
}

// GRPCConfig holds gRPC configuration
type GRPCConfig struct {
	Port                         string
	Host                         string
	Timeout                      int
	MaxRecvMsgSize               int
	MaxSendMsgSize               int
	KeepAlive                    bool
	KeepAliveTime                time.Duration
	KeepAliveTimeout             time.Duration
	KeepAlivePermitWithoutStream bool
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(config *GRPCConfig, logger *logrus.Logger) *GRPCServer {
	// Configure keepalive parameters
	kaep := keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second,
		PermitWithoutStream: config.KeepAlivePermitWithoutStream,
	}

	kasp := keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Second,
		MaxConnectionAge:      30 * time.Second,
		MaxConnectionAgeGrace: 5 * time.Second,
		Time:                  config.KeepAliveTime,
		Timeout:               config.KeepAliveTimeout,
	}

	// Create server with middleware
	server := grpc.NewServer(
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
		grpc.MaxRecvMsgSize(config.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(config.MaxSendMsgSize),
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(),
			logging.UnaryServerInterceptor(loggingInterceptor(logger)),
			// Add custom interceptors here
		),
		grpc.ChainStreamInterceptor(
			recovery.StreamServerInterceptor(),
			logging.StreamServerInterceptor(loggingInterceptor(logger)),
			// Add custom interceptors here
		),
	)

	// Enable reflection for development
	if config.Host == "localhost" || config.Host == "127.0.0.1" {
		reflection.Register(server)
	}

	return &GRPCServer{
		server: server,
		config: config,
		logger: logger,
	}
}

// NewGRPCClient creates a new gRPC client
func NewGRPCClient(config *GRPCConfig, logger *logrus.Logger) (*GRPCClient, error) {
	// Configure keepalive parameters
	kacp := keepalive.ClientParameters{
		Time:                config.KeepAliveTime,
		Timeout:             config.KeepAliveTimeout,
		PermitWithoutStream: config.KeepAlivePermitWithoutStream,
	}

	// Create connection with middleware
	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%s", config.Host, config.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithChainUnaryInterceptor(
			logging.UnaryClientInterceptor(loggingInterceptor(logger)),
			retry.UnaryClientInterceptor(),
			// Add custom interceptors here
		),
		grpc.WithChainStreamInterceptor(
			logging.StreamClientInterceptor(loggingInterceptor(logger)),
			retry.StreamClientInterceptor(),
			// Add custom interceptors here
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	return &GRPCClient{
		conn:   conn,
		config: config,
		logger: logger,
	}, nil
}

// Start starts the gRPC server
func (s *GRPCServer) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", s.config.Host, s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.logger.Infof("gRPC server starting on %s:%s", s.config.Host, s.config.Port)

	go func() {
		if err := s.server.Serve(lis); err != nil {
			s.logger.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	return nil
}

// Stop stops the gRPC server
func (s *GRPCServer) Stop() {
	s.logger.Info("Stopping gRPC server...")
	s.server.GracefulStop()
}

// GetServer returns the underlying gRPC server
func (s *GRPCServer) GetServer() *grpc.Server {
	return s.server
}

// RegisterService registers a service with the server
func (s *GRPCServer) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.server.RegisterService(desc, impl)
}

// Close closes the gRPC client connection
func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetConnection returns the underlying gRPC connection
func (c *GRPCClient) GetConnection() *grpc.ClientConn {
	return c.conn
}

// HealthCheck performs a health check on the gRPC connection
func (c *GRPCClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(c.config.Timeout)*time.Second)
	defer cancel()

	// Try to get connection state
	state := c.conn.GetState()
	if state.String() == "SHUTDOWN" || state.String() == "TRANSIENT_FAILURE" {
		return fmt.Errorf("gRPC connection is in %s state", state.String())
	}

	return nil
}

// loggingInterceptor creates a logging interceptor for gRPC
func loggingInterceptor(logger *logrus.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		entry := logger.WithFields(logrus.Fields{})

		// Parse fields
		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				key := fmt.Sprintf("%v", fields[i])
				value := fields[i+1]
				entry = entry.WithField(key, value)
			}
		}

		// Log based on level
		switch lvl {
		case logging.LevelDebug:
			entry.Debug(msg)
		case logging.LevelInfo:
			entry.Info(msg)
		case logging.LevelWarn:
			entry.Warn(msg)
		case logging.LevelError:
			entry.Error(msg)
		default:
			entry.Info(msg)
		}
	})
}

// UnaryServerInterceptor creates a custom unary server interceptor
func UnaryServerInterceptor(logger *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Log request
		logger.WithFields(logrus.Fields{
			"method":  info.FullMethod,
			"request": req,
		}).Debug("gRPC request received")

		// Call handler
		resp, err := handler(ctx, req)

		// Log response
		duration := time.Since(start)
		logger.WithFields(logrus.Fields{
			"method":   info.FullMethod,
			"duration": duration,
			"error":    err,
		}).Debug("gRPC request completed")

		return resp, err
	}
}

// StreamServerInterceptor creates a custom stream server interceptor
func StreamServerInterceptor(logger *logrus.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		// Log stream start
		logger.WithFields(logrus.Fields{
			"method": info.FullMethod,
		}).Debug("gRPC stream started")

		// Call handler
		err := handler(srv, ss)

		// Log stream end
		duration := time.Since(start)
		logger.WithFields(logrus.Fields{
			"method":   info.FullMethod,
			"duration": duration,
			"error":    err,
		}).Debug("gRPC stream completed")

		return err
	}
}

// UnaryClientInterceptor creates a custom unary client interceptor
func UnaryClientInterceptor(logger *logrus.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()

		// Log request
		logger.WithFields(logrus.Fields{
			"method":  method,
			"request": req,
		}).Debug("gRPC client request sent")

		// Call invoker
		err := invoker(ctx, method, req, reply, cc, opts...)

		// Log response
		duration := time.Since(start)
		logger.WithFields(logrus.Fields{
			"method":   method,
			"duration": duration,
			"error":    err,
		}).Debug("gRPC client request completed")

		return err
	}
}

// StreamClientInterceptor creates a custom stream client interceptor
func StreamClientInterceptor(logger *logrus.Logger) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		start := time.Now()

		// Log stream start
		logger.WithFields(logrus.Fields{
			"method": method,
		}).Debug("gRPC client stream started")

		// Create stream
		stream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"method": method,
				"error":  err,
			}).Error("gRPC client stream failed")
			return nil, err
		}

		// Wrap stream with logging
		wrappedStream := &loggingClientStream{
			ClientStream: stream,
			method:       method,
			start:        start,
			logger:       logger,
		}

		return wrappedStream, nil
	}
}

// loggingClientStream wraps a gRPC client stream with logging
type loggingClientStream struct {
	grpc.ClientStream
	method string
	start  time.Time
	logger *logrus.Logger
}

// SendMsg logs and sends a message
func (s *loggingClientStream) SendMsg(m interface{}) error {
	err := s.ClientStream.SendMsg(m)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"method": s.method,
			"error":  err,
		}).Error("gRPC client stream send failed")
	}
	return err
}

// RecvMsg logs and receives a message
func (s *loggingClientStream) RecvMsg(m interface{}) error {
	err := s.ClientStream.RecvMsg(m)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"method": s.method,
			"error":  err,
		}).Error("gRPC client stream receive failed")
	}
	return err
}

// CloseSend closes the send direction
func (s *loggingClientStream) CloseSend() error {
	err := s.ClientStream.CloseSend()
	duration := time.Since(s.start)

	s.logger.WithFields(logrus.Fields{
		"method":   s.method,
		"duration": duration,
		"error":    err,
	}).Debug("gRPC client stream closed")

	return err
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts int
	Backoff     time.Duration
	MaxBackoff  time.Duration
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		Backoff:     100 * time.Millisecond,
		MaxBackoff:  1 * time.Second,
	}
}

// RetryInterceptor creates a retry interceptor
func RetryInterceptor(config RetryConfig) grpc.UnaryClientInterceptor {
	return retry.UnaryClientInterceptor(
		retry.WithCodes(codes.Unavailable, codes.DeadlineExceeded),
		retry.WithMax(uint(config.MaxAttempts)),
		retry.WithBackoff(retry.BackoffLinear(config.Backoff)),
	)
}
