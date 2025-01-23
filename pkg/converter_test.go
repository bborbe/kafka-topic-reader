// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg_test

import (
	"context"

	"github.com/IBM/sarama"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/kafka-topic-reader/pkg"
)

var _ = Describe("Converter", func() {
	var ctx context.Context
	var err error
	var converter pkg.Converter
	var msg *sarama.ConsumerMessage
	var record *pkg.Record
	BeforeEach(func() {
		ctx = context.Background()
		converter = pkg.NewConverter()
		msg = &sarama.ConsumerMessage{
			Headers: []*sarama.RecordHeader{
				{
					Key:   []byte("k"),
					Value: []byte("v"),
				},
			},
			Key:   []byte(`my-key`),
			Value: []byte(`{"a":"b"}`),
		}
	})
	Context("Convert", func() {
		JustBeforeEach(func() {
			record, err = converter.Convert(ctx, msg)
		})
		Context("success", func() {
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
			It("returns record", func() {
				Expect(record).NotTo(BeNil())
			})
			It("returns record with correct key", func() {
				Expect(record).NotTo(BeNil())
				Expect(record.Key).To(Equal("my-key"))
			})
			It("returns record with correct value", func() {
				Expect(record).NotTo(BeNil())
				Expect(record.Value).To(HaveKeyWithValue("a", "b"))
			})
			It("has header", func() {
				Expect(record).NotTo(BeNil())
				Expect(record.Header).To(HaveLen(1))
				Expect(record.Header).To(HaveKey("k"))
				Expect(record.Header.Get("k")).To(Equal("v"))
			})
		})
		Context("unparse able json", func() {
			BeforeEach(func() {
				msg.Value = []byte("banana")
			})
			It("returns record with error msg in value", func() {
				Expect(record).NotTo(BeNil())
				Expect(record.Value).To(Equal("unmarshal json failed: invalid character 'b' looking for beginning of value"))
			})
		})
	})
})
