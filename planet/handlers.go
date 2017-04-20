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
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/venicegeo/bf-ia-broker/util"
	"github.com/venicegeo/geojson-go/geojson"
)

const noPlanetKey = "This operation requires a Planet Labs API key."
const noPlanetImageID = "This operation requires a Planet Labs image ID."
const invalidCloudCover = "Cloud Cover value of %v is invalid."

// DiscoverHandler is a handler for /planet/discover
// @Title planetDiscoverHandler
// @Description discovers scenes from Planet Labs
// @Accept  plain
// @Param   PL_API_KEY      query   string  true         "Planet Labs API Key"
// @Param   itemType        path    string  true         "Planet Labs Item Type, e.g., rapideye or planetscope"
// @Param   bbox            query   string  false        "The bounding box, as a GeoJSON Bounding box (x1,y1,x2,y2)"
// @Param   cloudCover      query   string  false        "The maximum cloud cover, as a percentage (0-100)"
// @Param   acquiredDate    query   string  false        "The minimum (earliest) acquired date, as RFC 3339"
// @Param   maxAcquiredDate query   string  false        "The maximum acquired date, as RFC 3339"
// @Param   tides           query   bool    false        "True: incorporate tide prediction in the output"
// @Success 200 {object}  geojson.FeatureCollection
// @Failure 400 {object}  string
// @Router /planet/discover/{itemType} [get]
func DiscoverHandler(writer http.ResponseWriter, request *http.Request) {
	var (
		fc         *geojson.FeatureCollection
		err        error
		itemType   string
		bytes      []byte
		bbox       geojson.BoundingBox
		ccStr      string
		cloudCover float64
		context    Context
	)
	util.LogAudit(&context, util.LogAuditInput{Actor: "anon user", Action: request.Method, Actee: request.URL.String(), Message: "Receiving /discover request", Severity: util.INFO})

	if util.Preflight(writer, request, &context) {
		return
	}

	context.PlanetKey = request.FormValue("PL_API_KEY")
	if context.PlanetKey == "" {
		util.LogSimpleErr(&context, noPlanetKey, nil)
		util.HTTPError(request, writer, &context, noPlanetKey, http.StatusBadRequest)
		return
	}

	tides, _ := strconv.ParseBool(request.FormValue("tides"))

	ccStr = request.FormValue("cloudCover")
	if ccStr != "" {
		if cloudCover, err = strconv.ParseFloat(ccStr, 64); err != nil {
			message := fmt.Sprintf(invalidCloudCover, ccStr)
			util.LogSimpleErr(&context, message, err)
			util.HTTPError(request, writer, &context, message, http.StatusBadRequest)
			return
		}
		cloudCover = cloudCover / 100.0
	}

	itemType = mux.Vars(request)["itemType"]
	switch itemType {
	case "REOrthoTile", "rapideye":
		itemType = "REOrthoTile"
	case "PSOrthoTile", "planetscope":
		itemType = "PSOrthoTile"
	case "PSScene4Band":
		// No op
	default:
		message := fmt.Sprintf("The item type value of %v is invalid", itemType)
		util.LogSimpleErr(&context, message, nil)
		util.HTTPError(request, writer, &context, message, http.StatusBadRequest)
		return
	}

	bboxString := request.FormValue("bbox")
	if bboxString != "" {
		if bbox, err = geojson.NewBoundingBox(bboxString); err != nil {
			message := fmt.Sprintf("The bbox value of %v is invalid", bboxString)
			util.LogSimpleErr(&context, message, err)
			util.HTTPError(request, writer, &context, message, http.StatusBadRequest)
			return
		}
	}

	options := SearchOptions{
		ItemType:        itemType,
		CloudCover:      cloudCover,
		AcquiredDate:    request.FormValue("acquiredDate"),
		MaxAcquiredDate: request.FormValue("maxAcquiredDate"),
		Tides:           tides,
		Bbox:            bbox}

	if fc, err = GetScenes(options, &context); err == nil {
		if bytes, err = geojson.Write(fc); err != nil {
			err = util.LogSimpleErr(&context, fmt.Sprintf("Failed to write output GeoJSON from:\n%#v", fc), err)
			util.HTTPError(request, writer, &context, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(bytes)
		util.LogAudit(&context, util.LogAuditInput{Actor: "anon user", Action: request.Method + " response", Actee: request.URL.String(), Message: "Sending /discover response", Severity: util.INFO})
	} else {
		switch herr := err.(type) {
		case util.HTTPErr:
			util.HTTPError(request, writer, &context, herr.Message, herr.Status)
		default:
			err = util.LogSimpleErr(&context, "Failed to get Planet Labs scenes. ", err)
			util.HTTPError(request, writer, &context, err.Error(), http.StatusInternalServerError)
		}
	}
}

// MetadataHandler is a handler for /planet
// @Title planetMetadataHandler
// @Description Gets image metadata from Planet Labs
// @Accept  plain
// @Param   PL_API_KEY      query   string  true         "Planet Labs API Key"
// @Param   itemType        path    string  true         "Planet Labs Item Type, e.g., rapideye or planetscope"
// @Param   id              path    string  true         "Planet Labs image ID"
// @Param   tides           query   bool    false        "True: incorporate tide prediction in the output"
// @Success 200 {object}  geojson.Feature
// @Failure 400 {object}  string
// @Router /planet/{itemType}/{id} [get]
func MetadataHandler(writer http.ResponseWriter, request *http.Request) {
	var (
		err     error
		context Context
		feature *geojson.Feature
		bytes   []byte
		options MetadataOptions
		asset   Asset
	)

	util.LogAudit(&context, util.LogAuditInput{Actor: "anon user", Action: request.Method, Actee: request.URL.String(), Message: "Receiving /planet/{itemType}/{id} request", Severity: util.INFO})

	if util.Preflight(writer, request, &context) {
		return
	}
	vars := mux.Vars(request)
	options.ID = vars["id"]
	if options.ID == "" {

		util.LogSimpleErr(&context, noPlanetImageID, nil)
		util.HTTPError(request, writer, &context, noPlanetImageID, http.StatusBadRequest)
		return
	}
	context.PlanetKey = request.FormValue("PL_API_KEY")

	if context.PlanetKey == "" {
		util.LogAlert(&context, noPlanetKey)
		util.HTTPError(request, writer, &context, noPlanetKey, http.StatusBadRequest)
		return
	}

	options.Tides, _ = strconv.ParseBool(request.FormValue("tides"))

	itemType := vars["itemType"]
	switch itemType {
	case "REOrthoTile", "rapideye":
		options.ItemType = "REOrthoTile"
	case "PSOrthoTile", "planetscope":
		options.ItemType = "PSOrthoTile"
	case "PSScene4Band":
		// No op
	default:
		message := fmt.Sprintf("The item type value of %v is invalid", itemType)
		util.LogSimpleErr(&context, message, nil)
		util.HTTPError(request, writer, &context, message, http.StatusBadRequest)
		return
	}

	if feature, err = GetMetadata(options, &context); err == nil {
		if asset, err = GetAsset(options, &context); err == nil {
			injectAssetIntoMetadata(feature, asset)
			if bytes, err = geojson.Write(feature); err != nil {
				err = util.LogSimpleErr(&context, fmt.Sprintf("Failed to write output GeoJSON from:\n%#v", feature), err)
				util.HTTPError(request, writer, &context, err.Error(), http.StatusInternalServerError)
				return
			}
			writer.Header().Set("Content-Type", "application/json")
			writer.Write(bytes)

			util.LogInfo(&context, "Asset: "+string(bytes))
			util.LogAudit(&context, util.LogAuditInput{Actor: request.URL.String(), Action: request.Method + " response" + " response", Actee: "anon user", Message: "Sending planet/{itemType}/{id} response", Severity: util.INFO})
		} else {
			switch herr := err.(type) {
			case util.HTTPErr:
				util.HTTPError(request, writer, &context, herr.Message, herr.Status)
			default:
				err = util.LogSimpleErr(&context, "Failed to get Planet Labs asset information. ", err)
				util.HTTPError(request, writer, &context, err.Error(), 0)
			}
		}
	} else {
		switch herr := err.(type) {
		case util.HTTPErr:
			util.HTTPError(request, writer, &context, herr.Message, herr.Status)
		default:
			err = util.LogSimpleErr(&context, "Failed to get Planet Labs scene metadata. ", err)
			util.HTTPError(request, writer, &context, err.Error(), 0)
		}
	}
}

// ActivateHandler is a handler for /planet
// @Title planetActivateHandler
// @Description Activates a scene
// @Accept  plain
// @Param   PL_API_KEY      query   string  true         "Planet Labs API Key"
// @Param   itemType        path    string  true         "Planet Labs Item Type, e.g., rapideye or planetscope"
// @Param   id              path    string  true         "Planet Labs image ID"
// @Success 200 {object}  geojson.Feature
// @Failure 400 {object}  string
// @Router /planet/activate/{itemType}/{id} [post]
func ActivateHandler(writer http.ResponseWriter, request *http.Request) {
	var (
		err      error
		context  Context
		options  MetadataOptions
		response *http.Response
	)

	util.LogAudit(&context, util.LogAuditInput{Actor: "anon user", Action: request.Method, Actee: request.URL.String(), Message: "Receiving /planet/activate/{itemType}/{id} request", Severity: util.INFO})

	if util.Preflight(writer, request, &context) {
		return
	}
	vars := mux.Vars(request)
	options.ID = vars["id"]
	if options.ID == "" {
		util.LogSimpleErr(&context, noPlanetImageID, nil)
		util.HTTPError(request, writer, &context, noPlanetImageID, http.StatusBadRequest)
		return
	}
	context.PlanetKey = request.FormValue("PL_API_KEY")

	if context.PlanetKey == "" {

		util.LogAlert(&context, noPlanetKey)
		util.HTTPError(request, writer, &context, noPlanetKey, http.StatusBadRequest)
		return
	}

	itemType := vars["itemType"]
	switch itemType {
	case "REOrthoTile", "rapideye":
		options.ItemType = "REOrthoTile"
	case "PSOrthoTile", "planetscope":
		options.ItemType = "PSOrthoTile"
	case "PSScene4Band":
		// No op
	default:
		message := fmt.Sprintf("The item type value of %v is invalid", itemType)
		util.LogSimpleErr(&context, message, nil)
		util.HTTPError(request, writer, &context, message, http.StatusBadRequest)
		return
	}

	if response, err = Activate(options, &context); err == nil {
		defer response.Body.Close()
		writer.Header().Set("Content-Type", response.Header.Get("Content-Type"))
		if (response.StatusCode >= 200) && (response.StatusCode < 300) {
			bytes, _ := ioutil.ReadAll(response.Body)
			writer.Write(bytes)
			util.LogAudit(&context, util.LogAuditInput{Actor: request.URL.String(), Action: request.Method + " response", Actee: "anon user", Message: "Sending planet/{itemType}/{id} response", Severity: util.INFO})
		} else {
			message := fmt.Sprintf("Failed to activate Planet Labs scene: " + response.Status)
			err = util.LogSimpleErr(&context, message, nil)
			util.HTTPError(request, writer, &context, err.Error(), response.StatusCode)
		}
	} else {
		switch herr := err.(type) {
		case util.HTTPErr:
			util.HTTPError(request, writer, &context, herr.Message, herr.Status)
		default:
			err = util.LogSimpleErr(&context, "Failed to activate Planet Labs scene. ", err)
			util.HTTPError(request, writer, &context, err.Error(), 0)
		}
	}
}
