package parser

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/dakaraj/gatling-to-influxdb/influx"
	"strings"
)

var (
	oldRequestLinesPattern = requestLinesPattern{blockCount: 7, group: 1, name: 2, startTime: 3, endTime: 4, status: 5, errorMessage: 6}
	newRequestLinesPattern = requestLinesPattern{blockCount: 8, group: 2, name: 3, startTime: 4, endTime: 5, status: 6, errorMessage: 7}
)

type requestLinesPattern struct {
	blockCount   int
	startTime    int
	endTime      int
	name         int
	group        int
	status       int
	errorMessage int
}

func requestLineParser(lb []byte) error {

	split := bytes.Split(lb, tabSep)
	patternParser := oldRequestLinesPattern
	if isGatlingLogFormatV2 {
		patternParser = newRequestLinesPattern
	}

	if len(split) != patternParser.blockCount {
		return errors.New("REQUEST line contains unexpected amount of values")
	}

	requestDuration, err := parseDuration(split, patternParser.startTime, patternParser.endTime, "request")
	if err != nil {
		return err
	}

	requestTimestamp, responseTimestamp, err := parseTimestamps(split, patternParser.startTime, patternParser.endTime, "request")
	if err != nil {
		return err
	}

	requestTags := getGlobalTags()
	requestTags["name"] = strings.TrimSpace(strings.ReplaceAll(string(split[patternParser.name]), " ", "_"))
	requestTags["group"] = strings.TrimSpace(strings.ReplaceAll(string(split[patternParser.group]), " ", "_"))
	requestTags["status"] = string(split[patternParser.status])

	requestData, err := influx.NewPoint(
		"requests",
		requestTags,
		map[string]interface{}{
			"duration":     requestDuration,
			"errorMessage": string(bytes.TrimSpace(split[patternParser.errorMessage])),
		},
		responseTimestamp,
	)
	if err != nil {
		return fmt.Errorf("Error creating new point with request data: %w", err)
	}

	throughputRequest, err := influx.NewPoint(
		"throughputRPS",
		requestTags,
		map[string]interface{}{"request": 1},
		requestTimestamp,
	)
	if err != nil {
		return fmt.Errorf("Error creating new point with request data: %w", err)
	}

	throughputResponce, err := influx.NewPoint(
		"throughputRPS",
		requestTags,
		map[string]interface{}{"response": 1},
		responseTimestamp,
	)
	if err != nil {
		return fmt.Errorf("Error creating new point with request data: %w", err)
	}

	influx.SendPoint(requestData)
	influx.SendPoint(throughputRequest)
	influx.SendPoint(throughputResponce)

	return nil
}
