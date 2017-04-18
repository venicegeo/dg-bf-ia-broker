// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tides

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/venicegeo/geojson-go/geojson"
)

func TestGetTides(t *testing.T) {
	fc, err := getTestingFeatureCollection()
	server := CreateMockTidesServer()
	context := Context{TidesURL: server.URL}

	if err != nil {
		t.Fatalf("Failed loading testing feature collection %v", err)
	}

	fc, err = GetTides(fc, &context)
	assert.Nil(t, err, "Expected GetTides to succeed but received: %v", err)

	_, err = geojson.Write(fc)
	assert.Nil(t, err, "Failed to export output from GeoJSON: %v\n%#v", err)
}
