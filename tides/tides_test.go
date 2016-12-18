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

	"github.com/venicegeo/geojson-go/geojson"
)

func TestTides(t *testing.T) {
	var (
		err     error
		fci     interface{}
		context Context
		fc      *geojson.FeatureCollection
		ok      bool
	)
	if fci, err = geojson.ParseFile("test/fc.geojson"); err != nil {
		t.Fatalf("Expected to load file but received: %v", err.Error())
	}
	if fc, ok = fci.(*geojson.FeatureCollection); !ok {
		t.Fatalf("Expected FeatureCollection but received %T", fci)
	}
	if fc, err = GetTides(fc, context); err != nil {
		t.Fatalf("Expected GetTides to succeed but received: %v", err.Error())
	}
	if _, err = geojson.Write(fc); err != nil {
		t.Errorf("Failed to export output from GeoJSON: %v\n%#v", err.Error(), fc)
	}
}
