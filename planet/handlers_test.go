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
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"strings"

	"encoding/base64"

	"github.com/gorilla/mux"
	"github.com/venicegeo/bf-ia-broker/util"
	"github.com/venicegeo/geojson-go/geojson"
)

const fakeDiscoverURL = "%s/planet/discover/rapideye?PL_API_KEY=%v"
const fakeMetadataURL = "%s/planet/rapideye/%v?PL_API_KEY=%v"
const fakeActivateURL = "%s/planet/activate/rapideye/%v?PL_API_KEY=%v"

const invalidKey = "INVALID_KEY"
const validKey = "VALID_KEY"

var defaultHandlerConfig = util.Configuration{}

func getDiscoverURL(host string, apiKey string) string {
	return fmt.Sprintf("%s/planet/discover/rapideye?PL_API_KEY=%s", host, apiKey)
}

func createMockPlanetAPIServer() *httptest.Server {
	router := mux.NewRouter()
	router.HandleFunc("/data/v1/quick-search", func(writer http.ResponseWriter, request *http.Request) {
		authFields := strings.Fields(request.Header.Get("Authorization"))
		if len(authFields) < 2 {
			writer.WriteHeader(401)
			return
		}
		authMethod := authFields[0]
		authKey, err := base64.StdEncoding.DecodeString(authFields[1])

		if authMethod != "Basic" {
			writer.WriteHeader(400)
			writer.Write([]byte("Queried Planet server with non-basic auth"))
			return
		}

		if err != nil || string(authKey) == invalidKey+":" {
			writer.WriteHeader(401)
			writer.Write([]byte("Bad auth token"))
			return
		}
		writer.WriteHeader(200)
		writer.Write([]byte(`{"type": "FeatureCollection", "features": []}`))
	})
	router.NotFoundHandler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("Route not available in mocked Planet server"))
		writer.WriteHeader(404)
	})
	server := httptest.NewServer(router)
	return server
}

func createTestRouter(planetAPIURL string) *mux.Router {
	handlerConfig := util.Configuration{BasePlanetAPIURL: planetAPIURL}
	router := mux.NewRouter()
	router.Handle("/planet/discover/{itemType}", DiscoverHandler{Config: handlerConfig})
	router.Handle("/planet/{itemType}/{id}", MetadataHandler{Config: handlerConfig})
	router.Handle("/planet/activate/{itemType}/{id}", ActivateHandler{Config: handlerConfig})
	return router
}

func createFixtures() (mockPlanet *httptest.Server, testRouter *mux.Router) {
	mockPlanet = createMockPlanetAPIServer()
	testRouter = createTestRouter(mockPlanet.URL)
	return
}

// ===========

func TestDiscoverHandlerNoAPIKey(t *testing.T) {
	mockServer, router := createFixtures()
	defer mockServer.Close()
	url := getDiscoverURL(mockServer.URL, "")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest("GET", url, nil))
	assert.NotEqual(t, http.StatusOK, recorder.Code,
		"Expected request to fail due to lack of API Key but received: %v, %v", recorder.Code, recorder.Body.String(),
	)
}

func TestDiscoverHandlerInvalidAPIKey(t *testing.T) {
	mockServer, router := createFixtures()
	defer mockServer.Close()
	url := getDiscoverURL(mockServer.URL, invalidKey)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest("GET", url, nil))
	assert.Equal(t, http.StatusUnauthorized, recorder.Code,
		"Expected request to fail due to unauthorized API Key but received: %v, %v", recorder.Code, recorder.Body.String(),
	)
}

func TestDiscoverHandlerSuccess(t *testing.T) {
	mockServer, router := createFixtures()
	defer mockServer.Close()
	url := getDiscoverURL(mockServer.URL, validKey)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest("GET", url, nil))
	assert.Equal(t, http.StatusOK, recorder.Code,
		"Expected request to succeed but received: %v, %v", recorder.Code, recorder.Body.String(),
	)

	_, err := geojson.Parse(recorder.Body.Bytes())
	assert.Nil(t, err, "Expected to parse GeoJSON but received: %v", err)
}

