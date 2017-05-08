package planet

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const notLandSatID = "NOT_LANDSAT"
const malformedLandSatID = "LC8ABC"
const goodLandSatID = "LC80060522017107LGN00"

func TestLandSatPrefixDetection(t *testing.T) {
	assert.True(t, isLandSatFeature(goodLandSatID))
	assert.True(t, isLandSatFeature(malformedLandSatID))
	assert.False(t, isLandSatFeature("NOT_LANDSAT"))
}

func TestAddLandSatBands_NoOpWhenNotLandSat(t *testing.T) {
	err := addLandsatS3BandsToProperties(notLandSatID, &map[string]interface{}{})
	assert.Nil(t, err)
}

func TestAddLandSatBands_ErrorWhenMalformedID(t *testing.T) {
	err := addLandsatS3BandsToProperties(malformedLandSatID, &map[string]interface{}{})
	assert.NotNil(t, err)
}

func TestAddLandSatBands(t *testing.T) {
	properties := map[string]interface{}{}
	err := addLandsatS3BandsToProperties(goodLandSatID, &properties)
	assert.Nil(t, err)

	bands, ok := properties["bands"]
	assert.True(t, ok, "missing 'bands' in properties")

	bandsMap := bands.(map[string]string)
	for band, suffix := range landSatBandsSuffixes {
		url, found := bandsMap[band]
		assert.True(t, found, "missing band: "+band)
		assert.Contains(t, url, "/800/605/", "URL does not contain correct AWS path")
		assert.Contains(t, url, goodLandSatID)
		assert.True(t, strings.HasSuffix(url, suffix), "wrong suffix for band")
	}
}
