package bus

import (
	"fmt"
	"strconv"
)

// BrokerSerialize defines the interface for serializing/deserializing TransportRequest
type BrokerSerialize interface {
	// Serialize converts TransportRequest to a map for Redis message
	Serialize(request *TransportRequest) (map[string]interface{}, error)
	// Deserialize converts Redis message map to TransportRequest
	Deserialize(messageData map[string]interface{}) (*TransportRequest, error)
}

// RedisBrokerSerialize implements BrokerSerialize for Redis messages
// Matches Python's RedisBrokerSerialize behavior
type RedisBrokerSerialize struct {
	requiredAttrs []string
}

// NewRedisBrokerSerialize creates a new RedisBrokerSerialize instance
func NewRedisBrokerSerialize() *RedisBrokerSerialize {
	return &RedisBrokerSerialize{
		requiredAttrs: []string{"i", "r", "p"},
	}
}

// Serialize converts TransportRequest to a map for Redis message
func (s *RedisBrokerSerialize) Serialize(request *TransportRequest) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Required fields
	result["i"] = request.RequestID
	result["r"] = strconv.Itoa(request.ReturnResult)
	result["p"] = string(request.Properties)

	// Optional fields
	if len(request.Message) > 0 {
		result["m"] = string(request.Message)
	}
	if request.Timeout > 0 {
		result["t"] = strconv.Itoa(request.Timeout)
	}
	if request.CreatedTimestamp > 0 {
		result["c"] = strconv.FormatFloat(request.CreatedTimestamp, 'f', -1, 64)
	}

	return result, nil
}

// Deserialize converts Redis message map to TransportRequest
func (s *RedisBrokerSerialize) Deserialize(messageData map[string]interface{}) (*TransportRequest, error) {
	// Check required attributes
	if err := s.checkRequiredAttrs(messageData); err != nil {
		return nil, err
	}

	req := &TransportRequest{}

	// Extract RequestID ("i") - required
	if val, ok := messageData["i"]; ok {
		v, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("invalid RequestID type: %T, expected string", val)
		}
		req.RequestID = v
	}

	// Extract ReturnResult ("r") - required, string only
	if val, ok := messageData["r"]; ok {
		v, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("invalid ReturnResult type: %T, expected string", val)
		}
		result, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid ReturnResult string format: %q", v)
		}
		req.ReturnResult = result
	}

	// Extract Properties ("p") - required
	if val, ok := messageData["p"]; ok {
		switch v := val.(type) {
		case []byte:
			req.Properties = v
		case string:
			req.Properties = []byte(v)
		default:
			return nil, fmt.Errorf("invalid Properties type: %T", val)
		}
	}

	// Extract Message ("m") - optional, string only
	if val, ok := messageData["m"]; ok {
		v, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("invalid Message type: %T, expected string", val)
		}
		req.Message = []byte(v)
	}

	// Extract Timeout ("t") - optional, string only
	if val, ok := messageData["t"]; ok {
		v, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("invalid Timeout type: %T, expected string", val)
		}
		timeout, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid Timeout string format: %q", v)
		}
		req.Timeout = timeout
	} else {
		req.Timeout = DefaultTimeout
	}

	// Extract CreatedTimestamp ("c") - optional, string only
	if val, ok := messageData["c"]; ok {
		v, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("invalid CreatedTimestamp type: %T, expected string", val)
		}
		timestamp, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid CreatedTimestamp string format: %q", v)
		}
		req.CreatedTimestamp = timestamp
	}

	return req, nil
}

// checkRequiredAttrs validates that all required attributes are present
func (s *RedisBrokerSerialize) checkRequiredAttrs(messageData map[string]interface{}) error {
	for _, attr := range s.requiredAttrs {
		if _, ok := messageData[attr]; !ok {
			return fmt.Errorf("attribute %q not found in message data", attr)
		}
	}
	return nil
}
