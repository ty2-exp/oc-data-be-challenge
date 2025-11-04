package client

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"
)

// DataPoint represents a data point received from the data server.
type DataPoint struct {
	Time  DataPointTime   `json:"time"`
	Value DataPointValue  `json:"value"`
	Tags  Value[[]string] `json:"tags"`
}

func (dp DataPoint) IsValid() (bool, error) {
	if !dp.Time.Processed {
		return false, errors.New("time field is not Processed")
	}

	if !dp.Value.Processed {
		return false, errors.New("value field is not Processed")
	}

	if !dp.Tags.Processed {
		return false, errors.New("tags field is not Processed")
	}

	return true, nil
}

// Value is a generic type that wraps a value and indicates whether it has been processed.
type Value[T any] struct {
	Value     T
	Processed bool
}

func (t *Value[T]) UnmarshalJSON(data []byte) error {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("failed to unmarshal value, raw: %s, err:%w", string(data), err)
	}
	*t = Value[T]{
		Value:     v,
		Processed: true,
	}

	return nil
}

// DataPointTime represents the Time of a data point.
type DataPointTime Value[time.Time]

func (t *DataPointTime) UnmarshalJSON(data []byte) error {
	ts, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse DataPointTime, raw: %s, err:%w", string(data), err)
	}

	*t = DataPointTime{
		Value:     time.Unix(ts, 0),
		Processed: true,
	}
	return nil
}

// DataPointValue represents the Value of a data point, specifically a float32.
type DataPointValue Value[float32]

func (t *DataPointValue) UnmarshalJSON(data []byte) error {
	var b []byte
	if err := json.Unmarshal(data, &b); err != nil {
		return fmt.Errorf("failed to unmarshal DataPointValue, raw: %s, err:%w", string(data), err)
	}

	if len(b) != 4 {
		return fmt.Errorf("invalid data length for DataPointValue, expected 4 bytes, got %d bytes", len(data))
	}

	v := math.Float32frombits(binary.LittleEndian.Uint32(b))
	*t = DataPointValue{
		Value:     v,
		Processed: true,
	}

	return nil
}

// DataServerClient is a client for fetching data points from a data server.
type DataServerClient struct {
	url    string
	client *http.Client
}

// NewDataServerClient creates a new DataServerClient with the given URL and HTTP client.
func NewDataServerClient(url string, client *http.Client) *DataServerClient {
	if client == nil {
		client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return &DataServerClient{url: url, client: client}
}

// DataPoint fetches a data point from the data server.
func (ds *DataServerClient) DataPoint(ctx context.Context) (DataPoint, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ds.url, nil)
	if err != nil {
		return DataPoint{}, fmt.Errorf("failed to create request, %w", err)
	}

	resp, err := ds.client.Do(req)
	if err != nil {
		return DataPoint{}, fmt.Errorf("failed to perform request, %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return DataPoint{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	datapoint, err := ds.decodeDatapointBody(resp.Body)
	if err != nil {
		return DataPoint{}, fmt.Errorf("failed to decode datapoint body, %w", err)
	}

	if valid, err := datapoint.IsValid(); !valid {
		return DataPoint{}, fmt.Errorf("invalid datapoint received: %v", err)
	}

	return datapoint, nil
}

// decodeDatapointBody decodes the response body into a DataPoint.
func (ds *DataServerClient) decodeDatapointBody(r io.Reader) (DataPoint, error) {
	datapoint := DataPoint{}

	bodyDecoder := json.NewDecoder(r)
	err := bodyDecoder.Decode(&datapoint)
	if err != nil {
		return DataPoint{}, fmt.Errorf("failed to decode response body, %w", err)
	}

	if valid, err := datapoint.IsValid(); !valid {
		return DataPoint{}, fmt.Errorf("invalid datapoint received: %v", err)
	}

	return datapoint, nil
}
