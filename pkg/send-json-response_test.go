// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/kafka-topic-reader/pkg"
)

var _ = Describe("SendJSONResponse", func() {
	var response *httptest.ResponseRecorder
	var data interface{}
	var statusCode int
	var err error

	BeforeEach(func() {
		response = httptest.NewRecorder()
	})

	JustBeforeEach(func() {
		err = pkg.SendJSONResponse(response, data, statusCode)
	})

	Context("with valid data", func() {
		BeforeEach(func() {
			data = map[string]interface{}{
				"message": "test",
				"count":   42,
			}
			statusCode = http.StatusOK
		})

		It("returns no error", func() {
			Expect(err).To(BeNil())
		})

		It("sets correct content type", func() {
			Expect(response.Header().Get("Content-Type")).To(Equal("application/json"))
		})

		It("sets correct status code", func() {
			Expect(response.Code).To(Equal(http.StatusOK))
		})

		It("writes JSON data to response", func() {
			var result map[string]interface{}
			err := json.Unmarshal(response.Body.Bytes(), &result)
			Expect(err).To(BeNil())
			Expect(result["message"]).To(Equal("test"))
			Expect(result["count"]).To(BeNumerically("==", 42))
		})
	})

	Context("with nil data", func() {
		BeforeEach(func() {
			data = nil
			statusCode = http.StatusOK
		})

		It("returns no error", func() {
			Expect(err).To(BeNil())
		})

		It("writes null to response", func() {
			Expect(response.Body.String()).To(Equal("null\n"))
		})
	})

	Context("with empty struct", func() {
		BeforeEach(func() {
			data = struct{}{}
			statusCode = http.StatusCreated
		})

		It("returns no error", func() {
			Expect(err).To(BeNil())
		})

		It("sets correct status code", func() {
			Expect(response.Code).To(Equal(http.StatusCreated))
		})

		It("writes empty JSON object", func() {
			Expect(response.Body.String()).To(Equal("{}\n"))
		})
	})

	Context("with slice data", func() {
		BeforeEach(func() {
			data = []string{"item1", "item2", "item3"}
			statusCode = http.StatusAccepted
		})

		It("returns no error", func() {
			Expect(err).To(BeNil())
		})

		It("sets correct status code", func() {
			Expect(response.Code).To(Equal(http.StatusAccepted))
		})

		It("writes JSON array", func() {
			var result []string
			err := json.Unmarshal(response.Body.Bytes(), &result)
			Expect(err).To(BeNil())
			Expect(result).To(Equal([]string{"item1", "item2", "item3"}))
		})
	})

	Context("with different status codes", func() {
		BeforeEach(func() {
			data = "test"
		})

		Context("400 Bad Request", func() {
			BeforeEach(func() {
				statusCode = http.StatusBadRequest
			})

			It("sets correct status code", func() {
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("500 Internal Server Error", func() {
			BeforeEach(func() {
				statusCode = http.StatusInternalServerError
			})

			It("sets correct status code", func() {
				Expect(response.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})

	Context("with unencodable data", func() {
		BeforeEach(func() {
			// channels cannot be JSON-encoded
			data = make(chan int)
			statusCode = http.StatusOK
		})

		It("returns error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("json: unsupported type"))
		})

		It("still sets headers and status code", func() {
			Expect(response.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(response.Code).To(Equal(http.StatusOK))
		})
	})
})
