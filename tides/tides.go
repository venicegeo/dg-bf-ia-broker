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
	"fmt"
	"math"
	"os"
	"time"

	"github.com/venicegeo/bf-ia-broker/util"
	"github.com/venicegeo/geojson-go/geojson"
)

var tidesURL string

func init() {
	tidesURL = os.Getenv("BF_TIDE_PREDICTION_URL")
	if tidesURL == "" {
		util.LogAlert(&util.BasicLogContext{}, "Didn't get Tide Prediction URL from the environment. Using default.")
		tidesURL = "https://bf-tideprediction.int.geointservices.io/tides"
	}
}

// Context is the context for this operation
type Context struct {
	TidesURL  string
	sessionID string
}

// AppName returns an empty string
func (c *Context) AppName() string {
	return "bf-ia-broker"
}

// SessionID returns a Session ID, creating one if needed
func (c *Context) SessionID() string {
	if c.sessionID == "" {
		c.sessionID, _ = util.PsuUUID()
	}
	return c.sessionID
}

// LogRootDir returns an empty string
func (c *Context) LogRootDir() string {
	return ""
}

type tideIn struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
	Dtg string  `json:"dtg"`
}

type tidesIn struct {
	Locations []tideIn                    `json:"locations"`
	Map       map[string]*geojson.Feature `json:"-"`
}

type tideOut struct {
	MinTide  float64 `json:"minimumTide24Hours"`
	MaxTide  float64 `json:"maximumTide24Hours"`
	CurrTide float64 `json:"currentTide"`
}

type tideWrapper struct {
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Dtg     string  `json:"dtg"`
	Results tideOut `json:"results"`
}

type out struct {
	Locations []tideWrapper `json:"locations"`
}

func toTideIn(bbox geojson.BoundingBox, timeStr string) *tideIn {
	var (
		center  *geojson.Point
		dtgTime time.Time
		err     error
	)
	if center = bbox.Centroid(); center == nil {
		return nil
	}
	if dtgTime, err = time.Parse("2006-01-02T15:04:05Z", timeStr); err != nil {
		return nil
	}
	return &tideIn{Lat: center.Coordinates[1], Lon: center.Coordinates[0], Dtg: dtgTime.Format("2006-01-02-15-04")}
}

func toTidesIn(features []*geojson.Feature, context util.LogContext) *tidesIn {
	var (
		result     tidesIn
		currTideIn *tideIn
	)
	result.Map = make(map[string]*geojson.Feature)
	for _, feature := range features {
		if feature.PropertyFloat("CurrentTide") != math.NaN() {
			if currTideIn = toTideIn(feature.ForceBbox(), feature.PropertyString("acquiredDate")); currTideIn == nil {
				util.LogInfo(context, fmt.Sprintf("Could not get tide information from feature %v because required elements did not exist. BBOX: %#v, Date: %v",
					feature.IDStr(),
					feature.ForceBbox(),
					feature.PropertyString("acquiredDate")))
				continue
			}
			result.Locations = append(result.Locations, *currTideIn)
			result.Map[currTideIn.Dtg] = feature
		}
	}
	switch len(result.Locations) {
	case 0:
		return nil
	default:
		return &result
	}
}

// GetTides returns the tide information for the features provided.
// Features must have a geometry and an acquiredDate property.
func GetTides(fc *geojson.FeatureCollection, context *Context) (*geojson.FeatureCollection, error) {
	var (
		err          error
		tin          *tidesIn
		tout         out
		currentScene *geojson.Feature
		result       *geojson.FeatureCollection
		ok           bool
	)
	tidesURL := context.TidesURL
	tin = toTidesIn(fc.Features, context)
	features := make([]*geojson.Feature, len(fc.Features))
	util.LogAudit(context, util.LogAuditInput{Actor: "anon user", Action: "POST", Actee: tidesURL, Message: "Requesting tide information", Severity: util.INFO})
	if _, err = util.ReqByObjJSON("POST", tidesURL, "", tin, &tout); err == nil {
		util.LogAudit(context, util.LogAuditInput{Actor: tidesURL, Action: "POST response", Actee: "anon user", Message: "Retrieving tide information", Severity: util.INFO})
		for inx, tideObj := range tout.Locations {
			if currentScene, ok = tin.Map[tideObj.Dtg]; !ok {
				util.LogInfo(context, "Failed to find location for "+tideObj.Dtg)
				continue
			}
			currentScene.Properties["CurrentTide"] = tideObj.Results.CurrTide
			currentScene.Properties["MinimumTide24Hours"] = tideObj.Results.MinTide
			currentScene.Properties["MaximumTide24Hours"] = tideObj.Results.MaxTide
			features[inx] = currentScene
		}
		result = geojson.NewFeatureCollection(features)
	} else {
		err = util.LogSimpleErr(context, "Failed to retrieve tides", err)
	}
	return result, err
}
