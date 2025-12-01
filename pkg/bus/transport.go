package bus

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/redis/go-redis/v9"
)

// TransportRequest represents a CQRS transport request from Python
// Serialized using CBOR with short keys to match Python's TransportRequest class
type TransportRequest struct {
	RedisMessageID string `cbor:"id"`
	// Created timestamp in Unix epoch format (e.g., 1714214741.926557)
	CreatedTimestamp float64 `cbor:"c"`
	// Request ID - hex string UUID (e.g., "8a55d93256964d0dbc2173e70b75bf2f")
	RequestID string `cbor:"i"`
	// Message - CBOR-encoded message class info
	Message []byte `cbor:"m"`
	// Properties - CBOR-encoded message payload/properties
	Properties []byte `cbor:"p"`
	// ReturnResult - 1 if response is needed, 0 otherwise
	ReturnResult int `cbor:"r"`
	// Timeout in seconds
	Timeout int `cbor:"t"`
}

// TransportResponse represents a CQRS transport response to Python
// Matches Python's TransportResponse attrs class
type TransportResponse struct {
	// Request ID that this response corresponds to
	ReqID string `cbor:"req_id"`
	// Result - result data (will be CBOR-encoded directly, optional)
	Result any `cbor:"result,omitempty"`
	// Error message (optional)
	Error string `cbor:"error,omitempty"`
	// Error class name (optional)
	ErrorClass string `cbor:"error_class,omitempty"`
}

// DecodeTransportRequest decodes a CBOR-encoded TransportRequest
func DecodeTransportRequest(data []byte) (*TransportRequest, error) {
	var req TransportRequest
	if err := cbor.Unmarshal(data, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// Encode encodes the TransportResponse to CBOR
func (r *TransportResponse) Encode() ([]byte, error) {
	return cbor.Marshal(r)
}

// DecodeProperties decodes the Properties field into the target struct
func (r *TransportRequest) DecodeProperties(target any) error {
	if len(r.Properties) == 0 {
		return nil
	}
	return cbor.Unmarshal(r.Properties, target)
}

// DecodeMessage decodes the Message field into the target struct
func (r *TransportRequest) DecodeMessage(target any) error {
	if len(r.Message) == 0 {
		return nil
	}
	return cbor.Unmarshal(r.Message, target)
}

// NeedsResponse returns true if the request expects a response
func (r *TransportRequest) NeedsResponse() bool {
	return r.ReturnResult == 1
}

// EncodeResult encodes any result value to CBOR bytes for TransportResponse
func EncodeResult(result any) ([]byte, error) {
	if result == nil {
		return nil, nil
	}
	return cbor.Marshal(result)
}

// extractTransportRequest extracts TransportRequest from Redis message
// Supports two formats:
// 1. CBOR-encoded data in "data" field (Go-to-Go messages)
// 2. Individual fields in msg.Values: "r", "p", "m", "t", "c" (Python messages)
// RequestID is extracted from msg.ID (Redis message ID)
func extractTransportRequest(msg redis.XMessage) (*TransportRequest, error) {
	// Try format 1: CBOR-encoded data in "data" field
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
		// Override RequestID with Redis message ID
		transportReq.RequestID = msg.ID
		return transportReq, nil
	}

	// Format 2: Individual fields directly in msg.Values
	req := &TransportRequest{}

	// Extract RequestID from Redis message ID
	req.RedisMessageID = msg.ID

	// Extract RequestID ("i") - string only
	if val, ok := msg.Values["i"]; ok {
		v, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("invalid RequestID type: %T, expected string", val)
		}
		req.RequestID = v
	}

	// Extract ReturnResult ("r") - string only
	if val, ok := msg.Values["r"]; ok {
		v, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("invalid ReturnResult type: %T, expected string", val)
		}
		var result int
		if _, err := fmt.Sscanf(v, "%d", &result); err != nil {
			return nil, fmt.Errorf("invalid ReturnResult string format: %q", v)
		}
		req.ReturnResult = result
	}

	// Extract Properties ("p")
	if val, ok := msg.Values["p"]; ok {
		switch v := val.(type) {
		case []byte:
			req.Properties = v
		case string:
			req.Properties = []byte(v)
		default:
			return nil, fmt.Errorf("invalid Properties type: %T", val)
		}
	}

	// Extract Message ("m") - string only
	if val, ok := msg.Values["m"]; ok {
		v, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("invalid Message type: %T, expected string", val)
		}
		req.Message = []byte(v)
	}

	// Extract Timeout ("t") - string only
	if val, ok := msg.Values["t"]; ok {
		v, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("invalid Timeout type: %T, expected string", val)
		}
		var timeout int
		if _, err := fmt.Sscanf(v, "%d", &timeout); err != nil {
			return nil, fmt.Errorf("invalid Timeout string format: %q", v)
		}
		req.Timeout = timeout
	}

	// Extract CreatedTimestamp ("c") - string only
	if val, ok := msg.Values["c"]; ok {
		v, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("invalid CreatedTimestamp type: %T, expected string", val)
		}
		var timestamp float64
		if _, err := fmt.Sscanf(v, "%f", &timestamp); err != nil {
			return nil, fmt.Errorf("invalid CreatedTimestamp string format: %q", v)
		}
		req.CreatedTimestamp = timestamp
	}

	return req, nil
}
