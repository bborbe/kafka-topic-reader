// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/bborbe/errors"
	libhttp "github.com/bborbe/http"
	libkafka "github.com/bborbe/kafka"
	"github.com/golang/glog"
)

type Page struct {
	NextOffset *libkafka.Offset `json:"nextOffset,omitempty"`
	Records    Records          `json:"records"`
}

type requestParams struct {
	topic     libkafka.Topic
	partition libkafka.Partition
	offset    libkafka.Offset
	limit     uint64
	filter    []byte
}

func parseRequestParams(ctx context.Context, req *http.Request) (*requestParams, error) {
	topic := libkafka.Topic(req.FormValue("topic"))
	if topic == "" {
		return nil, errors.New(ctx, "parameter topic missing")
	}

	offset, err := libkafka.ParseOffset(ctx, req.FormValue("offset"))
	if err != nil {
		return nil, errors.Wrap(ctx, err, "parse parameter offset failed")
	}

	limit, err := strconv.ParseUint(req.FormValue("limit"), 10, 64)
	if err != nil {
		limit = 100
	}

	partition, err := libkafka.ParsePartition(ctx, req.FormValue("partition"))
	if err != nil {
		return nil, errors.Wrap(ctx, err, "parse parameter partition failed")
	}

	filterValue := req.FormValue("filter")
	if len(filterValue) > 1024 {
		return nil, errors.New(ctx, "filter parameter exceeds maximum length of 1024 bytes")
	}

	return &requestParams{
		topic:     topic,
		partition: *partition,
		offset:    *offset,
		limit:     limit,
		filter:    []byte(filterValue),
	}, nil
}

func fetchChangesWithRetry(
	ctx context.Context,
	changesProvider ChangesProvider,
	params *requestParams,
) (Records, error) {
	changes, err := changesProvider.Changes(
		ctx,
		params.topic,
		params.partition,
		params.offset,
		params.limit,
		params.filter,
	)
	if err != nil {
		if !errors.Is(err, sarama.ErrOffsetOutOfRange) {
			return nil, errors.Wrap(ctx, err, "get changes failed")
		}
		glog.V(2).Infof("offset out of range error => fallbacktest to oldest")
		changes, err = changesProvider.Changes(
			ctx,
			params.topic,
			params.partition,
			libkafka.OffsetOldest,
			params.limit,
			params.filter,
		)
		if err != nil {
			return nil, errors.Wrap(ctx, err, "get changes failed")
		}
	}
	return changes, nil
}

func buildPage(changes Records, offset libkafka.Offset) Page {
	nextOffset := offset
	if len(changes) > 0 {
		nextOffset = changes[len(changes)-1].Offset + 1
	}
	return Page{
		Records:    changes,
		NextOffset: &nextOffset,
	}
}

func NewHandler(
	changesProvider ChangesProvider,
) libhttp.WithError {
	return libhttp.WithErrorFunc(
		func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
			ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			params, err := parseRequestParams(ctx, req)
			if err != nil {
				return err
			}

			glog.V(2).Infof(
				"read records from topic %s and partition %d and offset %d with limit %d started",
				params.topic, params.partition.Int32(), params.offset.Int64(), params.limit,
			)

			changes, err := fetchChangesWithRetry(ctx, changesProvider, params)
			if err != nil {
				return err
			}

			page := buildPage(changes, params.offset)

			if err := libhttp.SendJSONResponse(ctx, resp, page, http.StatusOK); err != nil {
				return errors.Wrap(ctx, err, "send json failed")
			}

			glog.V(2).Infof(
				"read %d records from topic %s and partition %d and offset %d with limit %d completed",
				len(
					page.Records,
				), params.topic, params.partition.Int32(), params.offset.Int64(), params.limit,
			)
			return nil
		},
	)
}
