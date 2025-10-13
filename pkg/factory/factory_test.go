// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package factory_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/kafka-topic-reader/pkg/factory"
)

var _ = Describe("Factory", func() {
	Context("CreateReadHandler", func() {
		It("returns a non-nil http.Handler", func() {
			handler := factory.CreateReadHandler(nil, nil, 100)
			Expect(handler).NotTo(BeNil())
		})

		It("implements http.Handler interface", func() {
			handler := factory.CreateReadHandler(nil, nil, 100)
			// Verify it implements http.Handler by using it as one
			var _ http.Handler = handler
			Expect(handler).NotTo(BeNil())
		})

		It("creates handler with factory pattern", func() {
			// Test that the factory can create the handler even with nil dependencies
			// This verifies the wiring is correct
			handler := factory.CreateReadHandler(nil, nil, 100)
			Expect(handler).NotTo(BeNil())
		})
	})
})
