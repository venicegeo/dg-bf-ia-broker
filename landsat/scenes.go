package landsat

import (
	"compress/gzip"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/venicegeo/dg-bf-ia-broker/util"
)

type sceneMapRecord struct {
	awsFolderURL string
	filePrefix   string
}

const defaultLandSatHost = "http://landsat-pds.s3.amazonaws.com"

var sceneMap = map[string]sceneMapRecord{}

// SceneMapIsReady contains a flag of whether the scene map has been loaded yet
var SceneMapIsReady = false

// UpdateSceneMap updates the global scene map from a remote source
func UpdateSceneMap() (err error) {
	landSatHost := os.Getenv("LANDSAT_HOST")
	if landSatHost == "" {
		landSatHost = defaultLandSatHost
	}
	sceneListURL := fmt.Sprintf("%s/c1/L8/scene_list.gz", landSatHost)

	c := util.HTTPClient()
	response, err := c.Get(sceneListURL)
	if err != nil {
		return
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		err = fmt.Errorf("Non-200 response code: %d", response.StatusCode)
		return
	}

	rawReader := response.Body
	gzipReader, err := gzip.NewReader(rawReader)
	if err != nil {
		return
	}

	csvReader := csv.NewReader(gzipReader)
	newSceneMap := map[string]sceneMapRecord{}
doneReading:
	for {
		record, readErr := csvReader.Read()
		switch readErr {
		case nil:
			// First column contains file prefix
			filePrefix := record[0]
			// Second column contains scene ID
			id := record[1]
			// Last column contains URL
			url := record[len(record)-1]
			// Strip the "index.html" file name to just get the directory path
			lastSlash := strings.LastIndex(url, "/")
			url = url[:lastSlash+1]

			newSceneMap[id] = sceneMapRecord{filePrefix: filePrefix, awsFolderURL: url}
		case io.EOF:
			break doneReading
		default:
			err = readErr
			return
		}
	}

	sceneMap = newSceneMap
	SceneMapIsReady = true
	return nil
}

// UpdateSceneMapAsync runs UpdateSceneMap asynchronously, returning
// completion signals via channels
func UpdateSceneMapAsync() (done chan bool, errored chan error) {
	done = make(chan bool)
	errored = make(chan error)
	go func() {
		err := UpdateSceneMap()
		if err == nil {
			done <- true
		} else {
			errored <- err
		}
		close(done)
		close(errored)
	}()
	return
}

// UpdateSceneMapOnTicker updates the scene map on a loop with a delay of
// a given duration. It logs any errors using the given LogContext
func UpdateSceneMapOnTicker(d time.Duration, ctx util.LogContext) {
	ticker := time.NewTicker(d)
	for {
		done, errored := UpdateSceneMapAsync()
		select {
		case <-done:
		case err := <-errored:
			util.LogAlert(ctx, "Failed to update scene ID to URL map: "+err.Error())
		}
		<-ticker.C
	}
}

// GetSceneFolderURL returns the AWS S3 URL at which the scene files for this
// particular scene are available
func GetSceneFolderURL(sceneID string, dataType string) (folderURL string, filePrefix string, err error) {
	if !IsValidLandSatID(sceneID) {
		return "", "", fmt.Errorf("Invalid scene ID: %s", sceneID)
	}

	if IsPreCollectionDataType(dataType) {
		return formatPreCollectionIDToURL(sceneID), sceneID, nil
	}
	if !IsCollection1DataType(dataType) {
		return "", "", errors.New("Unknown LandSat data type: " + dataType)
	}

	if !SceneMapIsReady {
		return "", "", errors.New("Scene map is not ready yet")
	}
	record, ok := sceneMap[sceneID]
	if !ok {
		return "", "", errors.New("Scene not found with ID: " + sceneID)
	}

	return record.awsFolderURL, record.filePrefix, nil
}
