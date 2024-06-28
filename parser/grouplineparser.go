package parser

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/dakaraj/gatling-to-influxdb/influx"
	"strconv"
	"strings"
)

var (
	oldGroupLinesPattern = groupLinesPattern{blockCount: 6, name: 1, startTime: 2, endTime: 3, rawDuration: 4, status: 5}
	newGroupLinesPattern = groupLinesPattern{blockCount: 7, name: 2, startTime: 3, endTime: 4, rawDuration: 5, status: 6}
)

type groupLinesPattern struct {
	blockCount  int
	startTime   int
	endTime     int
	rawDuration int
	name        int
	status      int
}

func groupLineParser(lb []byte) error {

	split := bytes.Split(lb, tabSep)
	patternParser := oldGroupLinesPattern
	if isGatlingLogFormatV2 {
		patternParser = newGroupLinesPattern
	}

	if len(split) != patternParser.blockCount {
		return errors.New("GROUP line contains unexpected amount of values")
	}

	groupDuration, err := parseDuration(split, patternParser.startTime, patternParser.endTime, "group")
	if err != nil {
		return err
	}

	groupRawDuration, err := strconv.ParseInt(string(split[patternParser.rawDuration]), 10, 32)
	if err != nil {
		return fmt.Errorf("Failed to parse group raw duration in line as integer: %w", err)
	}

	groupStartTimestamp, groupEndTimestamp, err := parseTimestamps(split, patternParser.startTime, patternParser.endTime, "group")
	if err != nil {
		return err
	}

	groupTags := getGlobalTags()
	groupTags["name"] = strings.TrimSpace(strings.ReplaceAll(string(split[patternParser.name]), " ", "_"))
	groupTags["status"] = string(split[patternParser.status][:2])

	point, err := influx.NewPoint(
		"groups",
		groupTags,
		map[string]interface{}{
			"totalDuration": groupDuration,
			"rawDuration":   int(groupRawDuration),
		},
		groupEndTimestamp,
	)
	if err != nil {
		return fmt.Errorf("Error creating new point with group data: %w", err)
	}

	groupStartHit, err := influx.NewPoint(
		"throughputTPS",
		groupTags,
		map[string]interface{}{"start": 1},
		groupStartTimestamp,
	)
	if err != nil {
		return fmt.Errorf("Error creating new point with group data: %w", err)
	}

	groupFinishHit, err := influx.NewPoint(
		"throughputTPS",
		groupTags,
		map[string]interface{}{"finish": 1},
		groupEndTimestamp,
	)
	if err != nil {
		return fmt.Errorf("Error creating new point with group data: %w", err)
	}

	influx.SendPoint(point)
	influx.SendPoint(groupStartHit)
	influx.SendPoint(groupFinishHit)

	return nil
}
