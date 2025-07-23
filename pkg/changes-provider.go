// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"
	"runtime"

	"github.com/IBM/sarama"
	"github.com/bborbe/errors"
	libkafka "github.com/bborbe/kafka"
	"github.com/bborbe/log"
	"github.com/bborbe/run"
	"github.com/bborbe/sentry"
	"github.com/golang/glog"
)

//counterfeiter:generate -o ../mocks/changes-provider.go --fake-name ChangesProvider . ChangesProvider
type ChangesProvider interface {
	Changes(
		ctx context.Context,
		topic libkafka.Topic,
		partition libkafka.Partition,
		offset libkafka.Offset,
		limit uint64,
	) (Records, error)
}

func NewChangesProvider(
	sentryClient sentry.Client,
	saramaClient libkafka.SaramaClient,
	converter Converter,
	logSamplerFactory log.SamplerFactory,
) ChangesProvider {
	return &changesProvider{
		sentryClient:      sentryClient,
		saramaClient:      saramaClient,
		converter:         converter,
		logSamplerFactory: logSamplerFactory,
	}
}

type changesProvider struct {
	saramaClient      libkafka.SaramaClient
	converter         Converter
	sentryClient      sentry.Client
	logSamplerFactory log.SamplerFactory
}

func (c *changesProvider) Changes(
	ctx context.Context,
	topic libkafka.Topic,
	partition libkafka.Partition,
	offset libkafka.Offset,
	limit uint64,
) (Records, error) {
	var records Records
	ch := make(chan Record, runtime.NumCPU())
	err := run.CancelOnFirstError(
		ctx,
		func(ctx context.Context) error {
			defer close(ch)

			highWaterMark, err := libkafka.HighWaterMark(ctx, c.saramaClient, topic, partition)
			if err != nil {
				return errors.Wrapf(ctx, err, "get highwater marks failed")
			}

			if offset < 0 {
				newOffset := offset + *highWaterMark
				glog.V(2).Infof("offset(%d) < 0 => use %d", offset, newOffset)
				offset = newOffset
			}

			trigger := run.NewTrigger()

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			var counter uint64

			go func() {
				select {
				case <-ctx.Done():
				case <-trigger.Done():
					glog.V(2).Infof("read finished with %d records", counter)
					cancel()
				}
			}()

			return libkafka.NewSimpleConsumer(
				c.saramaClient,
				topic,
				offset,
				libkafka.MessageHanderList{
					libkafka.MessageHandlerFunc(func(ctx context.Context, msg *sarama.ConsumerMessage) error {
						record, err := c.converter.Convert(ctx, msg)
						if err != nil {
							return errors.Wrap(ctx, err, "convert msg to record failed")
						}
						select {
						case <-ctx.Done():
							return ctx.Err()
						case ch <- *record:
							counter++
							if counter == limit {
								trigger.Fire()
								// wait until trigger ctx is canceld
								<-ctx.Done()
							}
							return nil
						}
					}),
					libkafka.NewOffsetTriggerMessageHandler(
						map[libkafka.Partition]libkafka.Offset{partition: *highWaterMark},
						topic,
						trigger,
					),
				},
				c.logSamplerFactory,
			).Consume(ctx)
		},
		func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case record, ok := <-ch:
					if !ok {
						return nil
					}
					records = append(records, record)
				}
			}
		},
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return records, nil
		}
		return nil, errors.Wrap(ctx, err, "run failed")
	}
	return records, nil
}
