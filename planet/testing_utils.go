package planet

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
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
const testingValidItemType = "wumpus456"

var testingSampleSearchResult string
var testingSampleFeatureResult string
var testingSampleAssetsResult string
var testingSampleActivateResult string

func initSampleTestingFiles() {
	var err error
	var data []byte
	panicCheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	data, err = ioutil.ReadFile("testdata/testingSampleSearchResult.json")
	panicCheck(err)
	testingSampleSearchResult = string(data)

	data, err = ioutil.ReadFile("testdata/testingSampleFeatureResult.json")
	panicCheck(err)
	testingSampleFeatureResult = string(data)

	data, err = ioutil.ReadFile("testdata/testingSampleAssetsResult.json")
	panicCheck(err)
	testingSampleAssetsResult = strings.Replace(string(data), "https://api.planet.com", "++API_URL_PLACEHOLDER++", -1)

	data, err = ioutil.ReadFile("testdata/testingSampleActivateResult.json")
	panicCheck(err)
	testingSampleActivateResult = string(data)
}

func makeDiscoverTestingURL(host string, apiKey string) string {
	return fmt.Sprintf("%s/planet/discover/%s?PL_API_KEY=%s", host, "rapideye", apiKey)
}

func makeMetadataTestingURL(host string, apiKey string, itemType string, id string) string {
	return fmt.Sprintf("%s/planet/%s/%s?PL_API_KEY=%s", host, itemType, id, apiKey)
}

func makeActivateTestingURL(host string, apiKey string, itemType string, id string) string {
	return fmt.Sprintf("%s/planet/activate/%s/%s?PL_API_KEY=%s", host, itemType, id, apiKey)
}

func testingCheckAuthorization(authHeader string) bool {
	authFields := strings.Fields(authHeader)
	if len(authFields) < 2 {
		fmt.Fprintln(os.Stderr, " [MOCK AUTH ERROR] Fewer than 2 Authorization fields found")
		return false
	}
	authMethod := authFields[0]
	authKey, err := base64.StdEncoding.DecodeString(authFields[1])

	if authMethod != "Basic" {
		fmt.Fprintln(os.Stderr, " [MOCK AUTH ERROR] Non-Basic authorization mode")
		return false
	}

	if err != nil || string(authKey) != testingValidKey+":" {
		fmt.Fprintln(os.Stderr, " [MOCK AUTH ERROR] Bad auth key", string(authKey), "vs", testingValidKey+":")
		return false
	}
	return true
}

func createMockPlanetAPIServer() (server *httptest.Server) {
	router := mux.NewRouter()
	router.StrictSlash(false)
	router.HandleFunc("/data/v1/quick-search", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(os.Stdout, " ***** DEBUG ***** Headers for request:\n")
		request.Header.Write(os.Stdout)
		if testingCheckAuthorization(request.Header.Get("Authorization")) {
			writer.WriteHeader(200)
			writer.Write([]byte(testingSampleSearchResult))
		} else {
			writer.WriteHeader(401)
			writer.Write([]byte("Unauthorized"))
		}
	})

	router.HandleFunc("/data/v1/item-types/{itemType}/items/{itemID}", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(os.Stdout, " ***** DEBUG ***** Headers for request:\n")
		request.Header.Write(os.Stdout)
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
		fmt.Fprintf(os.Stdout, " ***** DEBUG ***** Headers for request:\n")
		request.Header.Write(os.Stdout)
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
		result := strings.Replace(testingSampleAssetsResult, "++API_URL_PLACEHOLDER++", server.URL, -1)
		writer.Write([]byte(result))
	})

	router.HandleFunc("/data/v1/assets/{assetID}/activate", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(os.Stdout, " ***** DEBUG ***** Headers for request:\n")
		request.Header.Write(os.Stdout)
		writer.WriteHeader(200)
		writer.Write([]byte(testingSampleActivateResult))
	})

	router.NotFoundHandler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(404)
		writer.Write([]byte("Route not available in mocked Planet server" + request.URL.String()))
	})
	server = httptest.NewServer(router)
	return
}

// createTestRouter creates a router for testing use only,
// providing a way mock a server for the handlers being tested
// to live in
func createTestRouter(planetAPIURL string, tidesAPIURL string) *mux.Router {
	os.Setenv("PL_API_URL", planetAPIURL)
	os.Setenv("BF_TIDE_PREDICTION_URL", tidesAPIURL)
	router := mux.NewRouter()
	router.Handle("/planet/discover/{itemType}", NewDiscoverHandler())
	router.Handle("/planet/activate/{itemType}/{id}", NewActivateHandler())
	router.Handle("/planet/{itemType}/{id}", NewMetadataHandler())
	return router
}

func createTestFixtures() (mockPlanet *httptest.Server, mockTides *httptest.Server, testRouter *mux.Router) {
	mockPlanet = createMockPlanetAPIServer()
	mockTides = tides.CreateMockTidesServer()
	testRouter = createTestRouter(mockPlanet.URL, mockTides.URL)
	return
}
