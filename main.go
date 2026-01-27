// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"os"
	"time"

	"github.com/bborbe/errors"
	libhttp "github.com/bborbe/http"
	libkafka "github.com/bborbe/kafka"
	"github.com/bborbe/log"
	"github.com/bborbe/run"
	"github.com/bborbe/sentry"
	"github.com/bborbe/service"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/bborbe/kafka-topic-reader/pkg/factory"
)

func main() {
	app := &application{}
	os.Exit(service.Main(context.Background(), app, &app.SentryDSN, &app.SentryProxy))
}

type application struct {
	SentryDSN                 string            `required:"true"  arg:"sentry-dsn"                   env:"SENTRY_DSN"                   usage:"SentryDSN"                                                               display:"length"`
	SentryProxy               string            `required:"false" arg:"sentry-proxy"                 env:"SENTRY_PROXY"                 usage:"Sentry Proxy"`
	Listen                    string            `required:"true"  arg:"listen"                       env:"LISTEN"                       usage:"address to listen to"`
	KafkaBrokers              string            `required:"true"  arg:"kafka-brokers"                env:"KAFKA_BROKERS"                usage:"Comma separated list of Kafka brokers"`
	ErrorPreviewContentLength int               `required:"false" arg:"error-preview-content-length" env:"ERROR_PREVIEW_CONTENT_LENGTH" usage:"Maximum length in bytes for error message preview. Use -1 for unlimited"                  default:"100"`
	BuildGitCommit            string            `required:"false" arg:"build-git-commit" env:"BUILD_GIT_COMMIT" usage:"Build Git commit hash"                                  default:"none"`
	BuildDate                 *libtime.DateTime `required:"false" arg:"build-date"       env:"BUILD_DATE"       usage:"Build timestamp (RFC3339)"`
}

func (a *application) Run(
	ctx context.Context,
	sentryClient sentry.Client,
) error {
	libmetrics.NewBuildInfoMetrics().SetBuildInfo(a.BuildDate)

	saramaClient, err := libkafka.CreateSaramaClient(
		ctx,
		libkafka.ParseBrokersFromString(a.KafkaBrokers),
	)
	if err != nil {
		return errors.Wrapf(ctx, err, "create sarama client failed")
	}
	defer saramaClient.Close()

	return service.Run(
		ctx,
		a.createHTTPServer(sentryClient, saramaClient),
	)
}

func (a *application) createHTTPServer(
	sentryClient sentry.Client,
	saramaClient libkafka.SaramaClient,
) run.Func {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		router := mux.NewRouter()
		router.Path("/healthz").Handler(libhttp.NewPrintHandler("OK"))
		router.Path("/readiness").Handler(libhttp.NewPrintHandler("OK"))
		router.Path("/metrics").Handler(promhttp.Handler())
		router.Path("/setloglevel/{level}").Handler(
			log.NewSetLoglevelHandler(ctx, log.NewLogLevelSetter(2, 5*time.Minute)),
		)
		router.Path("/read").
			Handler(factory.CreateReadHandler(sentryClient, saramaClient, a.ErrorPreviewContentLength))

		glog.V(2).Infof("starting http server listen on %s", a.Listen)
		return libhttp.NewServer(
			a.Listen,
			router,
		).Run(ctx)
	}
}
