package planet

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/venicegeo/bf-ia-broker/tides"
)

const testingInvalidKey = "INVALID_KEY"
const testingValidKey = "VALID_KEY"
const testingValidItemID = "foobar123"

const testingSampleSearchResult = `{
	"type": "FeatureCollection",
	"bbox": [100.0, 0.0, 105.0, 1.0],
	"acquiredDate": "2006-01-02T15:04:05Z",
	"features": [{
		"id": "foobar123",
		"type": "Polygon",
		"bbox": [100.0, 0.0, 105.0, 1.0],
		"geometry": {
			"type": "Polygon",
			"coordinates": [[
				[-10.0, -10.0], [10.0, -10.0], [10.0, 10.0], [-10.0, 10.0]
				]]
			},
		"properties": {
			"currentTide": 1.0,
			"acquired": "2006-01-02T15:04:05Z",
			"gsd": 6.5,
			"satellite_id": "RapidEye-1",
			"cloud_cover": 0.5
		},
		"_permissions": ["assets.analytic:download"]
}]}`

const testingSampleFeatureResult = `{
	"id": "foobar123",
	"type": "Polygon",
	"bbox": [100.0, 0.0, 105.0, 1.0],
	"geometry": {
		"type": "Polygon",
		"coordinates": [[
			[-10.0, -10.0], [10.0, -10.0], [10.0, 10.0], [-10.0, 10.0]
			]]
		},
	"properties": {
		"currentTide": 1.0,
		"acquired": "2006-01-02T15:04:05Z",
		"gsd": 6.5,
		"satellite_id": "RapidEye-1",
		"cloud_cover": 0.5
	},
	"_permissions": ["assets.analytic:download"]
	}`

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
		fmt.Fprintln(os.Stderr, " [AUTH ERROR] Fewer than 2 Authorization fields found")
		return false
	}
	authMethod := authFields[0]
	authKey, err := base64.StdEncoding.DecodeString(authFields[1])

	if authMethod != "Basic" {
		fmt.Fprintln(os.Stderr, " [AUTH ERROR] Non-Basic authorization mode")
		return false
	}

	if err != nil || string(authKey) != testingValidKey+":" {
		fmt.Fprintln(os.Stderr, " [AUTH ERROR] Bad auth key", string(authKey), "vs", testingValidKey+":")
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
			writer.Write([]byte(testingSampleSearchResult))
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

		if itemType == "" || itemID != testingValidItemID {
			writer.WriteHeader(404)
			writer.Write([]byte("Not found"))
			return
		}

		writer.WriteHeader(200)
		writer.Write([]byte(testingSampleFeatureResult))
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

func createTestRouter(planetAPIURL string, tidesAPIURL string) *mux.Router {
	os.Setenv("PL_API_URL", planetAPIURL)
	os.Setenv("BF_TIDE_PREDICTION_URL", tidesAPIURL)
	router := mux.NewRouter()
	router.Handle("/planet/discover/{itemType}", NewDiscoverHandler())
	router.Handle("/planet/{itemType}/{id}", NewMetadataHandler())
	router.Handle("/planet/activate/{itemType}/{id}", NewActivateHandler())
	return router
}

func createTestFixtures() (mockPlanet *httptest.Server, mockTides *httptest.Server, testRouter *mux.Router) {
	mockPlanet = createMockPlanetAPIServer()
	mockTides = tides.CreateMockTidesServer()
	testRouter = createTestRouter(mockPlanet.URL, mockTides.URL)
	return
}
