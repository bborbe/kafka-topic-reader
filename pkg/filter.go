// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"bytes"

	"github.com/IBM/sarama"
)

// MatchesFilter checks if a Kafka message matches the given filter bytes.
// If filter is empty, all messages match. Otherwise, performs exact
// byte substring search in the raw binary message value.
func MatchesFilter(msg *sarama.ConsumerMessage, filter []byte) bool {
	if len(filter) == 0 {
		return true // No filtering
	}

	if len(msg.Value) == 0 {
		return false // No value to search
	}

	// Exact byte matching for binary data
	return bytes.Contains(msg.Value, filter)
}
