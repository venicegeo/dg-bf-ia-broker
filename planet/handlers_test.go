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
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/venicegeo/dg-geojson-go/geojson"
)

func TestDiscoverHandlerNoAPIKey(t *testing.T) {
	mockServer, _, router := createTestFixtures()
	url := makeDiscoverTestingURL(mockServer.URL, "")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest("GET", url, nil))
	assert.NotEqual(t, http.StatusOK, recorder.Code,
		"Expected request to fail due to lack of API Key but received: %v, %v", recorder.Code, recorder.Body.String(),
	)
}

func TestDiscoverHandlerInvalidAPIKey(t *testing.T) {
	mockServer, _, router := createTestFixtures()
	url := makeDiscoverTestingURL(mockServer.URL, testingInvalidKey)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest("GET", url, nil))
	assert.Equal(t, http.StatusUnauthorized, recorder.Code,
		"Expected request to fail due to unauthorized API Key but received: %v, %v", recorder.Code, recorder.Body.String(),
	)
}

func TestDiscoverHandlerSuccess(t *testing.T) {
	mockServer, _, router := createTestFixtures()
	url := makeDiscoverTestingURL(mockServer.URL, testingValidKey)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest("GET", url, nil))
	assert.Equal(t, http.StatusOK, recorder.Code,
		"Expected request to succeed but received: %v, %v", recorder.Code, recorder.Body.String(),
	)

	_, err := geojson.Parse(recorder.Body.Bytes())
	assert.Nil(t, err, "Expected to parse GeoJSON but received: %v", err)
}

func TestMetadataHandlerSuccess(t *testing.T) {
	mockServer, _, router := createTestFixtures()
	url := makeMetadataTestingURL(mockServer.URL, testingValidKey, "rapideye", testingValidItemID)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest("GET", url, nil))
	assert.Equal(t, http.StatusOK, recorder.Code,
		"Expected request to succeed but received: %v, %v", recorder.Code, recorder.Body.String(),
	)
}

func TestMetadataHandlerImageIDNotFound(t *testing.T) {
	mockServer, _, router := createTestFixtures()
	url := makeMetadataTestingURL(mockServer.URL, testingValidKey, "rapideye", "")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest("GET", url, nil))
	assert.Equal(t, http.StatusNotFound, recorder.Code,
		"Expected request to return a 404 but it returned a %v.", recorder.Code,
	)
}

func TestActivateHandlerInvalidKey(t *testing.T) {
	mockServer, _, router := createTestFixtures()
	url := makeActivateTestingURL(mockServer.URL, testingInvalidKey, testingValidItemType, testingValidItemID)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest("POST", url, nil))
	assert.Equal(t, http.StatusUnauthorized, recorder.Code,
		"Expected request to return a 401 but it returned a %v.", recorder.Code,
	)
}

func TestActivateHandlerSuccess(t *testing.T) {
	mockServer, _, router := createTestFixtures()
	url := makeActivateTestingURL(mockServer.URL, testingValidKey, testingValidItemType, testingValidItemID)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest("POST", url, nil))
	assert.Equal(t, http.StatusOK, recorder.Code,
		"Unexpected error in response to request. %v %v", recorder.Code, recorder.Body.String(),
	)

	assert.Equal(t, testingSampleActivateResult, recorder.Body.String(),
		"Unexpected result for asset activation query",
	)
}
