package landsat

import (
	"compress/gzip"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	"github.com/venicegeo/bf-ia-broker/util"
)

const defaultLandSatHost = "http://landsat-pds.s3.amazonaws.com"

var sceneMap = map[string]string{}
var sceneMapIsReady = false

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
	newSceneMap := map[string]string{}
doneReading:
	for {
		record, readErr := csvReader.Read()
		switch readErr {
		case nil:
			id := record[0]
			url := record[len(record)-1]
			newSceneMap[id] = url
		case io.EOF:
			break doneReading
		default:
			err = readErr
			return
		}
	}

	sceneMap = newSceneMap
	sceneMapIsReady = true
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
func GetSceneFolderURL(sceneID string) (string, error) {
	if !isValidLandsatID(sceneID) {
		return "", fmt.Errorf("Invalid scene ID: %s", sceneID)
	}

	if isOldLandSatID(sceneID) {
		return formatOldIDToURL(sceneID), nil
	}
	if !sceneMapIsReady {
		return "", errors.New("Scene map is not ready yet")
	}
	url, ok := sceneMap[sceneID]
	if !ok {
		return "", errors.New("Scene not found with that ID")
	}
	return url, nil
}

// Old LandSat IDs come back in the form LC80060522017107LGN00
var oldLandSatIDPattern = regexp.MustCompile("LC([0-9]{3})([0-9]{3}).*")

func isOldLandSatID(sceneID string) bool {
	return oldLandSatIDPattern.MatchString(sceneID)
}

// Reference https://landsat.usgs.gov/landsat-collections
var c1LandSatIDPattern = regexp.MustCompile("LC[0-9]{2}_.*")

func isC1LandSatID(sceneID string) bool {
	return c1LandSatIDPattern.MatchString(sceneID)
}

func isValidLandsatID(sceneID string) bool {
	return isOldLandSatID(sceneID) || isC1LandSatID(sceneID)
}

const oldLandSatAWSURL = "https://landsat-pds.s3.amazonaws.com/L8/%s/%s/%s/%s"

func formatOldIDToURL(sceneID string) string {
	m := oldLandSatIDPattern.FindStringSubmatch(sceneID)[1:]
	return fmt.Sprintf(oldLandSatAWSURL, m[0], m[1], sceneID, "")
}
