// Copyright (c) 2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	libtime "github.com/bborbe/time"
	"github.com/prometheus/client_golang/prometheus"
)

//counterfeiter:generate -o ../mocks/build-info-metrics.go --fake-name BuildInfoMetrics . BuildInfoMetrics
type BuildInfoMetrics interface {
	SetBuildInfo(buildDate *libtime.DateTime)
}

func NewBuildInfoMetrics(
	registerer prometheus.Registerer,
	namespace string,
) BuildInfoMetrics {
	b := &buildInfoMetrics{
		buildInfo: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "build_info",
				Help:      "Build timestamp as Unix time. Service identified by Prometheus job label.",
			},
		),
	}
	registerer.MustRegister(b.buildInfo)
	return b
}

type buildInfoMetrics struct {
	buildInfo prometheus.Gauge
}

func (m *buildInfoMetrics) SetBuildInfo(buildDate *libtime.DateTime) {
	if buildDate == nil {
		return
	}
	m.buildInfo.Set(float64(buildDate.Unix()))
}
