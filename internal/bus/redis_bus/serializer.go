package redis_bus

// import (
// 	"fmt"
// 	"strconv"

// 	"github.com/as-master/redis_grpc_bridge/internal/domain/query"

// 	"github.com/fxamacker/cbor/v2"
// 	"github.com/redis/go-redis/v9"
// )

// type TransportRequest struct {
// 	Message          any
// 	RequestID        string
// 	ReturnResult     bool
// 	Properties       string
// 	Timeout          int
// 	CreatedTimestamp *float64
// }

// func NewTransportRequest(
// 	message any,
// 	requestID string,
// 	returnResult bool,
// 	properties string,
// 	timeout int,
// 	createdTimestamp *float64,
// ) *TransportRequest {
// 	return &TransportRequest{
// 		Message:          message,
// 		RequestID:        requestID,
// 		ReturnResult:     returnResult,
// 		Properties:       properties,
// 		Timeout:          timeout,
// 		CreatedTimestamp: createdTimestamp,
// 	}
// }

// // DecodeCBORToStruct — десериализует CBOR-данные в переданную структуру
// func DecodeCBORToStruct(data []byte, out interface{}) error {
// 	decMode, err := cbor.DecOptions{
// 		TimeTag: 0, // Не требуется тег для времени
// 	}.DecMode()
// 	if err != nil {
// 		return fmt.Errorf("failed to create CBOR decoder mode: %w", err)
// 	}

// 	err = decMode.Unmarshal(data, out)
// 	if err != nil {
// 		return fmt.Errorf("failed to decode CBOR into struct: %w", err)
// 	}
// 	return nil
// }

// func RadisMsgToTransportReq(
// 	messageData redis.XMessage,
// 	entity any,
// ) (*TransportRequest, error) {
// 	values := messageData.Values
// 	// 1. request_id
// 	requestID, ok := values["i"].(string)
// 	if !ok {
// 		return nil, fmt.Errorf("missing or invalid 'i'")
// 	}

// 	// 2. return_result
// 	returnRaw, ok := values["r"].(string)
// 	if !ok {
// 		return nil, fmt.Errorf("missing or invalid 'r'")
// 	}
// 	returnResult := returnRaw == "1"

// 	// 3. main payload
// 	payload, ok := values["p"].(string)
// 	if !ok {
// 		return nil, fmt.Errorf("missing or invalid 'p'")
// 	}
// 	// var decoded map[string]interface{}
// 	var q query.TrainLoadQuery
// 	if err := cbor.Unmarshal([]byte(payload), &q); err != nil {
// 		return nil, fmt.Errorf("failed to decode payload: %w", err)
// 	}

// 	// 4. optional properties
// 	var props string
// 	// if meta, ok := messageData["m"].(string); ok && len(meta) > 0 {
// 	// 	var metaObj MessagePropertiesCTO
// 	// 	if err := cbor.Unmarshal([]byte(meta), &metaObj); err != nil {
// 	// 		return nil, fmt.Errorf("failed to decode properties: %w", err)
// 	// 	}
// 	// 	props = &metaObj
// 	// }

// 	// 5. timeout
// 	timeout := 300 // default
// 	if tStr, ok := values["t"].(string); ok {
// 		if parsed, err := strconv.Atoi(tStr); err == nil {
// 			timeout = parsed
// 		}
// 	}

// 	// 6. created timestamp
// 	var created *float64
// 	if cStr, ok := values["c"].(string); ok {
// 		if ts, err := strconv.ParseFloat(cStr, 64); err == nil {
// 			created = &ts
// 		}
// 	}

// 	return NewTransportRequest(entity, requestID, returnResult, props, timeout, created), nil
// }
