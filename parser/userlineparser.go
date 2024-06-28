package parser

import (
	"bytes"
	"errors"
	"github.com/dakaraj/gatling-to-influxdb/influx"
)

var (
	oldUserLinesPattern = userLinesPattern{blockCount: 4, scenario: 1, timestamp: 3, status: 2}
	newUserLinesPattern = userLinesPattern{blockCount: 6, scenario: 1, timestamp: 5, status: 3}
)

type userLinesPattern struct {
	blockCount int
	scenario   int
	timestamp  int
	status     int
}

func userLineParser(lb []byte) error {
	split := bytes.Split(lb, tabSep)

	patternParser := oldUserLinesPattern
	if isGatlingLogFormatV2 {
		patternParser = newUserLinesPattern
	}

	if len(split) != patternParser.blockCount {
		return errors.New("USER line contains unexpected amount of values")
	}

	timestamp, err := timeFromUnixBytes(bytes.TrimSpace(split[patternParser.timestamp]))
	if err != nil {
		return err
	}

	parsedUserData := influx.UserLineData{
		Timestamp: timestamp,
		Scenario:  string(split[patternParser.scenario]),
		Status:    string(split[patternParser.status]),
	}

	influx.SendUserLineData(parsedUserData)

	return nil
}
