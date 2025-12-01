package bus

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// ResponseStreamSuffix is appended to stream name for response delivery
	ResponseStreamSuffix = ":responses"
	// DefaultTimeout in seconds for requests
	DefaultTimeout = 300
)

// Subscriber defines the interface for message consumers
type Subscriber interface {
	Handle(ctx context.Context) (any, error)
}

// Publisher defines the interface for message producers
type Publisher interface {
	String() string
	Serialize() ([]byte, error)
}

// Response represents the response from a subscriber
type Response struct {
	Data  any
	Error error
}

// RedisClient defines the interface for Redis operations used by Bus
type RedisClient interface {
	XAdd(ctx context.Context, args *redis.XAddArgs) *redis.StringCmd
	XRead(ctx context.Context, args *redis.XReadArgs) *redis.XStreamSliceCmd
	Pipeline() redis.Pipeliner
}

// Bus is the main message bus implementation using Redis streams
type Bus struct {
	redis       RedisClient
	serializer  BrokerSerialize
	subscribers map[string]func(data []byte) (Subscriber, error)
	mu          sync.RWMutex
	responses   map[string]chan Response
	responseMu  sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewBus creates a new Bus instance with the provided Redis client
// Uses RedisBrokerSerialize as default serializer
func NewBus(redis RedisClient, ctx context.Context) *Bus {
	return &Bus{
		redis:       redis,
		serializer:  NewRedisBrokerSerialize(),
		subscribers: make(map[string]func(data []byte) (Subscriber, error)),
		responses:   make(map[string]chan Response),
		ctx:         ctx,
	}
}

// SetSerializer sets a custom serializer for the Bus
func (b *Bus) SetSerializer(serializer BrokerSerialize) {
	b.serializer = serializer
}

// Register registers a constructor of a subscriber for a specific stream
func (b *Bus) Register(streamName string, constructor func(data []byte) (Subscriber, error)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[streamName] = constructor
	log.Printf("Registered constructor for stream: %s", streamName)
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Execute sends a message and waits for a response
func (b *Bus) Execute(ctx context.Context, pub Publisher) (Response, error) {
	streamName := pub.String()
	requestID := generateRequestID()

	// Serialize the publisher payload
	payload, err := pub.Serialize()
	if err != nil {
		return Response{}, fmt.Errorf("failed to serialize publisher: %w", err)
	}

	// Create TransportRequest matching Python's format
	transportReq := TransportRequest{
		CreatedTimestamp: float64(time.Now().UnixNano()) / 1e9,
		RequestID:        requestID,
		Message:          []byte{0xa0}, // Empty CBOR map
		Properties:       payload,
		ReturnResult:     1, // Request response
		Timeout:          DefaultTimeout,
	}

	// Serialize using broker serializer
	values, err := b.serializer.Serialize(&transportReq)
	if err != nil {
		return Response{}, fmt.Errorf("failed to serialize transport request: %w", err)
	}

	// Create response channel
	responseCh := make(chan Response, 1)
	b.responseMu.Lock()
	b.responses[requestID] = responseCh
	b.responseMu.Unlock()

	// Clean up response channel after timeout or completion
	defer func() {
		b.responseMu.Lock()
		delete(b.responses, requestID)
		close(responseCh)
		b.responseMu.Unlock()
	}()

	// Add message to stream
	msgID, err := b.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: values,
	}).Result()
	if err != nil {
		return Response{}, fmt.Errorf("failed to add message to stream: %w", err)
	}

	log.Printf("Sent message to stream %s with ID %s, request_id: %s", streamName, msgID, requestID)

	// Wait for response with timeout
	timeout := time.Duration(DefaultTimeout) * time.Second
	select {
	case response := <-responseCh:
		return response, nil
	case <-ctx.Done():
		return Response{}, fmt.Errorf("context cancelled: %w", ctx.Err())
	case <-time.After(timeout):
		return Response{}, fmt.Errorf("timeout waiting for response (request_id: %s)", requestID)
	}
}

// Emit sends a message without waiting for a response
func (b *Bus) Emit(ctx context.Context, pub Publisher) error {
	streamName := pub.String()
	requestID := generateRequestID()

	// Serialize the publisher payload
	payload, err := pub.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize publisher: %w", err)
	}

	// Create TransportRequest matching Python's format
	transportReq := TransportRequest{
		CreatedTimestamp: float64(time.Now().UnixNano()) / 1e9,
		RequestID:        requestID,
		Message:          []byte{0xa0}, // Empty CBOR map
		Properties:       payload,
		ReturnResult:     0, // No response needed
		Timeout:          DefaultTimeout,
	}

	// Serialize using broker serializer
	values, err := b.serializer.Serialize(&transportReq)
	if err != nil {
		return fmt.Errorf("failed to serialize transport request: %w", err)
	}

	// Add message to stream
	msgID, err := b.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: values,
	}).Result()
	if err != nil {
		return fmt.Errorf("failed to add message to stream: %w", err)
	}

	log.Printf("Emitted message to stream %s with ID %s, request_id: %s", streamName, msgID, requestID)
	return nil
}

