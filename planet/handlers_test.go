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
	"github.com/venicegeo/geojson-go/geojson"
	pzsvc "github.com/venicegeo/pzsvc-lib"
)

const fakeDiscoverURL = "foo://bar/planet/discover?PL_API_KEY=%v"
const fakeActivateURL = "foo://bar/planet/activate/%v?PL_API_KEY=%v"

func TestHandlers(t *testing.T) {
	var (
		err     error
		request *http.Request
		fci     interface{}
	)
	// Test: No API Key
	if request, err = http.NewRequest("GET", fmt.Sprintf(fakeDiscoverURL, ""), nil); err != nil {
		t.Error(err.Error())
	}
	writer, _, _ := pzsvc.GetMockResponseWriter()
	DiscoverPlanetHandler(writer, request)
	if writer.StatusCode == http.StatusOK {
		t.Errorf("Expected request to fail due to lack of API Key but received: %v, %v", writer.StatusCode, writer.OutputString)
	}

	// Test: Discover (Happy)
	if request, err = http.NewRequest("GET", fmt.Sprintf(fakeDiscoverURL, os.Getenv("PL_API_KEY")), nil); err != nil {
		t.Error(err.Error())
	}
	writer, _, _ = pzsvc.GetMockResponseWriter()
	DiscoverPlanetHandler(writer, request)
	if writer.StatusCode != http.StatusOK {
		t.Errorf("Expected request to succeed but received: %v, %v", writer.StatusCode, writer.OutputString)
	}

	if fci, err = geojson.Parse([]byte(writer.OutputString)); err != nil {
		t.Fatalf("Expected to parse GeoJSON but received: %v", err.Error())
	}
	id := fci.(*geojson.FeatureCollection).Features[0].IDStr()

	// Test: Activate, no Image ID
	if request, err = http.NewRequest("GET", fmt.Sprintf(fakeActivateURL, "", ""), nil); err != nil {
		t.Error(err.Error())
	}
	writer, _, _ = pzsvc.GetMockResponseWriter()
	router := mux.NewRouter()

	router.HandleFunc("/planet/activate/{id}", ActivatePlanetHandler)
	router.ServeHTTP(writer, request)
	if writer.StatusCode == http.StatusOK {
		t.Errorf("Expected request to fail due to lack of Image ID but received: %v, %v", writer.StatusCode, writer.OutputString)
	}

	// Test: Activate, no API Key
	if request, err = http.NewRequest("GET", fmt.Sprintf(fakeActivateURL, id, ""), nil); err != nil {
		t.Error(err.Error())
	}
	writer, _, _ = pzsvc.GetMockResponseWriter()
	router.ServeHTTP(writer, request)
	if writer.StatusCode == http.StatusOK {
		t.Errorf("Expected request to fail due to lack of API Key but received: %v, %v", writer.StatusCode, writer.OutputString)
	}

	// Test: Activate (happy)
	if request, err = http.NewRequest("GET", fmt.Sprintf(fakeActivateURL, id, os.Getenv("PL_API_KEY")), nil); err != nil {
		t.Error(err.Error())
	}
	writer, _, _ = pzsvc.GetMockResponseWriter()
	router.ServeHTTP(writer, request)
	if writer.StatusCode != http.StatusOK {
		t.Errorf("Expected request to succeed but received: %v, %v", writer.StatusCode, writer.OutputString)
	}

	fmt.Print(writer.OutputString)
}
