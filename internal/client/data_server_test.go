package client

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"
	"testing"
	"time"
)

// TestDataPointIsValid tests the DataPoint.IsValid() method
func TestDataPointIsValid(t *testing.T) {
	tests := []struct {
		name      string
		datapoint DataPoint
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid datapoint",
			datapoint: DataPoint{
				Time: DataPointTime{
					Value:     time.Now(),
					Processed: true,
				},
				Value: DataPointValue{
					Value:     42.5,
					Processed: true,
				},
				Tags: Value[[]string]{
					Value:     []string{"tag1", "tag2"},
					Processed: true,
				},
			},
			expectErr: false,
		},
		{
			name: "time field not Processed",
			datapoint: DataPoint{
				Time: DataPointTime{
					Value:     time.Now(),
					Processed: false,
				},
				Value: DataPointValue{
					Value:     42.5,
					Processed: true,
				},
			},
			expectErr: true,
			errMsg:    "time field is not Processed",
		},
		{
			name: "Value field not Processed",
			datapoint: DataPoint{
				Time: DataPointTime{
					Value:     time.Now(),
					Processed: true,
				},
				Value: DataPointValue{
					Value:     42.5,
					Processed: false,
				},
			},
			expectErr: true,
			errMsg:    "Value field is not Processed",
		},
		{
			name: "neither time nor Value Processed",
			datapoint: DataPoint{
				Time: DataPointTime{
					Value:     time.Now(),
					Processed: false,
				},
				Value: DataPointValue{
					Value:     42.5,
					Processed: false,
				},
			},
			expectErr: true,
			errMsg:    "time field is not Processed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.datapoint.IsValid()

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if err != nil && err.Error() != tt.errMsg {
					t.Errorf("expected error message %q, got %q", tt.errMsg, err.Error())
				}
				if valid {
					t.Errorf("expected valid=false, got true")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !valid {
					t.Errorf("expected valid=true, got false")
				}
			}
		})
	}
}

// TestValueUnmarshalJSON tests the Value[T].UnmarshalJSON() method
func TestValueUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
		expectVal interface{}
	}{
		{
			name:      "unmarshal valid string",
			input:     `"test"`,
			expectErr: false,
			expectVal: "test",
		},
		{
			name:      "unmarshal valid number",
			input:     `42`,
			expectErr: false,
			expectVal: float64(42),
		},
		{
			name:      "unmarshal valid array",
			input:     `["a","b","c"]`,
			expectErr: false,
			expectVal: []string{"a", "b", "c"},
		},
		{
			name:      "unmarshal invalid JSON",
			input:     `{invalid}`,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v Value[interface{}]
			err := v.UnmarshalJSON([]byte(tt.input))

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !v.Processed {
					t.Errorf("expected Processed=true, got false")
				}
			}
		})
	}
}

// TestDataPointValueUnmarshalJSON tests the DataPointValue.UnmarshalJSON() method
func TestDataPointValueUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		expectErr bool
		expectVal float32
	}{
		{
			name:      "valid float32 bytes",
			input:     []byte(toJSONByteArray(float32ToBytes(3.14))),
			expectErr: false,
			expectVal: 3.14,
		},
		{
			name:      "valid zero Value",
			input:     []byte(toJSONByteArray(float32ToBytes(0.0))),
			expectErr: false,
			expectVal: 0.0,
		},
		{
			name:      "valid negative Value",
			input:     []byte(toJSONByteArray(float32ToBytes(-42.5))),
			expectErr: false,
			expectVal: -42.5,
		},
		{
			name:      "valid large Value",
			input:     []byte(toJSONByteArray(float32ToBytes(1e6))),
			expectErr: false,
			expectVal: 1e6,
		},
		{
			name:      "invalid JSON - too few bytes",
			input:     []byte(`[1,2,3]`), // 3 bytes instead of 4
			expectErr: true,
		},
		{
			name:      "invalid JSON - too many bytes",
			input:     []byte(`[1,2,3,4,5]`), // 5 bytes instead of 4
			expectErr: true,
		},
		{
			name:      "invalid JSON - not a byte array",
			input:     []byte(`"not a byte array"`),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dpv DataPointValue
			err := dpv.UnmarshalJSON(tt.input)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !dpv.Processed {
					t.Errorf("expected Processed=true, got false")
				}
				// Allow for small floating point errors
				if math.Abs(float64(dpv.Value-tt.expectVal)) > 1e-5 {
					t.Errorf("expected Value %f, got %f", tt.expectVal, dpv.Value)
				}
			}
		})
	}
}

