package tides

import (
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
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
