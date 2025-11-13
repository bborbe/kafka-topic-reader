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
		filter []byte,
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
	filter []byte,
) (Records, error) {
	var records Records
	ch := make(chan Record, runtime.NumCPU())
	err := run.CancelOnFirstError(
		ctx,
		c.produceRecords(ch, topic, partition, offset, limit, filter),
		c.collectRecords(ch, &records),
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return records, nil
		}
		return nil, errors.Wrap(ctx, err, "run failed")
	}
	return records, nil
}

func (c *changesProvider) produceRecords(
	ch chan Record,
	topic libkafka.Topic,
	partition libkafka.Partition,
	offset libkafka.Offset,
	limit uint64,
	filter []byte,
) func(context.Context) error {
	return func(ctx context.Context) error {
		defer close(ch)

		highWaterMark, err := libkafka.HighWaterMark(ctx, c.saramaClient, topic, partition)
		if err != nil {
			return errors.Wrapf(ctx, err, "get highwater marks failed")
		}

		offset = c.adjustNegativeOffset(offset, highWaterMark)

		trigger := run.NewTrigger()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var counter uint64
		c.startTriggerWatcher(ctx, trigger, cancel, &counter)

		return libkafka.NewSimpleConsumer(
			c.saramaClient,
			topic,
			offset,
			libkafka.MessageHanderList{
				c.createMessageHandler(ch, filter, &counter, limit, trigger),
				libkafka.NewOffsetTriggerMessageHandler(
					map[libkafka.Partition]libkafka.Offset{partition: *highWaterMark},
					topic,
					trigger,
				),
			},
			c.logSamplerFactory,
		).Consume(ctx)
	}
}

func (c *changesProvider) adjustNegativeOffset(
	offset libkafka.Offset,
	highWaterMark *libkafka.Offset,
) libkafka.Offset {
	if offset < 0 {
		newOffset := offset + *highWaterMark
		glog.V(2).Infof("offset(%d) < 0 => use %d", offset, newOffset)
		return newOffset
	}
	return offset
}

func (c *changesProvider) startTriggerWatcher(
	ctx context.Context,
	trigger run.Trigger,
	cancel context.CancelFunc,
	counter *uint64,
) {
	go func() {
		select {
		case <-ctx.Done():
		case <-trigger.Done():
			glog.V(2).Infof("read finished with %d records", *counter)
			cancel()
		}
	}()
}

func (c *changesProvider) createMessageHandler(
	ch chan Record,
	filter []byte,
	counter *uint64,
	limit uint64,
	trigger run.Trigger,
) libkafka.MessageHandler {
	return libkafka.MessageHandlerFunc(
		func(ctx context.Context, msg *sarama.ConsumerMessage) error {
			if !MatchesFilter(msg, filter) {
				return nil
			}

			record, err := c.converter.Convert(ctx, msg)
			if err != nil {
				return errors.Wrap(ctx, err, "convert msg to record failed")
			}

			return c.sendRecordOrCancel(ctx, ch, record, counter, limit, trigger)
		},
	)
}

func (c *changesProvider) sendRecordOrCancel(
	ctx context.Context,
	ch chan Record,
	record *Record,
	counter *uint64,
	limit uint64,
	trigger run.Trigger,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case ch <- *record:
		*counter++
		if *counter == limit {
			trigger.Fire()
			<-ctx.Done()
		}
		return nil
	}
}

func (c *changesProvider) collectRecords(
	ch chan Record,
	records *Records,
) func(context.Context) error {
	return func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case record, ok := <-ch:
				if !ok {
					return nil
				}
				*records = append(*records, record)
			}
		}
	}
}
