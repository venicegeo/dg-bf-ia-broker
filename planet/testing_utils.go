package planet

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gorilla/mux"
	"github.com/venicegeo/bf-ia-broker/util"
)

const testingInvalidKey = "INVALID_KEY"
const testingValidKey = "VALID_KEY"
const testingValidItemID = "foobar123"

func makeDiscoverTestingURL(host string, apiKey string) string {
	return fmt.Sprintf("%s/planet/discover/rapideye?PL_API_KEY=%s", host, apiKey)
}

func makeMetadataTestingURL(host string, apiKey string, itemType string, id string) string {
	return fmt.Sprintf("%s/planet/%s/%s?PL_API_KEY=%s", host, itemType, id, apiKey)
}

func makeActivateTestingURL(host string, apiKey string, id string) string {
	return fmt.Sprintf("%s/planet/rapideye/%s?PL_API_KEY=%s", host, id, apiKey)
}

func testingCheckAuthorization(authHeader string) bool {
	authFields := strings.Fields(authHeader)
	if len(authFields) < 2 {
		return false
	}
	authMethod := authFields[0]
	authKey, err := base64.StdEncoding.DecodeString(authFields[1])

	if authMethod != "Basic" {
		return false
	}

	if err != nil || string(authKey) != testingValidKey+":" {
		return false
	}
	return true
}

func createMockPlanetAPIServer() *httptest.Server {
	router := mux.NewRouter()
	router.StrictSlash(true)
	router.HandleFunc("/data/v1/quick-search", func(writer http.ResponseWriter, request *http.Request) {
		if testingCheckAuthorization(request.Header.Get("Authorization")) {
			writer.WriteHeader(200)
			writer.Write([]byte(`{"type": "FeatureCollection", "features": []}`))
		} else {
			writer.WriteHeader(401)
			writer.Write([]byte("Unauthorized"))
		}
	})

	router.HandleFunc("/data/v1/item-types/{itemType}/items/{itemID}", func(writer http.ResponseWriter, request *http.Request) {
		if !testingCheckAuthorization(request.Header.Get("Authorization")) {
			writer.WriteHeader(401)
			writer.Write([]byte("Unauthorized"))
			return
		}
		itemType := mux.Vars(request)["itemType"]
		itemID := mux.Vars(request)["itemID"]

		if itemType == "" || itemID == "" {
			writer.WriteHeader(404)
			writer.Write([]byte("Not found"))
			return
		}

		writer.WriteHeader(200)
		writer.Write([]byte("{}"))
	})

	router.HandleFunc("/data/v1/item-types/{itemType}/items/{itemID}/assets", func(writer http.ResponseWriter, request *http.Request) {
		if !testingCheckAuthorization(request.Header.Get("Authorization")) {
			writer.WriteHeader(401)
			writer.Write([]byte("Unauthorized"))
			return
		}
		itemType := mux.Vars(request)["itemType"]
		itemID := mux.Vars(request)["itemID"]

		if itemType == "" || itemID != testingValidItemID {
			writer.WriteHeader(404)
			writer.Write([]byte("Not found"))
			return
		}

		writer.WriteHeader(200)
		writer.Write([]byte("{}"))
	})
	router.NotFoundHandler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(404)
		writer.Write([]byte("Route not available in mocked Planet server" + request.URL.String()))
	})
	server := httptest.NewServer(router)
	return server
}

func createMockTidesServer() *httptest.Server {
	router := mux.NewRouter()
	router.StrictSlash(true)
	router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		writer.Write([]byte("{}"))
	})
	router.NotFoundHandler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(404)
		writer.Write([]byte("Route not available in mocked Tides server: " + request.URL.String()))
	})
	server := httptest.NewServer(router)
	return server
}

func createTestRouter(planetAPIURL string, tidesAPIURL string) *mux.Router {
	handlerConfig := util.Configuration{
		BasePlanetAPIURL: planetAPIURL,
		TidesAPIURL:      tidesAPIURL,
	}
	router := mux.NewRouter()
	router.Handle("/planet/discover/{itemType}", DiscoverHandler{Config: handlerConfig})
	router.Handle("/planet/{itemType}/{id}", MetadataHandler{Config: handlerConfig})
	router.Handle("/planet/activate/{itemType}/{id}", ActivateHandler{Config: handlerConfig})
	return router
}

func createTestFixtures() (mockPlanet *httptest.Server, mockTides *httptest.Server, testRouter *mux.Router) {
	mockPlanet = createMockPlanetAPIServer()
	mockTides = createMockTidesServer()
	testRouter = createTestRouter(mockPlanet.URL, mockTides.URL)
	return
}
