// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	libkafka "github.com/bborbe/kafka"
	"github.com/golang/glog"
)

//counterfeiter:generate -o ../mocks/converter.go --fake-name Converter . Converter
type Converter interface {
	Convert(ctx context.Context, msg *sarama.ConsumerMessage) (*Record, error)
}

func NewConverter() Converter {
	return &converter{}
}

type converter struct {
}

// Convert transforms a Sarama consumer message into a Record.
//
// If the message value cannot be unmarshaled as JSON, the value field will contain
// an error map with the following structure:
//   - error: string describing the JSON parsing error
//   - valueLength: total size of the original message in bytes
//   - previewBase64: base64-encoded preview of first 100 bytes
//   - previewHex: hex-encoded preview of first 100 bytes
//
// Both preview fields are limited to 100 bytes to prevent memory exhaustion
// from large malformed messages.
func (c *converter) Convert(ctx context.Context, msg *sarama.ConsumerMessage) (*Record, error) {
	record := Record{
		Key:       string(msg.Key),
		Offset:    libkafka.Offset(msg.Offset),
		Partition: libkafka.Partition(msg.Partition),
		Topic:     libkafka.Topic(msg.Topic),
		Header:    libkafka.ParseHeader(msg.Headers),
	}
	if len(msg.Value) != 0 {
		if err := json.Unmarshal(msg.Value, &record.Value); err != nil {
			glog.V(4).Infof("unmarshal json failed: %v", err)
			previewLength := min(100, len(msg.Value))
			record.Value = map[string]interface{}{
				"error":         fmt.Sprintf("unmarshal value as JSON failed: %v", err),
				"valueLength":   len(msg.Value),
				"previewBase64": base64.StdEncoding.EncodeToString(msg.Value[:previewLength]),
				"previewHex":    fmt.Sprintf("%x", msg.Value[:previewLength]),
			}
		}
	}
	return &record, nil
}