/*
func TestHandlers(t *testing.T) {
	var (
		err     error
		request *http.Request
		fci     interface{}
	)

	mockPlanetRouter := mux.NewRouter()
	mockPlanet := httptest.NewUnstartedServer(mockPlanetRouter)
	mockPlanet.Start()

	handlerConfig := util.Configuration{BasePlanetAPIURL: mockPlanet.URL}
	fmt.Println(handlerConfig)

	discoverHandler := DiscoverHandler{Config: handlerConfig}
	activateHandler := ActivateHandler{Config: handlerConfig}
	metadataHandler := MetadataHandler{Config: handlerConfig}

	// Test: No API Key
	if request, err = http.NewRequest("GET", fmt.Sprintf(fakeDiscoverURL, mockPlanet.URL, ""), nil); err != nil {
		t.Error(err.Error())
	}
	writer, _, _ := util.GetMockResponseWriter()
	discoverHandler.ServeHTTP(writer, request)
	if writer.StatusCode == http.StatusOK {
		t.Errorf("Expected request to fail due to lack of API Key but received: %v, %v", writer.StatusCode, writer.OutputString)
	}

	// Test: Invalid API Key
	if request, err = http.NewRequest("GET", fmt.Sprintf(fakeDiscoverURL, mockPlanet.URL, "foo"), nil); err != nil {
		t.Error(err.Error())
	}
	writer, _, _ = util.GetMockResponseWriter()
	router.ServeHTTP(writer, request)
	if writer.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected request to fail due to unauthorized API Key but received: %v, %v", writer.StatusCode, writer.OutputString)
	}

	// Test: Discover (Happy)
	if request, err = http.NewRequest("GET", fmt.Sprintf(fakeDiscoverURL, mockPlanet.URL, os.Getenv("PL_API_KEY")), nil); err != nil {
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
	metadataURL := fmt.Sprintf(fakeMetadataURL, mockPlanet.URL, id, os.Getenv("PL_API_KEY"))

	if request, err = http.NewRequest("GET", metadataURL, nil); err != nil {
		t.Error(err.Error())
	}
	writer, _, _ = util.GetMockResponseWriter()
	router.ServeHTTP(writer, request)
	if writer.StatusCode != http.StatusOK {
		t.Errorf("Expected request to succeed but received: %v, %v", writer.StatusCode, writer.OutputString)
	}

	// Test: Metadata (no image ID)
	metadataURL = fmt.Sprintf(fakeMetadataURL, mockPlanet.URL, "", os.Getenv("PL_API_KEY"))

	if request, err = http.NewRequest("GET", metadataURL, nil); err != nil {
		t.Error(err.Error())
	}
	writer, _, _ = util.GetMockResponseWriter()
	router.ServeHTTP(writer, request)
	if writer.StatusCode != http.StatusNotFound {
		t.Errorf("Expected request to return a 404 but it returned a %v.", writer.StatusCode)
	}

	// Test: Activate (invalid PL key)
	activateURL := fmt.Sprintf(fakeActivateURL, mockPlanet.URL, id, "foo")
	if request, err = http.NewRequest("POST", activateURL, nil); err != nil {
		t.Error(err.Error())
	}
	writer, _, _ = util.GetMockResponseWriter()
	router.ServeHTTP(writer, request)
	if writer.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected request to return a 401 but it returned a %v.", writer.StatusCode)
	}

	// Test: Activate (happy)
	activateURL = fmt.Sprintf(fakeActivateURL, mockPlanet.URL, id, os.Getenv("PL_API_KEY"))
	if request, err = http.NewRequest("POST", activateURL, nil); err != nil {
		t.Error(err.Error())
	}
	writer, _, _ = util.GetMockResponseWriter()
	// Since this request will routinely fail, we do not check its status
	router.ServeHTTP(writer, request)
}
*/
