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

package planet

import (
	"fmt"
	"os"
	"testing"

	"github.com/venicegeo/geojson-go/geojson"
)

func TestPlanet(t *testing.T) {
	var (
		options SearchOptions
		err     error
		// response string
		context Context
	)

	context.PlanetKey = os.Getenv("PL_API_KEY")
	options.ItemType = "REOrthoTile"

	body := `{ 
   "item_types":[  
      "REOrthoTile"
   ],
   "filter":{  
      "type":"AndFilter",
      "config":[
        ]
      }
}`

	// Test 1 - No parameters
	if _, err = doRequest(doRequestInput{method: "POST", inputURL: "data/v1/quick-search", body: []byte(body), contentType: "application/json"}, context); err != nil {
		t.Errorf("Expected request to succeed; received: %v", err.Error())
	}

	// Test 2 - Bbox
	if options.Bbox, err = geojson.NewBoundingBox("139,50,140,51"); err != nil {
		t.Errorf("Expected NewBoundingBox to succeed; received: %v", err.Error())
	}
	if _, err = GetScenes(options, context); err != nil {
		t.Errorf("Expected GetScenes to succeed; received: %v", err.Error())
	}

	// Test 3 - Cloud Cover
	options.CloudCover = 0.01
	if _, err = GetScenes(options, context); err != nil {
		t.Errorf("Expected GetScenes to succeed; received: %v", err.Error())
	}

	// Test 4 - Acquired Date
	options.AcquiredDate = "2016-01-01T00:00:00Z"
	if _, err = GetScenes(options, context); err != nil {
		t.Errorf("Expected GetScenes to succeed; received: %v", err.Error())
	}

	// Test 5 - Tides
	options.Tides = true
	var scenes *geojson.FeatureCollection
	if scenes, err = GetScenes(options, context); err != nil {
		t.Errorf("Expected GetScenes to succeed; received: %v", err.Error())
	}

	// Test - Metadata
	var feature *geojson.Feature
	aOptions := AssetOptions{ID: scenes.Features[0].IDStr(), activate: true, ItemType: "REOrthoTile"}
	if feature, err = GetMetadata(aOptions, context); err != nil {
		t.Errorf("Failed to get asset; received: %v", err.Error())
	}
	b, _ := geojson.Write(feature)
	fmt.Print(string(b))

	// Test - Activation
	if _, err = GetAsset(aOptions, context); err != nil {
		t.Errorf("Failed to get asset; received: %v", err.Error())
	}
}
