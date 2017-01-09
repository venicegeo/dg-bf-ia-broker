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
	"os"
	"testing"

	"github.com/venicegeo/bf-ia-broker/util"
	"github.com/venicegeo/geojson-go/geojson"
)

func TestBatch(t *testing.T) {
	var (
		options SearchOptions
		context Context
	)

	context.PlanetKey = os.Getenv("PL_API_KEY")
	options.ItemType = "REOrthoTile"

	coordinates := []float64{8.5, 105.0}
	point := geojson.NewPoint(coordinates)
	options.Bbox = point.ForceBbox()
	// util.LogInfo(&context, fmt.Sprintf("%#v", options))

	if best, err := BestScene(options, &context); err == nil {
		util.LogInfo(&context, fmt.Sprintf("Found best scene: %v", best))
	} else {
		t.Error(err)
	}
}
