package landsat

import "regexp"

// Old LandSat IDs come back in the form LC80060522017107LGN00

var oldLandSatIDPattern = regexp.MustCompile("LC8([0-9]{3})([0-9]{3}).*")

// IsOldLandSatID returns whether an ID is in the legacy ID form
func IsOldLandSatID(sceneID string) bool {
	return oldLandSatIDPattern.MatchString(sceneID)
}

var c1LandSatIDPattern = regexp.MustCompile("LC[0-9]{2}_.*")

// IsC1LandSatID returns whether an ID is in the new "Collection 1" ID form
// Reference https://landsat.usgs.gov/landsat-collections
func IsC1LandSatID(sceneID string) bool {
	return c1LandSatIDPattern.MatchString(sceneID)
}

// IsValidLandSatID returns whether an ID is a valid LandSat ID
func IsValidLandSatID(sceneID string) bool {
	return IsOldLandSatID(sceneID) || IsC1LandSatID(sceneID)
}
