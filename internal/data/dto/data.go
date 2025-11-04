package dto

import "time"

type DataPoint struct {
	Time       time.Time `json:"time,omitempty"`
	Value      float32   `json:"value,omitempty"`
	Tags       []string  `json:"tags,omitempty"`
	ReceivedAt time.Time `json:"received_at,omitempty"`
}