// TestDecodeDatapointBody tests the decodeDatapointBody() method
func TestDecodeDatapointBody(t *testing.T) {
	tests := []struct {
		name      string
		body      io.Reader
		expectErr bool
		errMsg    string
		validate  func(t *testing.T, dp DataPoint)
	}{
		{
			name: "valid datapoint",
			body: createValidDataPointBody(),
			validate: func(t *testing.T, dp DataPoint) {
				if !dp.Time.Processed {
					t.Errorf("expected time.Processed=true, got false")
				}
				if !dp.Value.Processed {
					t.Errorf("expected Value.Processed=true, got false")
				}
				if !dp.Tags.Processed {
					t.Errorf("expected tags.Processed=true, got false")
				}
			},
		},
		{
			name:      "invalid JSON",
			body:      strings.NewReader(`{invalid json}`),
			expectErr: true,
			errMsg:    "failed to decode response body",
		},
		{
			name:      "empty body",
			body:      strings.NewReader(``),
			expectErr: true,
			errMsg:    "failed to decode response body",
		},
		{
			name:      "missing time field",
			body:      createDataPointBodyWithoutTime(),
			expectErr: true,
			errMsg:    "invalid datapoint received",
		},
		{
			name:      "missing Value field",
			body:      createDataPointBodyWithoutValue(),
			expectErr: true,
			errMsg:    "invalid datapoint received",
		},
		{
			name:      "invalid Value field (wrong byte length)",
			body:      createDataPointBodyWithInvalidValue(),
			expectErr: true,
			errMsg:    "failed to decode response body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &DataServerClient{url: "http://example.com"}
			dp, err := client.decodeDatapointBody(tt.body)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if err != nil && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.validate != nil {
					tt.validate(t, dp)
				}
			}
		})
	}
}

// Helper functions

func float32ToBytes(f float32) []byte {
	bits := math.Float32bits(f)
	return []byte{
		byte(bits & 0xFF),
		byte((bits >> 8) & 0xFF),
		byte((bits >> 16) & 0xFF),
		byte((bits >> 24) & 0xFF),
	}
}

func toJSONByteArray(b []byte) string {
	jsonBytes, _ := json.Marshal(b)
	return string(jsonBytes)
}

func createValidDataPointBody() io.Reader {
	now := time.Now()
	timeStr := now.Format(time.RFC3339Nano)
	valueBytes := float32ToBytes(3.14)

	// Create a custom JSON manually to have correct structure
	bodyJSON := fmt.Sprintf(`{"time":"%s","Value":%s,"tags":["tag1","tag2"]}`,
		timeStr,
		toJSONByteArray(valueBytes),
	)

	return strings.NewReader(bodyJSON)
}

func createDataPointBodyWithoutTime() io.Reader {
	valueBytes := float32ToBytes(3.14)

	bodyJSON := fmt.Sprintf(`{"Value":%s,"tags":["tag1","tag2"]}`,
		toJSONByteArray(valueBytes),
	)

	return strings.NewReader(bodyJSON)
}

func createDataPointBodyWithoutValue() io.Reader {
	now := time.Now()
	timeStr := now.Format(time.RFC3339Nano)

	bodyJSON := fmt.Sprintf(`{"time":"%s","tags":["tag1","tag2"]}`, timeStr)

	return strings.NewReader(bodyJSON)
}

func createDataPointBodyWithInvalidValue() io.Reader {
	now := time.Now()
	timeStr := now.Format(time.RFC3339Nano)
	invalidValueBytes := []byte{1, 2, 3} // Only 3 bytes instead of 4

	bodyJSON := fmt.Sprintf(`{"time":"%s","Value":%s,"tags":["tag1","tag2"]}`,
		timeStr,
		toJSONByteArray(invalidValueBytes),
	)

	return strings.NewReader(bodyJSON)
}
