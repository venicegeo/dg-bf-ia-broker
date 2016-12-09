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
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/venicegeo/geojson-go/geojson"
	"github.com/venicegeo/pzsvc-lib"
)

const noPlanetKey = "This operation requires a Planet Labs API key."

// DiscoverPlanetHandler is a handler for /discoverPlanet
// @Title discoverPlanet
// @Description discovers scenes from Planet Labs
// @Accept  plain
// @Param   PL_API_KEY      query   string  true         "Planet Labs API Key"
// @Param   bbox            query   string  false        "The bounding box, as a GeoJSON Bounding box (x1,y1,x2,y2)"
// @Param   acquiredDate    query   string  false        "The minimum (earliest) acquired date, as RFC 3339"
// @Param   maxAcquiredDate query   string  false        "The maximum acquired date, as RFC 3339"
// @Param   tides           query   bool    false        "True: incorporate tide prediction in the output"
// @Success 200 {object}  geojson.FeatureCollection
// @Failure 400 {object}  string
// @Router /planet/discover [get]
func DiscoverPlanetHandler(writer http.ResponseWriter, request *http.Request) {
	var (
		responseString string
		err            error
		planetKey      string
		bbox           geojson.BoundingBox
	)
	if pzsvc.Preflight(writer, request) {
		return
	}

	tides, _ := strconv.ParseBool(request.FormValue("tides"))
	planetKey = request.FormValue("PL_API_KEY")
	if planetKey == "" {
		http.Error(writer, noPlanetKey, http.StatusBadRequest)
		return
	}

	bboxString := request.FormValue("bbox")
	if bboxString != "" {
		if bbox, err = geojson.NewBoundingBox(bboxString); err != nil {
			http.Error(writer, "The bbox value of "+bboxString+" is invalid: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	planetKey = request.FormValue("PL_API_KEY")
	options := SearchOptions{
		Bbox: bbox}
	context := Context{
		Tides:     tides,
		PlanetKey: planetKey}

	if responseString, err = GetScenes(options, context); err == nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.Write([]byte(responseString))
	} else {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

// ActivatePlanetHandler is a handler for /activatePlanet
// @Title activatePlanet
// @Description activates scenes from Planet Labs
// @Accept  plain
// @Param   PL_API_KEY      query   string  true         "Planet Labs API Key"
// @Param   id              path    string  true         "Planet Labs image ID"
// @Success 200 {object}  string
// @Failure 400 {object}  string
// @Router /planet/discover [get]
func ActivatePlanetHandler(writer http.ResponseWriter, r *http.Request) {
	var (
		err     error
		context Context
		result  []byte
	)
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(writer, "This operation requires a Planet Labs image ID.", http.StatusBadRequest)
		return
	}
	context.PlanetKey = r.FormValue("PL_API_KEY")

	if context.PlanetKey == "" {
		http.Error(writer, "This operation requires a Planet Labs API key.", http.StatusBadRequest)
		return
	}

	if result, err = Activate(id, context); err == nil {
		writer.Write(result)
	} else {
		http.Error(writer, "Failed to acquire activation information for "+id+": "+err.Error(), http.StatusBadRequest)
	}
}
