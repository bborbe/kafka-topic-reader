// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package factory

import (
	"net/http"

	libhttp "github.com/bborbe/http"
	libkafka "github.com/bborbe/kafka"
	"github.com/bborbe/log"
	"github.com/bborbe/sentry"

	"github.com/bborbe/kafka-topic-reader/pkg"
)

func CreateReadHandler(
	sentryClient sentry.Client,
	saramaClient libkafka.SaramaClient,
	errorPreviewContentLength int,
) http.Handler {
	return libhttp.NewErrorHandler(
		pkg.NewHandler(
			pkg.NewChangesProvider(
				sentryClient,
				saramaClient,
				pkg.NewConverter(errorPreviewContentLength),
				log.DefaultSamplerFactory,
			),
		),
	)
}
