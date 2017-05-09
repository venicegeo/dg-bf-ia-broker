package planet

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// To date LandSat IDs come back in the form LC80060522017107LGN00
// but this could obviously change without notice
var landSatIDPattern = regexp.MustCompile("LC8([0-9]{3})([0-9]{3}).*")

// Inputs: first 3 digits of ID, next 3 digits of ID, ID, filename
const landSatAWSURL = "https://landsat-pds.s3.amazonaws.com/L8/%s/%s/%s/%s"

var landSatBandsSuffixes = map[string]string{
	"coastal":      "_B1.TIF",
	"blue":         "_B2.TIF",
	"green":        "_B3.TIF",
	"red":          "_B4.TIF",
	"nir":          "_B5.TIF",
	"swir1":        "_B6.TIF",
	"swir2":        "_B7.TIF",
	"panchromatic": "_B8.TIF",
	"cirrus":       "_B9.TIF",
	"tirs1":        "_B10.TIF",
	"tirs2":        "_B11.TIF",
}

func isLandSatFeature(id string) bool {
	return strings.HasPrefix(id, "LC8")
}

func addLandsatS3BandsToProperties(landSatID string, properties *map[string]interface{}) error {
	if !isLandSatFeature(landSatID) {
		return nil // Not a LandSat product
	}

	if !landSatIDPattern.MatchString(landSatID) {
		return errors.New("Product ID had 'LC8' prefix but did not match expected LandSat format")
	}

	m := landSatIDPattern.FindStringSubmatch(landSatID)[1:]

	bands := make(map[string]string)
	for band, suffix := range landSatBandsSuffixes {
		filename := landSatID + suffix
		bands[band] = fmt.Sprintf(landSatAWSURL, m[0], m[1], landSatID, filename)
	}
	(*properties)["bands"] = bands

	return nil
}
