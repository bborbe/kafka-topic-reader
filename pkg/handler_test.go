// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/IBM/sarama"
	"github.com/bborbe/errors"
	libhttp "github.com/bborbe/http"
	libkafka "github.com/bborbe/kafka"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/kafka-topic-reader/mocks"
	"github.com/bborbe/kafka-topic-reader/pkg"
)

var _ = Describe("Handler", func() {
	var ctx context.Context
	var changesProvider *mocks.ChangesProvider
	var handler libhttp.WithError
	var request *http.Request
	var response *httptest.ResponseRecorder
	var err error

	BeforeEach(func() {
		ctx = context.Background()
		changesProvider = &mocks.ChangesProvider{}
		handler = pkg.NewHandler(changesProvider)
		response = httptest.NewRecorder()
	})

	Context("NewHandler", func() {
		It("returns handler", func() {
			Expect(handler).NotTo(BeNil())
		})
	})

	Context("ServeHTTP", func() {
		JustBeforeEach(func() {
			err = handler.ServeHTTP(ctx, response, request)
		})

		Context("missing topic parameter", func() {
			BeforeEach(func() {
				request = httptest.NewRequest("GET", "/read", nil)
			})

			It("returns error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("parameter topic missing"))
			})
		})

		Context("missing offset parameter", func() {
			BeforeEach(func() {
				values := url.Values{}
				values.Set("topic", "test-topic")
				request = httptest.NewRequest("GET", "/read?"+values.Encode(), nil)
			})

			It("returns error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("parse parameter offset failed"))
			})
		})

		Context("invalid offset parameter", func() {
			BeforeEach(func() {
				values := url.Values{}
				values.Set("topic", "test-topic")
				values.Set("offset", "invalid")
				request = httptest.NewRequest("GET", "/read?"+values.Encode(), nil)
			})

			It("returns error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("parse parameter offset failed"))
			})
		})

		Context("missing partition parameter", func() {
			BeforeEach(func() {
				values := url.Values{}
				values.Set("topic", "test-topic")
				values.Set("offset", "0")
				request = httptest.NewRequest("GET", "/read?"+values.Encode(), nil)
			})

			It("returns error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("parse parameter partition failed"))
			})
		})

		Context("invalid partition parameter", func() {
			BeforeEach(func() {
				values := url.Values{}
				values.Set("topic", "test-topic")
				values.Set("offset", "0")
				values.Set("partition", "invalid")
				request = httptest.NewRequest("GET", "/read?"+values.Encode(), nil)
			})

			It("returns error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("parse parameter partition failed"))
			})
		})

		Context("successful request with default limit", func() {
			var records pkg.Records

			BeforeEach(func() {
				values := url.Values{}
				values.Set("topic", "test-topic")
				values.Set("offset", "0")
				values.Set("partition", "0")
				request = httptest.NewRequest("GET", "/read?"+values.Encode(), nil)

				records = pkg.Records{
					{
						Key:       "test-key",
						Value:     map[string]interface{}{"test": "value"},
						Offset:    libkafka.Offset(0),
						Partition: libkafka.Partition(0),
						Topic:     libkafka.Topic("test-topic"),
					},
				}
				changesProvider.ChangesReturns(records, nil)
			})

			It("returns no error", func() {
				Expect(err).To(BeNil())
			})

			It("calls changes provider with correct parameters", func() {
				Expect(changesProvider.ChangesCallCount()).To(Equal(1))
				_, topic, partition, offset, limit, filter := changesProvider.ChangesArgsForCall(0)
				Expect(topic).To(Equal(libkafka.Topic("test-topic")))
				Expect(partition).To(Equal(libkafka.Partition(0)))
				Expect(offset).To(Equal(libkafka.Offset(0)))
				Expect(limit).To(Equal(uint64(100))) // default limit
				Expect(filter).To(Equal([]byte{}))   // no filter specified
			})

			It("returns OK status", func() {
				Expect(response.Code).To(Equal(http.StatusOK))
			})

			It("returns JSON content type", func() {
				Expect(response.Header().Get("Content-Type")).To(Equal("application/json"))
			})

			It("returns page with records", func() {
				body := response.Body.String()
				Expect(body).To(ContainSubstring("test-key"))
				Expect(body).To(ContainSubstring("nextOffset"))
			})
		})

		Context("successful request with custom limit", func() {
			BeforeEach(func() {
				values := url.Values{}
				values.Set("topic", "test-topic")
				values.Set("offset", "0")
				values.Set("partition", "0")
				values.Set("limit", "50")
				request = httptest.NewRequest("GET", "/read?"+values.Encode(), nil)

				changesProvider.ChangesReturns(pkg.Records{}, nil)
			})

			It("calls changes provider with custom limit", func() {
				Expect(changesProvider.ChangesCallCount()).To(Equal(1))
				_, _, _, _, limit, _ := changesProvider.ChangesArgsForCall(0)
				Expect(limit).To(Equal(uint64(50)))
			})
		})

		Context("invalid limit parameter", func() {
			BeforeEach(func() {
				values := url.Values{}
				values.Set("topic", "test-topic")
				values.Set("offset", "0")
				values.Set("partition", "0")
				values.Set("limit", "invalid")
				request = httptest.NewRequest("GET", "/read?"+values.Encode(), nil)

				changesProvider.ChangesReturns(pkg.Records{}, nil)
			})

			It("uses default limit of 100", func() {
				Expect(changesProvider.ChangesCallCount()).To(Equal(1))
				_, _, _, _, limit, _ := changesProvider.ChangesArgsForCall(0)
				Expect(limit).To(Equal(uint64(100)))
			})
		})

		Context("offset out of range error", func() {
			BeforeEach(func() {
				values := url.Values{}
				values.Set("topic", "test-topic")
				values.Set("offset", "1000")
				values.Set("partition", "0")
				request = httptest.NewRequest("GET", "/read?"+values.Encode(), nil)

				// First call returns offset out of range, second call succeeds
				changesProvider.ChangesReturnsOnCall(0, nil, sarama.ErrOffsetOutOfRange)
				changesProvider.ChangesReturnsOnCall(1, pkg.Records{}, nil)
			})

			It("retries with oldest offset", func() {
				Expect(changesProvider.ChangesCallCount()).To(Equal(2))

				// First call with original offset
				_, topic1, partition1, offset1, limit1, _ := changesProvider.ChangesArgsForCall(0)
				Expect(topic1).To(Equal(libkafka.Topic("test-topic")))
				Expect(partition1).To(Equal(libkafka.Partition(0)))
				Expect(offset1).To(Equal(libkafka.Offset(1000)))
				Expect(limit1).To(Equal(uint64(100)))

				// Second call with oldest offset
				_, topic2, partition2, offset2, limit2, _ := changesProvider.ChangesArgsForCall(1)
				Expect(topic2).To(Equal(libkafka.Topic("test-topic")))
				Expect(partition2).To(Equal(libkafka.Partition(0)))
				Expect(offset2).To(Equal(libkafka.OffsetOldest))
				Expect(limit2).To(Equal(uint64(100)))
			})

			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
		})

		Context("changes provider error (non-offset)", func() {
			BeforeEach(func() {
				values := url.Values{}
				values.Set("topic", "test-topic")
				values.Set("offset", "0")
				values.Set("partition", "0")
				request = httptest.NewRequest("GET", "/read?"+values.Encode(), nil)

				changesProvider.ChangesReturns(nil, errors.New(ctx, "provider error"))
			})

			It("returns error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("get changes failed"))
			})
		})

		Context("changes provider error after offset retry", func() {
			BeforeEach(func() {
				values := url.Values{}
				values.Set("topic", "test-topic")
				values.Set("offset", "1000")
				values.Set("partition", "0")
				request = httptest.NewRequest("GET", "/read?"+values.Encode(), nil)

				// First call returns offset out of range, second call also fails
				changesProvider.ChangesReturnsOnCall(0, nil, sarama.ErrOffsetOutOfRange)
				changesProvider.ChangesReturnsOnCall(1, nil, errors.New(ctx, "provider error"))
			})

			It("returns error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("get changes failed"))
			})
		})

		Context("next offset calculation", func() {
			Context("with records", func() {
				BeforeEach(func() {
					values := url.Values{}
					values.Set("topic", "test-topic")
					values.Set("offset", "5")
					values.Set("partition", "0")
					request = httptest.NewRequest("GET", "/read?"+values.Encode(), nil)

					records := pkg.Records{
						{Offset: libkafka.Offset(5)},
						{Offset: libkafka.Offset(6)},
						{Offset: libkafka.Offset(7)},
					}
					changesProvider.ChangesReturns(records, nil)
				})

				It("calculates next offset correctly", func() {
					body := response.Body.String()
					Expect(body).To(ContainSubstring(`"nextOffset":8`))
				})
			})

			Context("without records", func() {
				BeforeEach(func() {
					values := url.Values{}
					values.Set("topic", "test-topic")
					values.Set("offset", "5")
					values.Set("partition", "0")
					request = httptest.NewRequest("GET", "/read?"+values.Encode(), nil)

					changesProvider.ChangesReturns(pkg.Records{}, nil)
				})

				It("keeps original offset", func() {
					body := response.Body.String()
					Expect(body).To(ContainSubstring(`"nextOffset":5`))
				})
			})
		})

		Context("with filter parameter", func() {
			BeforeEach(func() {
				request = httptest.NewRequest(
					"GET",
					"/read?topic=test-topic&partition=0&offset=0&filter=test-value",
					nil,
				)
				records := pkg.Records{
					{Key: "key1", Value: "test-value here", Offset: libkafka.Offset(1)},
				}
				changesProvider.ChangesReturns(records, nil)
			})

			It("returns no error", func() {
				Expect(err).To(BeNil())
			})

			It("passes filter to changes provider", func() {
				Expect(changesProvider.ChangesCallCount()).To(Equal(1))
				_, _, _, _, _, filter := changesProvider.ChangesArgsForCall(0)
				Expect(filter).To(Equal([]byte("test-value")))
			})
		})

		Context("with empty filter parameter", func() {
			BeforeEach(func() {
				request = httptest.NewRequest(
					"GET",
					"/read?topic=test-topic&partition=0&offset=0&filter=",
					nil,
				)
				records := pkg.Records{
					{Key: "key1", Value: "any value", Offset: libkafka.Offset(1)},
				}
				changesProvider.ChangesReturns(records, nil)
			})

			It("passes empty filter to changes provider", func() {
				Expect(changesProvider.ChangesCallCount()).To(Equal(1))
				_, _, _, _, _, filter := changesProvider.ChangesArgsForCall(0)
				Expect(filter).To(Equal([]byte{}))
			})
		})

		Context("with filter parameter exceeding maximum length", func() {
			BeforeEach(func() {
				// Create a filter that exceeds 1024 bytes
				longFilter := make([]byte, 1025)
				for i := range longFilter {
					longFilter[i] = 'a'
				}
				request = httptest.NewRequest(
					"GET",
					"/read?topic=test-topic&partition=0&offset=0&filter="+string(longFilter),
					nil,
				)
			})

			It("returns an error for filter exceeding maximum length", func() {
				Expect(err).To(HaveOccurred())
				Expect(
					err.Error(),
				).To(ContainSubstring("filter parameter exceeds maximum length of 1024 bytes"))
			})

			It("does not call ChangesProvider", func() {
				Expect(changesProvider.ChangesCallCount()).To(Equal(0))
			})
		})
	})
})
