// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg_test

import (
	"github.com/IBM/sarama"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/kafka-topic-reader/pkg"
)

var _ = Describe("MatchesFilter", func() {
	var msg *sarama.ConsumerMessage
	var filter []byte
	var result bool

	BeforeEach(func() {
		msg = &sarama.ConsumerMessage{
			Key:       []byte("test-key"),
			Value:     []byte(`{"message": "Hello World", "status": "active"}`),
			Topic:     "test-topic",
			Partition: 0,
			Offset:    123,
		}
	})

	JustBeforeEach(func() {
		result = pkg.MatchesFilter(msg, filter)
	})

	Context("with empty filter", func() {
		BeforeEach(func() {
			filter = []byte{}
		})

		It("returns true for any message", func() {
			Expect(result).To(BeTrue())
		})
	})

	Context("binary value matching", func() {
		BeforeEach(func() {
			msg.Value = []byte("simple string value")
		})

		Context("exact substring match", func() {
			BeforeEach(func() {
				filter = []byte("string")
			})

			It("matches exact substring", func() {
				Expect(result).To(BeTrue())
			})
		})

		Context("case sensitive match", func() {
			BeforeEach(func() {
				filter = []byte("STRING")
			})

			It("does not match different case", func() {
				Expect(result).To(BeFalse())
			})
		})

		Context("no match", func() {
			BeforeEach(func() {
				filter = []byte("nonexistent")
			})

			It("returns false", func() {
				Expect(result).To(BeFalse())
			})
		})
	})

	Context("JSON binary value matching", func() {
		BeforeEach(func() {
			msg.Value = []byte(`{"user":"john.doe","message":"This is a test message","count":42}`)
		})

		Context("matching field name", func() {
			BeforeEach(func() {
				filter = []byte("user")
			})

			It("finds field names in JSON bytes", func() {
				Expect(result).To(BeTrue())
			})
		})

		Context("matching field value", func() {
			BeforeEach(func() {
				filter = []byte("john.doe")
			})

			It("finds field values in JSON bytes", func() {
				Expect(result).To(BeTrue())
			})
		})

		Context("matching partial field value", func() {
			BeforeEach(func() {
				filter = []byte("test message")
			})

			It("finds partial matches in JSON bytes", func() {
				Expect(result).To(BeTrue())
			})
		})

		Context("matching number as string", func() {
			BeforeEach(func() {
				filter = []byte("42")
			})

			It("finds numeric values in JSON bytes", func() {
				Expect(result).To(BeTrue())
			})
		})
	})

	Context("with empty value", func() {
		BeforeEach(func() {
			msg.Value = []byte{}
			filter = []byte("anything")
		})

		It("handles empty values gracefully", func() {
			Expect(result).To(BeFalse())
		})
	})

	Context("with nil value", func() {
		BeforeEach(func() {
			msg.Value = nil
			filter = []byte("anything")
		})

		It("handles nil values gracefully", func() {
			Expect(result).To(BeFalse())
		})
	})

	Context("complex nested JSON structures", func() {
		BeforeEach(func() {
			msg.Value = []byte(`{
				"data": {
					"nested": [
						"first item",
						{"key": "target_value"},
						42
					]
				},
				"metadata": {
					"version": "1.0",
					"tags": ["important", "filtered"]
				}
			}`)
		})

		Context("searching in nested objects", func() {
			BeforeEach(func() {
				filter = []byte("target_value")
			})

			It("finds values in nested JSON structures", func() {
				Expect(result).To(BeTrue())
			})
		})

		Context("searching in arrays", func() {
			BeforeEach(func() {
				filter = []byte("first item")
			})

			It("finds values in JSON arrays", func() {
				Expect(result).To(BeTrue())
			})
		})

		Context("searching for array values", func() {
			BeforeEach(func() {
				filter = []byte("important")
			})

			It("finds values in JSON string arrays", func() {
				Expect(result).To(BeTrue())
			})
		})

		Context("case sensitive search", func() {
			BeforeEach(func() {
				filter = []byte("FILTERED")
			})

			It("does not match different case in binary data", func() {
				Expect(result).To(BeFalse())
			})
		})
	})

	Context("binary data patterns", func() {
		BeforeEach(func() {
			// Test with actual binary data (not text)
			msg.Value = []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
		})

		Context("matching binary sequence", func() {
			BeforeEach(func() {
				filter = []byte{0x01, 0x02, 0xFF}
			})

			It("finds binary patterns", func() {
				Expect(result).To(BeTrue())
			})
		})

		Context("non-matching binary sequence", func() {
			BeforeEach(func() {
				filter = []byte{0x03, 0x04}
			})

			It("does not find non-existent patterns", func() {
				Expect(result).To(BeFalse())
			})
		})
	})
})
