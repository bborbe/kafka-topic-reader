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
		converter = pkg.NewConverter(100)
		msg = &sarama.ConsumerMessage{
			Headers: []*sarama.RecordHeader{
				{
					Key:   []byte("k"),
					Value: []byte("v"),
				},
			},
			Key:       []byte(`my-key`),
			Value:     []byte(`{"a":"b"}`),
			Topic:     "test-topic",
			Partition: 2,
			Offset:    123,
		}
	})

	Context("NewConverter", func() {
		It("returns converter", func() {
			Expect(converter).NotTo(BeNil())
		})
	})

	Context("Convert", func() {
		JustBeforeEach(func() {
			record, err = converter.Convert(ctx, msg)
		})

		Context("successful conversion with valid JSON", func() {
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

			It("returns record with correct topic", func() {
				Expect(record).NotTo(BeNil())
				Expect(string(record.Topic)).To(Equal("test-topic"))
			})

			It("returns record with correct partition", func() {
				Expect(record).NotTo(BeNil())
				Expect(int32(record.Partition)).To(Equal(int32(2)))
			})

			It("returns record with correct offset", func() {
				Expect(record).NotTo(BeNil())
				Expect(int64(record.Offset)).To(Equal(int64(123)))
			})

			It("has header", func() {
				Expect(record).NotTo(BeNil())
				Expect(record.Header).To(HaveLen(1))
				Expect(record.Header).To(HaveKey("k"))
				Expect(record.Header.Get("k")).To(Equal("v"))
			})
		})

		Context("with multiple headers", func() {
			BeforeEach(func() {
				msg.Headers = []*sarama.RecordHeader{
					{Key: []byte("header1"), Value: []byte("value1")},
					{Key: []byte("header2"), Value: []byte("value2")},
					{Key: []byte("header3"), Value: []byte("value3")},
				}
			})

			It("parses all headers correctly", func() {
				Expect(record.Header).To(HaveLen(3))
				Expect(record.Header.Get("header1")).To(Equal("value1"))
				Expect(record.Header.Get("header2")).To(Equal("value2"))
				Expect(record.Header.Get("header3")).To(Equal("value3"))
			})
		})

		Context("with no headers", func() {
			BeforeEach(func() {
				msg.Headers = nil
			})

			It("has empty header", func() {
				Expect(record.Header).To(HaveLen(0))
			})
		})

		Context("with empty key", func() {
			BeforeEach(func() {
				msg.Key = nil
			})

			It("returns record with empty key", func() {
				Expect(record.Key).To(Equal(""))
			})
		})

		Context("with complex JSON value", func() {
			BeforeEach(func() {
				msg.Value = []byte(`{
					"string": "test",
					"number": 42,
					"boolean": true,
					"null": null,
					"array": [1, 2, "three"],
					"object": {"nested": "value"}
				}`)
			})

			It("parses complex JSON correctly", func() {
				value, ok := record.Value.(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(value["string"]).To(Equal("test"))
				Expect(value["number"]).To(BeNumerically("==", 42))
				Expect(value["boolean"]).To(Equal(true))
				Expect(value["null"]).To(BeNil())

				array, ok := value["array"].([]interface{})
				Expect(ok).To(BeTrue())
				Expect(array).To(HaveLen(3))
				Expect(array[0]).To(BeNumerically("==", 1))
				Expect(array[1]).To(BeNumerically("==", 2))
				Expect(array[2]).To(Equal("three"))

				obj, ok := value["object"].(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(obj["nested"]).To(Equal("value"))
			})
		})

		Context("with empty value", func() {
			BeforeEach(func() {
				msg.Value = []byte{}
			})

			It("returns record with nil value", func() {
				Expect(record.Value).To(BeNil())
			})
		})

		Context("with nil value", func() {
			BeforeEach(func() {
				msg.Value = nil
			})

			It("returns record with nil value", func() {
				Expect(record.Value).To(BeNil())
			})
		})

		Context("with unparseable JSON", func() {
			BeforeEach(func() {
				msg.Value = []byte("banana")
			})

			It("returns record with structured error information", func() {
				Expect(record).NotTo(BeNil())
				errorMap, ok := record.Value.(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(errorMap).To(HaveKey("error"))
				Expect(errorMap["error"]).To(ContainSubstring("unmarshal value as JSON failed:"))
				Expect(errorMap).To(HaveKey("valueLength"))
				Expect(errorMap["valueLength"]).To(Equal(6))
				Expect(errorMap).To(HaveKey("previewBase64"))
				Expect(errorMap["previewBase64"]).To(Equal("YmFuYW5h"))
				Expect(errorMap).To(HaveKey("previewHex"))
				Expect(errorMap["previewHex"]).To(Equal("62616e616e61"))
			})
		})

		Context("with malformed JSON", func() {
			BeforeEach(func() {
				msg.Value = []byte(`{"incomplete": `)
			})

			It("returns record with structured error information", func() {
				Expect(record).NotTo(BeNil())
				errorMap, ok := record.Value.(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(errorMap).To(HaveKey("error"))
				Expect(errorMap["error"]).To(ContainSubstring("unmarshal value as JSON failed:"))
				Expect(errorMap).To(HaveKey("valueLength"))
				Expect(errorMap["valueLength"]).To(Equal(15))
				Expect(errorMap).To(HaveKey("previewBase64"))
				Expect(errorMap["previewBase64"]).To(Equal("eyJpbmNvbXBsZXRlIjog"))
				Expect(errorMap).To(HaveKey("previewHex"))
				Expect(errorMap["previewHex"]).To(Equal("7b22696e636f6d706c657465223a20"))
			})
		})

		Context("with exactly 100 bytes of unparseable data", func() {
			BeforeEach(func() {
				msg.Value = make([]byte, 100)
				for i := range msg.Value {
					msg.Value[i] = 'x'
				}
			})

			It("includes full preview without truncation", func() {
				Expect(record).NotTo(BeNil())
				errorMap, ok := record.Value.(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(errorMap["valueLength"]).To(Equal(100))
				Expect(errorMap["previewHex"]).To(HaveLen(200))
			})
		})

		Context("with more than 100 bytes of unparseable data", func() {
			BeforeEach(func() {
				msg.Value = make([]byte, 200)
				for i := range msg.Value {
					msg.Value[i] = 'y'
				}
			})

			It("truncates preview to 100 bytes", func() {
				Expect(record).NotTo(BeNil())
				errorMap, ok := record.Value.(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(errorMap["valueLength"]).To(Equal(200))
				Expect(errorMap["previewHex"]).To(HaveLen(200))
			})
		})

		Context("with binary non-UTF8 data", func() {
			BeforeEach(func() {
				msg.Value = []byte{0x00, 0xFF, 0xFE, 0xFD, 0xFC}
			})

			It("handles binary data safely", func() {
				Expect(record).NotTo(BeNil())
				errorMap, ok := record.Value.(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(errorMap["valueLength"]).To(Equal(5))
				Expect(errorMap["previewBase64"]).To(Equal("AP/+/fw="))
				Expect(errorMap["previewHex"]).To(Equal("00fffefdfc"))
			})
		})

		Context("with JSON array", func() {
			BeforeEach(func() {
				msg.Value = []byte(`[1, 2, "three", {"key": "value"}]`)
			})

			It("parses JSON array correctly", func() {
				array, ok := record.Value.([]interface{})
				Expect(ok).To(BeTrue())
				Expect(array).To(HaveLen(4))
				Expect(array[0]).To(BeNumerically("==", 1))
				Expect(array[1]).To(BeNumerically("==", 2))
				Expect(array[2]).To(Equal("three"))

				obj, ok := array[3].(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(obj["key"]).To(Equal("value"))
			})
		})

		Context("with JSON string value", func() {
			BeforeEach(func() {
				msg.Value = []byte(`"just a string"`)
			})

			It("parses JSON string correctly", func() {
				Expect(record.Value).To(Equal("just a string"))
			})
		})

		Context("with JSON number value", func() {
			BeforeEach(func() {
				msg.Value = []byte(`42`)
			})

			It("parses JSON number correctly", func() {
				Expect(record.Value).To(BeNumerically("==", 42))
			})
		})

		Context("with JSON boolean value", func() {
			BeforeEach(func() {
				msg.Value = []byte(`true`)
			})

			It("parses JSON boolean correctly", func() {
				Expect(record.Value).To(Equal(true))
			})
		})

		Context("with JSON null value", func() {
			BeforeEach(func() {
				msg.Value = []byte(`null`)
			})

			It("parses JSON null correctly", func() {
				Expect(record.Value).To(BeNil())
			})
		})

		Context("with configurable preview length", func() {
			Context("with preview length of 10", func() {
				BeforeEach(func() {
					converter = pkg.NewConverter(10)
					msg.Value = make([]byte, 50)
					for i := range msg.Value {
						msg.Value[i] = 'x'
					}
				})

				It("limits preview to 10 bytes", func() {
					Expect(record).NotTo(BeNil())
					errorMap, ok := record.Value.(map[string]interface{})
					Expect(ok).To(BeTrue())
					Expect(errorMap["valueLength"]).To(Equal(50))
					Expect(errorMap["previewHex"]).To(HaveLen(20)) // 10 bytes = 20 hex chars
				})
			})

			Context("with preview length of -1 (unlimited)", func() {
				BeforeEach(func() {
					converter = pkg.NewConverter(-1)
					msg.Value = make([]byte, 200)
					for i := range msg.Value {
						msg.Value[i] = 'y'
					}
				})

				It("includes full preview without truncation", func() {
					Expect(record).NotTo(BeNil())
					errorMap, ok := record.Value.(map[string]interface{})
					Expect(ok).To(BeTrue())
					Expect(errorMap["valueLength"]).To(Equal(200))
					Expect(errorMap["previewHex"]).To(HaveLen(400)) // 200 bytes = 400 hex chars
				})
			})

			Context("with preview length of 0", func() {
				BeforeEach(func() {
					converter = pkg.NewConverter(0)
					msg.Value = []byte("test")
				})

				It("includes empty preview", func() {
					Expect(record).NotTo(BeNil())
					errorMap, ok := record.Value.(map[string]interface{})
					Expect(ok).To(BeTrue())
					Expect(errorMap["valueLength"]).To(Equal(4))
					Expect(errorMap["previewHex"]).To(HaveLen(0))
					Expect(errorMap["previewBase64"]).To(Equal(""))
				})
			})

			Context("with preview length larger than value", func() {
				BeforeEach(func() {
					converter = pkg.NewConverter(1000)
					msg.Value = []byte("short")
				})

				It("includes full value", func() {
					Expect(record).NotTo(BeNil())
					errorMap, ok := record.Value.(map[string]interface{})
					Expect(ok).To(BeTrue())
					Expect(errorMap["valueLength"]).To(Equal(5))
					Expect(errorMap["previewHex"]).To(Equal("73686f7274"))
				})
			})
		})
	})
})
