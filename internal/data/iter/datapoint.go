package iter

import (
	"fmt"
	"oc-data-be-challenge/internal/data/dto"
	"time"

	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
)

type DataPointIter struct {
	iterator *influxdb3.QueryIterator
}

func NewDataPointIter(iterator *influxdb3.QueryIterator) *DataPointIter {
	return &DataPointIter{iterator: iterator}
}

func (dpIter *DataPointIter) Next() bool {
	return dpIter.iterator.Next()
}

func (dpIter *DataPointIter) Value() (dto.DataPoint, error) {
	t, ok := dpIter.iterator.Value()["time"].(time.Time)
	if !ok {
		return dto.DataPoint{}, fmt.Errorf("failed to parse time from iterator %v", dpIter.iterator.Value()["time"])
	}

	val, ok := dpIter.iterator.Value()["value"].(float64)
	if !ok {
		return dto.DataPoint{}, fmt.Errorf("failed to parse value from iterator %v", dpIter.iterator.Value()["value"])
	}

	return dto.DataPoint{
		Time:  t,
		Value: float32(val),
	}, nil
}
