package bus

import (
	"fmt"
	"log"
	"sync"
)

// Repository defines the interface for repositories used by handlers
type Repository interface {
	// Repository methods can be defined by specific implementations
}

// HandlerConstructor is a function that creates a Subscriber from CBOR data and a repository
type HandlerConstructor func(data []byte, repo Repository) (Subscriber, error)

// HandlerFactory manages handler registration and creation with repositories
type HandlerFactory struct {
	mu           sync.RWMutex
	constructors map[string]HandlerConstructor
	repositories map[string]Repository
}

// NewHandlerFactory creates a new HandlerFactory instance
func NewHandlerFactory() *HandlerFactory {
	return &HandlerFactory{
		constructors: make(map[string]HandlerConstructor),
		repositories: make(map[string]Repository),
	}
}

// RegisterHandler registers a handler constructor for a specific stream
func (f *HandlerFactory) RegisterHandler(streamName string, constructor HandlerConstructor) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.constructors[streamName] = constructor
	log.Printf("HandlerFactory: Registered handler constructor for stream: %s", streamName)
}

// RegisterRepository registers a repository for a specific stream
func (f *HandlerFactory) RegisterRepository(streamName string, repo Repository) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.repositories[streamName] = repo
	log.Printf("HandlerFactory: Registered repository for stream: %s", streamName)
}

// CreateHandler creates a handler instance for the given stream using registered constructor and repository
func (f *HandlerFactory) CreateHandler(streamName string, data []byte) (Subscriber, error) {
	f.mu.RLock()
	constructor, hasConstructor := f.constructors[streamName]
	repo, hasRepo := f.repositories[streamName]
	f.mu.RUnlock()

	if !hasConstructor {
		return nil, fmt.Errorf("no handler constructor registered for stream: %s", streamName)
	}

	// Repository is optional - pass nil if not registered
	if !hasRepo {
		log.Printf("HandlerFactory: No repository registered for stream: %s, creating handler without repository", streamName)
	}

	return constructor(data, repo)
}

// HasHandler checks if a handler is registered for the given stream
func (f *HandlerFactory) HasHandler(streamName string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, exists := f.constructors[streamName]
	return exists
}

// GetStreams returns all registered stream names
func (f *HandlerFactory) GetStreams() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	streams := make([]string, 0, len(f.constructors))
	for streamName := range f.constructors {
		streams = append(streams, streamName)
	}
	return streams
}
