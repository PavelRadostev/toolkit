package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/PavelRadostev/toolkit/pkg/bus"
	"github.com/PavelRadostev/toolkit/pkg/config"
	"github.com/fxamacker/cbor/v2"
	"github.com/redis/go-redis/v9"
)

// EnterpriseId represents enterprise ID
type EnterpriseId int

// Filetype represents file type
type Filetype string

// SchemaDto represents the schema DTO
type SchemaDto struct {
	Type string                            `cbor:"type"`
	Data map[string]map[string]interface{} `cbor:"data"`
}

// GGISImportTemplateEntity represents the main entity
type GGISImportTemplateEntity struct {
	ID         EnterpriseId `cbor:"id"`
	Name       string       `cbor:"name"`
	Enterprise EnterpriseId `cbor:"enterprise"`
	Filetype   Filetype     `cbor:"filetype"`
	Delimiter  string       `cbor:"delimiter"`
	Schema     SchemaDto    `cbor:"schema"`
}

// String returns string representation of the entity
func (e *GGISImportTemplateEntity) String() string {
	return fmt.Sprintf("GGISImportTemplateEntity{ID: %d, Name: %s, Enterprise: %d, Filetype: %s, Delimiter: %s}",
		e.ID, e.Name, e.Enterprise, e.Filetype, e.Delimiter)
}

// NewSubscriber creates a new instance of GGISImportTemplateEntity as Subscriber
func (e *GGISImportTemplateEntity) NewSubscriber() *GGISImportTemplateEntity {
	return &GGISImportTemplateEntity{}
}

// Handle processes the entity (implementation can be added as needed)
func (e *GGISImportTemplateEntity) Handle(ctx context.Context) (any, error) {
	// Add your business logic here
	// For example: process the import template, validate schema, etc.

	fmt.Printf("Handling GGISImportTemplateEntity: %s\n", e.String())

	// Return the processed result
	return map[string]interface{}{
		"processed": true,
		"entity":    e,
	}, nil
}

// Deserialize decodes CBOR data into the entity
func (e *GGISImportTemplateEntity) Deserialize(data []byte) error {
	return cbor.Unmarshal(data, e)
}

// Serialize encodes the entity to CBOR (optional helper method)
func (e *GGISImportTemplateEntity) Serialize() ([]byte, error) {
	return cbor.Marshal(e)
}

type AllGGISImportTemplatesQueryHandler struct {
	EnterpriseID EnterpriseId `cbor:"enterprise_id"`
}

func NewAllGGISImportTemplatesQueryFromCBOR(data []byte) (bus.Subscriber, error) {
	var handler AllGGISImportTemplatesQueryHandler
	if err := cbor.Unmarshal(data, &handler); err != nil {
		return nil, err
	}
	return &handler, nil
}

func (a *AllGGISImportTemplatesQueryHandler) Handle(ctx context.Context) (any, error) {

	fmt.Printf("HandlingAllGGISImportTemplatesQuery")

	// Всегда возвращаем успешный результат со случайными данными
	return map[string]any{
		"id":         EnterpriseId(rand.Intn(1000) + 1),
		"name":       fmt.Sprintf("ProcessedTemplate_%d", rand.Intn(100)),
		"enterprise": EnterpriseId(rand.Intn(100) + 1),
		"filetype":   Filetype("CSV"),
		"delimiter":  ",",
		"status":     "success",
		"processed":  true,
		"records":    rand.Intn(10000),
	}, nil
}

//IsPlanApprovedQueryHandler

type IsPlanApprovedQueryHandler struct {
	EnterpriseID EnterpriseId `cbor:"enterprise_id"`
}

func NewIsPlanApprovedQueryFromCBOR(data []byte) (bus.Subscriber, error) {
	var handler IsPlanApprovedQueryHandler
	if err := cbor.Unmarshal(data, &handler); err != nil {
		return nil, err
	}
	return &handler, nil
}

func (i *IsPlanApprovedQueryHandler) Handle(ctx context.Context) (any, error) {

	fmt.Printf("HandlingIsPlanApprovedQuery")

	return true, nil
}

func main() {
	ctx := context.Background()

	fmt.Println("Hello, World!")
	cfg := config.Load()
	fmt.Println(cfg)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	bus := bus.NewBus(redisClient, ctx)

	bus.Register("vist_domain.query.ggis_import.AllGGISImportTemplatesQuery", NewAllGGISImportTemplatesQueryFromCBOR)
	bus.Register("vist_domain.query.pit.plan.IsPlanApprovedQuery", NewIsPlanApprovedQueryFromCBOR)

	bus.Run()

	time.Sleep(time.Second * 600)

}
