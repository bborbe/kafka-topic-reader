// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import (
	"encoding/json"
	"net/http"
)

const ContentTypeApplicationJSON = "application/json"

const ContentTypeField = "Content-Type"

func SendJSONResponse(resp http.ResponseWriter, data interface{}, statusCode int) error {
	resp.Header().Add(ContentTypeField, ContentTypeApplicationJSON)
	resp.WriteHeader(statusCode)
	return json.NewEncoder(resp).Encode(data)
}
