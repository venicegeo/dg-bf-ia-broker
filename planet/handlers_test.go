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
	"net/http"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/venicegeo/bf-ia-broker/util"
	"github.com/venicegeo/geojson-go/geojson"
)

const fakeDiscoverURL = "foo://bar/planet/discover/rapideye?PL_API_KEY=%v"
const fakeMetadataURL = "foo://bar/planet/rapideye/%v?PL_API_KEY=%v"
const fakeActivateURL = "foo://bar/planet/activate/rapideye/%v?PL_API_KEY=%v"

func TestHandlers(t *testing.T) {
	var (
		err     error
		request *http.Request
		fci     interface{}
	)

	router := mux.NewRouter()
	router.HandleFunc("/planet/discover/{itemType}", DiscoverHandler)
	router.HandleFunc("/planet/activate/{itemType}/{id}", ActivateHandler)
	router.HandleFunc("/planet/{itemType}/{id}", MetadataHandler)

	// Test: No API Key
	if request, err = http.NewRequest("GET", fmt.Sprintf(fakeDiscoverURL, ""), nil); err != nil {
		t.Error(err.Error())
	}
	writer, _, _ := util.GetMockResponseWriter()
	router.ServeHTTP(writer, request)
	if writer.StatusCode == http.StatusOK {
		t.Errorf("Expected request to fail due to lack of API Key but received: %v, %v", writer.StatusCode, writer.OutputString)
	}

	// Test: Invalid API Key
	if request, err = http.NewRequest("GET", fmt.Sprintf(fakeDiscoverURL, "foo"), nil); err != nil {
		t.Error(err.Error())
	}
	writer, _, _ = util.GetMockResponseWriter()
	router.ServeHTTP(writer, request)
	if writer.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected request to fail due to unauthorized API Key but received: %v, %v", writer.StatusCode, writer.OutputString)
	}

	// Test: Discover (Happy)
	if request, err = http.NewRequest("GET", fmt.Sprintf(fakeDiscoverURL, os.Getenv("PL_API_KEY")), nil); err != nil {
		t.Error(err.Error())
	}
	writer, _, _ = util.GetMockResponseWriter()
	router.ServeHTTP(writer, request)

	if writer.StatusCode != http.StatusOK {
		t.Errorf("Expected request to succeed but received: %v, %v", writer.StatusCode, writer.OutputString)
	}

	if fci, err = geojson.Parse([]byte(writer.OutputString)); err != nil {
		t.Fatalf("Expected to parse GeoJSON but received: %v", err.Error())
	}
	id := fci.(*geojson.FeatureCollection).Features[0].IDStr()

	// Test: Activate, no Image ID
	// We can't currently run activate tests because some images we receive are not activatable
	// if request, err = http.NewRequest("GET", fmt.Sprintf(fakeAssetURL, "", ""), nil); err != nil {
	// 	t.Error(err.Error())
	// }
	// writer, _, _ = util.GetMockResponseWriter()

	// // Test: Activate, no API Key
	// if request, err = http.NewRequest("POST", fmt.Sprintf(fakeAssetURL, id, ""), nil); err != nil {
	// 	t.Error(err.Error())
	// }
	// writer, _, _ = util.GetMockResponseWriter()
	// router.ServeHTTP(writer, request)
	// if writer.StatusCode == http.StatusOK {
	// 	t.Errorf("Expected request to fail due to lack of API Key but received: %v, %v", writer.StatusCode, writer.OutputString)
	// }
	//
	// Test: Metadata (happy)
	metadataURL := fmt.Sprintf(fakeMetadataURL, id, os.Getenv("PL_API_KEY"))

	if request, err = http.NewRequest("GET", metadataURL, nil); err != nil {
		t.Error(err.Error())
	}
	writer, _, _ = util.GetMockResponseWriter()
	router.ServeHTTP(writer, request)
	if writer.StatusCode != http.StatusOK {
		t.Errorf("Expected request to succeed but received: %v, %v", writer.StatusCode, writer.OutputString)
	}

	// Test: Metadata (no image ID)
	metadataURL = fmt.Sprintf(fakeMetadataURL, "", os.Getenv("PL_API_KEY"))

	if request, err = http.NewRequest("GET", metadataURL, nil); err != nil {
		t.Error(err.Error())
	}
	writer, _, _ = util.GetMockResponseWriter()
	router.ServeHTTP(writer, request)
	if writer.StatusCode != http.StatusNotFound {
		t.Errorf("Expected request to return a 404 but it returned a %v.", writer.StatusCode)
	}

	// Test: Activate (happy)
	activateURL := fmt.Sprintf(fakeActivateURL, id, os.Getenv("PL_API_KEY"))
	if request, err = http.NewRequest("POST", activateURL, nil); err != nil {
		t.Error(err.Error())
	}
	writer, _, _ = util.GetMockResponseWriter()
	// Since this request will routinely fail, we do not check its status
	router.ServeHTTP(writer, request)
}
