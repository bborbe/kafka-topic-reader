// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	libkafka "github.com/bborbe/kafka"
)

type Record struct {
	Key       string             `json:"key"`
	Value     interface{}        `json:"value"`
	Offset    libkafka.Offset    `json:"offset"`
	Partition libkafka.Partition `json:"partition"`
	Topic     libkafka.Topic     `json:"topic"`
	Header    libkafka.Header    `json:"header"`
}

type Records []Record
