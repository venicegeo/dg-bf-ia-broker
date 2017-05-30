// Copyright 2017, RadiantBlue Technologies, Inc.
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

package planet

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/venicegeo/dg-bf-ia-broker/util"
	"github.com/venicegeo/dg-geojson-go/geojson"
)

func TestBestScene(t *testing.T) {
	planetServer, tidesServer, _ := createTestFixtures()
	context := makeTestingContext(planetServer, tidesServer)
	options := SearchOptions{ItemType: "REOrthoTile"}

	coordinates := []float64{105.0, 8.5}
	point := geojson.NewPoint(coordinates)
	options.Bbox = point.ForceBbox()

	best, err := BestScene(options, &context)
	assert.Nil(t, err, "Retrieving best scene failed with %v", err)
	if err == nil {
		util.LogInfo(&context, fmt.Sprintf("Found best scene: %v", best))
	}

	assert.NotEmpty(t, best, "Expected non-empty best scene ID, got empty string")
}
