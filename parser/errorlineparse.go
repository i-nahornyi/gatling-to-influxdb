package parser

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/dakaraj/gatling-to-influxdb/influx"
)

// / For future extentions
var (
	errorPattern = errorLinePattern{blockCount: 3, errorMessage: 1, timestamp: 2}
)

type errorLinePattern struct {
	blockCount   int
	errorMessage int
	timestamp    int
}

func errorLineParser(lb []byte) error {

	split := bytes.Split(lb, tabSep)
	if len(split) != errorPattern.blockCount {
		return errors.New("ERROR line contains unexpected amount of values")
	}

	timestamp, err := timeFromUnixBytes(bytes.TrimSpace(split[errorPattern.timestamp]))
	if err != nil {
		return err
	}

	errorTags := getGlobalTags()

	point, err := influx.NewPoint(
		"errors",
		errorTags,
		map[string]interface{}{
			"errorMessage": string(split[errorPattern.errorMessage]),
		},
		timestamp,
	)
	if err != nil {
		return fmt.Errorf("Error creating new point with error data: %w", err)
	}

	influx.SendPoint(point)

	return nil
}
