// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg_test

import (
	libkafka "github.com/bborbe/kafka"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/kafka-topic-reader/pkg"
)

var _ = Describe("ChangesProvider", func() {
	Context("NewChangesProvider", func() {
		It("returns changes provider", func() {
			changesProvider := pkg.NewChangesProvider(nil, nil, nil, nil)
			Expect(changesProvider).NotTo(BeNil())
		})
	})
})

var _ = Describe("Record", func() {
	Context("struct fields", func() {
		It("has all required fields", func() {
			record := pkg.Record{
				Key:       "test-key",
				Value:     map[string]interface{}{"test": "value"},
				Offset:    libkafka.Offset(123),
				Partition: libkafka.Partition(0),
				Topic:     libkafka.Topic("test-topic"),
				Header:    libkafka.Header{},
			}

			Expect(record.Key).To(Equal("test-key"))
			Expect(record.Value).To(HaveKeyWithValue("test", "value"))
			Expect(int64(record.Offset)).To(Equal(int64(123)))
			Expect(int32(record.Partition)).To(Equal(int32(0)))
			Expect(string(record.Topic)).To(Equal("test-topic"))
			Expect(record.Header).NotTo(BeNil())
		})
	})
})

var _ = Describe("Records", func() {
	Context("slice operations", func() {
		It("can be created and manipulated", func() {
			records := pkg.Records{
				{Key: "key1", Offset: libkafka.Offset(1)},
				{Key: "key2", Offset: libkafka.Offset(2)},
			}

			Expect(records).To(HaveLen(2))
			Expect(records[0].Key).To(Equal("key1"))
			Expect(records[1].Key).To(Equal("key2"))
		})
	})
})
