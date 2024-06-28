package parser

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/dakaraj/gatling-to-influxdb/influx"
	l "github.com/dakaraj/gatling-to-influxdb/logger"
	"github.com/hashicorp/go-version"
)

var (
	runPattern = runLinePattern{blockCount: 6, simulationName: 1, testStartTime: 3, description: 4, gatlingVersion: 5}
)

type runLinePattern struct {
	blockCount     int
	simulationName int
	description    int
	testStartTime  int
	gatlingVersion int
}

func runLineParser(lb []byte) error {

	split := bytes.Split(lb, tabSep)
	if len(split) != runPattern.blockCount {
		return errors.New("RUN line contains unexpected amount of values")
	}

	simulationName = string(split[runPattern.simulationName])[strings.LastIndex(string(split[1]), ".")+1:]
	description := string(split[runPattern.description])
	testStartTime, err := timeFromUnixBytes(split[runPattern.testStartTime])
	if err != nil {
		return err
	}
	currentGatlingVersion, err := version.NewVersion(parseGatlingVersion(split, runPattern.gatlingVersion))
	if err != nil {
		return fmt.Errorf("failed to parse version: %w", err)
	}

	// Define the version constraint
	constraint, err := version.NewConstraint(">= 3.5.0")

	// Check if the version meets the constraint
	isGatlingLogFormatV2 := constraint.Check(currentGatlingVersion)

	l.Infof("Gatling log format version %s (Gatling version at least 3.5.0: %v)", currentGatlingVersion.String(), isGatlingLogFormatV2)

	// This will initialize required data for influx client
	influx.InitTestInfo(systemUnderTest, testEnvironment, simulationName, description, nodeName, testStartTime)

	runTags := getGlobalTags()
	runTags["action"] = "start"

	point, err := influx.NewPoint(
		"tests",
		runTags,
		map[string]interface{}{"description": description},
		testStartTime,
	)
	if err != nil {
		return fmt.Errorf("Error creating new point with test start data: %w", err)
	}

	influx.SendPoint(point)

	return nil
}
