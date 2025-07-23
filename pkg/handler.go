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

func NewHandler(
	changesProvider ChangesProvider,
) libhttp.WithError {
	return libhttp.WithErrorFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
		ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		topic := libkafka.Topic(req.FormValue("topic"))
		if topic == "" {
			return errors.New(ctx, "parameter topic missing")
		}

		offset, err := libkafka.ParseOffset(ctx, req.FormValue("offset"))
		if err != nil {
			return errors.Wrap(ctx, err, "parse parameter offset failed")
		}
		limit, err := strconv.ParseUint(req.FormValue("limit"), 10, 64)
		if err != nil {
			limit = 100
		}
		partition, err := libkafka.ParsePartition(ctx, req.FormValue("partition"))
		if err != nil {
			return errors.Wrap(ctx, err, "parse parameter partition failed")
		}

		filterValue := req.FormValue("filter")
		if len(filterValue) > 1024 {
			return errors.New(ctx, "filter parameter exceeds maximum length of 1024 bytes")
		}
		filter := []byte(filterValue)

		glog.V(2).Infof("read records from topic %s and partition %d and offset %d with limit %d started", topic, partition.Int32(), offset.Int64(), limit)
		changes, err := changesProvider.Changes(ctx, topic, *partition, *offset, limit, filter)
		if err != nil {
			if errors.Is(err, sarama.ErrOffsetOutOfRange) == false {
				return errors.Wrap(ctx, err, "get changes failed")
			}
			glog.V(2).Infof("offset out of range error => fallbacktest to oldest")
			changes, err = changesProvider.Changes(ctx, topic, *partition, libkafka.OffsetOldest, limit, filter)
			if err != nil {
				return errors.Wrap(ctx, err, "get changes failed")
			}
		}
		nextOffset := *offset
		if len(changes) > 0 {
			nextOffset = changes[len(changes)-1].Offset + 1
		}
		page := Page{
			Records:    changes,
			NextOffset: &nextOffset,
		}
		if err := SendJSONResponse(resp, page, http.StatusOK); err != nil {
			return errors.Wrap(ctx, err, "send json failed")
		}
		glog.V(2).Infof("read %d records from topic %s and partition %d and offset %d with limit %d completed", len(page.Records), topic, partition.Int32(), offset.Int64(), limit)
		return nil
	})
}
