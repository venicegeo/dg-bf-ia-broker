package tides

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
	"github.com/venicegeo/dg-geojson-go/geojson"
)

const testingSampleTidesResult = `{"locations":[{
	"lat": 0,
	"lon": 0,
	"dtg": "2006-01-02-15-04",
	"results": {
		"minimumTide24Hours": 10,
		"maximumTide24Hours": 20,
		"currentTide": 15
	}
}]}`

// CreateMockTidesServer creates a mocked Tides server instance
// This is exported because it is needed in testing the planet module
func CreateMockTidesServer() *httptest.Server {
	router := mux.NewRouter()
	router.StrictSlash(true)
	router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		writer.Write([]byte(testingSampleTidesResult))
	})
	router.NotFoundHandler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(404)
		writer.Write([]byte("Route not available in mocked Tides server: " + request.URL.String()))
	})
	server := httptest.NewServer(router)
	return server
}

func getTestingFeatureCollection() (fc *geojson.FeatureCollection, err error) {
	fci, err := geojson.ParseFile("testdata/fc.geojson")
	if err != nil {
		return nil, err
	}

	var ok bool
	if fc, ok = fci.(*geojson.FeatureCollection); !ok {
		return nil, fmt.Errorf("Expected FeatureCollection but received %T", fci)
	}
	return fc, nil
}
