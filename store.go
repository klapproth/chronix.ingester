package main

import (
	"fmt"
	"time"

	"github.com/ChronixDB/chronix.go/chronix"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/storage/local/chunk"
)

type chronixStore struct {
	chronix      chronix.Client
	commitWithin time.Duration
}

func (s *chronixStore) Put(metric model.Metric, descs []*chunk.Desc) error {
	for _, desc := range descs {
		ts, err := transcodeChunk(metric, desc)
		if err != nil {
			return fmt.Errorf("error transcoding chunk: %v", err)
		}
		if err := s.chronix.Store([]*chronix.TimeSeries{ts}, false, s.commitWithin); err != nil {
			return fmt.Errorf("error storing chunk: %v", err)
		}
	}
	return nil
}

func transcodeChunk(metric model.Metric, desc *chunk.Desc) (*chronix.TimeSeries, error) {
	ts := &chronix.TimeSeries{
		Name:       string(metric[model.MetricNameLabel]),
		Type:       "metric",
		Attributes: map[string]string{},
	}
	for k, v := range metric {
		if k == model.MetricNameLabel {
			continue
		}
		ts.Attributes[string(k)] = string(v)
	}

	it := desc.C.NewIterator()
	for it.Scan() {
		sp := it.Value()
		ts.Points = append(ts.Points, chronix.Point{
			Value:     float64(sp.Value),
			Timestamp: sp.Timestamp.UnixNano() / 1e6,
		})
	}
	if it.Err() != nil {
		return nil, it.Err()
	}
	return ts, nil
}
