package planet

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Inputs: mgrs1, mgrs2, mgrs3, year, month, day, filename
const sentinelAWSURL = "https://sentinel-s2-l1c.s3.amazonaws.com/tiles/%s/%s/%s/%d/%d/%d/0/%s"

// https://earth.esa.int/web/sentinel/user-guides/sentinel-2-msi/naming-convention
// TODO: add support for old-style product IDs (which do not contain MGRS info in them)
var sentinelIDPattern = regexp.MustCompile("S2(A|B)_MSIL1C_([0-9]{4})([0-9]{2})([0-9]{2})T[0-9]+_[A-Z0-9]+_[A-Z0-9]+_T([0-9]+)([A-Z])([A-Z]+)_[0-9]{8}T[0-9]")

var sentinelBandsFilenames = map[string]string{
	"coastal":      "B01.jp2",
	"blue":         "B02.jp2",
	"green":        "B03.jp2",
	"red":          "B04.jp2",
	"nir":          "B05.jp2",
	"swir1":        "B06.jp2",
	"swir2":        "B07.jp2",
	"panchromatic": "B08.jp2",
	"cirrus":       "B09.jp2",
	"tirs1":        "B10.jp2",
	"tirs2":        "B11.jp2",
}

func isSentinelFeature(productID string) bool {
	return strings.HasPrefix(productID, "S2A") || strings.HasPrefix(productID, "S2B")
}

func addSentinelS3BandsToProperties(sentinelID string, properties *map[string]interface{}) error {
	if !isSentinelFeature(sentinelID) {
		return nil // Not a Sentinel-2 product
	}
	if !sentinelIDPattern.MatchString(sentinelID) {
		return fmt.Errorf("Product ID had '%s' prefix but did not match expected Sentinel-2 format", sentinelID[:3])
	}

	m := sentinelIDPattern.FindStringSubmatch(sentinelID)
	m = m[2:] // Skip over whole string match and satellite A/B match
	year, err := strconv.Atoi(m[0])
	if err != nil {
		return err
	}
	month, err := strconv.Atoi(m[1])
	if err != nil {
		return err
	}
	day, err := strconv.Atoi(m[2])
	if err != nil {
		return err
	}

	bands := map[string]string{}
	for band, filename := range sentinelBandsFilenames {
		bands[band] = fmt.Sprintf(sentinelAWSURL, m[3], m[4], m[5], year, month, day, filename)
	}
	(*properties)["bands"] = bands

	return nil
}
