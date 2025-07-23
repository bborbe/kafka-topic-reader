// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"
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
			record.Value = fmt.Sprintf("unmarshal json failed: %v", err)
		}
	}
	return &record, nil
}
