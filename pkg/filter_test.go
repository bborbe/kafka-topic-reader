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

	DescribeTable("JSON binary value matching",
		func(msgValue []byte, filterValue []byte, expectedMatch bool) {
			msg.Value = msgValue
			filter = filterValue
			result := pkg.MatchesFilter(msg, filter)
			Expect(result).To(Equal(expectedMatch))
		},
		Entry(
			"finds field names in JSON bytes",
			[]byte(`{"user":"john.doe","message":"This is a test message","count":42}`),
			[]byte("user"),
			true,
		),
		Entry(
			"finds field values in JSON bytes",
			[]byte(`{"user":"john.doe","message":"This is a test message","count":42}`),
			[]byte("john.doe"),
			true,
		),
		Entry(
			"finds partial matches in JSON bytes",
			[]byte(`{"user":"john.doe","message":"This is a test message","count":42}`),
			[]byte("test message"),
			true,
		),
		Entry(
			"finds numeric values in JSON bytes",
			[]byte(`{"user":"john.doe","message":"This is a test message","count":42}`),
			[]byte("42"),
			true,
		),
	)

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

	DescribeTable("complex nested JSON structures",
		func(filterValue []byte, expectedMatch bool) {
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
			filter = filterValue
			result := pkg.MatchesFilter(msg, filter)
			Expect(result).To(Equal(expectedMatch))
		},
		Entry("finds values in nested JSON structures", []byte("target_value"), true),
		Entry("finds values in JSON arrays", []byte("first item"), true),
		Entry("finds values in JSON string arrays", []byte("important"), true),
		Entry("does not match different case in binary data", []byte("FILTERED"), false),
	)

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
