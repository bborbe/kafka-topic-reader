// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg_test

import (
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
