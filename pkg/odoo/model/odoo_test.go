package model

import (
	"time"
)

var zurichTZ *time.Location
var vancouverTZ *time.Location

func init() {
	zue, err := time.LoadLocation("Europe/Zurich")
	if err != nil {
		panic(err)
	}
	zurichTZ = zue
	van, err := time.LoadLocation("America/Vancouver")
	if err != nil {
		panic(err)
	}
	vancouverTZ = van
}