// Run starts listening to all registered streams and processing messages
func (b *Bus) Run() {
	b.mu.RLock()
	if len(b.subscribers) == 0 {
		b.mu.RUnlock()
		log.Println("No subscribers registered, nothing to run")
		return
	}

	// Build streams list
	streams := make([]string, 0, len(b.subscribers))
	for streamName := range b.subscribers {
		streams = append(streams, streamName)
	}
	b.mu.RUnlock()

	log.Printf("Starting bus listener for %d streams", len(streams))

	for _, stream := range streams {
		go func() {
			b.processStream(stream)
		}()
	}
}

// processStream reads messages from Redis stream and processes them
func (b *Bus) processStream(streamName string) {
	lastID := "$" // читать только новые сообщения
	// lastID := "0" // читать все сообщения
	constructor := b.subscribers[streamName]

	for {
		res, err := b.redis.XRead(b.ctx, &redis.XReadArgs{
			Streams: []string{streamName, lastID},
			Count:   1,
			Block:   0, // ждём пока появятся
		}).Result()

		if err != nil {
			log.Printf("XRead error: %v", err)
			continue
		}

		// XRead может вернуть массив stream-результатов (обычно 1)
		for _, stream := range res {
			for _, msg := range stream.Messages {
				// Deserialize TransportRequest from message using broker serializer
				transportReq, err := b.deserializeMessage(msg)
				if err != nil {
					log.Printf("failed to deserialize TransportRequest for stream %s, message ID %s: %v", streamName, msg.ID, err)
					continue
				}
				// Create subscriber using properties from TransportRequest
				subscriber, err := constructor(transportReq.Properties)
				if err != nil {
					log.Printf("failed to create subscriber for stream %s: %v", streamName, err)
					continue
				}

				b.processMessage(streamName, subscriber, transportReq)
			}

		}
	}
}

// processMessage processes a single message from a stream
func (b *Bus) processMessage(streamName string, subscriber Subscriber, req *TransportRequest) {
	result, err := subscriber.Handle(context.Background())

	if !req.NeedsResponse() {
		return
	}

	if result == nil && err == nil {
		return
	}

	response := Response{
		Data:  result,
		Error: err,
	}
	b.sendResponse(streamName, req.RequestID, req.RedisMessageID, response)
}

// sendResponse sends a response back via Redis and notifies local waiting calls
func (b *Bus) sendResponse(streamName string, requestID string, redisMessageID string, response Response) {
	const fn = "sendResponse"

	// Create TransportResponse with result data (will be CBOR-encoded by Encode())
	transportResp := TransportResponse{
		ReqID:  requestID,
		Result: response.Data,
	}

	if response.Error != nil {
		transportResp.Error = response.Error.Error()
		transportResp.ErrorClass = fmt.Sprintf("%T", response.Error)
	}

	// Encode TransportResponse to CBOR
	responseBytes, err := transportResp.Encode()
	if err != nil {
		log.Printf("%s: failed to encode TransportResponse for request_id %s: %v", fn, requestID, err)
		return
	}

	// Ответ в Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipe := b.redis.Pipeline()
	pipe.RPush(ctx, requestID, responseBytes)
	pipe.Expire(ctx, requestID, 30*time.Second)
	// Удаляем сообщение из потока
	pipe.XDel(ctx, streamName, redisMessageID)

	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("%s: Failed to write response or delete message: %v", fn, err)
	}

	// Also notify local waiting channel if exists
	b.responseMu.RLock()
	responseCh, exists := b.responses[requestID]
	b.responseMu.RUnlock()

	if exists {
		select {
		case responseCh <- response:
			log.Printf("Sent response for request_id: %s", requestID)
		case <-time.After(5 * time.Second):
			log.Printf("Timeout sending response for request_id: %s", requestID)
		}
	}
}

// deserializeMessage deserializes a Redis message to TransportRequest
// Supports both formats: CBOR-encoded "data" field and individual fields
func (b *Bus) deserializeMessage(msg redis.XMessage) (*TransportRequest, error) {
	// Try format 1: CBOR-encoded data in "data" field (legacy Go-to-Go format)
	if dataRaw, ok := msg.Values["data"]; ok {
		var data []byte
		switch v := dataRaw.(type) {
		case string:
			data = []byte(v)
		case []byte:
			data = v
		default:
			return nil, fmt.Errorf("invalid data type: %T", dataRaw)
		}
		transportReq, err := DecodeTransportRequest(data)
		if err != nil {
			return nil, err
		}
		// Set Redis message ID
		transportReq.RedisMessageID = msg.ID
		return transportReq, nil
	}

	// Format 2: Use broker serializer to deserialize from individual fields
	transportReq, err := b.serializer.Deserialize(msg.Values)
	if err != nil {
		return nil, err
	}
	// Set Redis message ID
	transportReq.RedisMessageID = msg.ID
	return transportReq, nil
}

// Stop stops the bus and all its listeners
func (b *Bus) Stop() {
	log.Println("Stopping bus...")
	b.cancel()
	b.wg.Wait()
	log.Println("Bus stopped")
}
