package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func parseDuration(split [][]byte, startIndex int, endIndex int, entryType string) (int, error) {
	start, err := strconv.ParseInt(string(split[startIndex]), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Failed to parse %w start time in line as integer: %w", entryType, err)
	}
	end, err := strconv.ParseInt(string(split[endIndex]), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Failed to parse %w end time in line as integer: %w", entryType, err)
	}
	return int(end - start), nil
}

func parseTimestamps(split [][]byte, startIndex int, endIndex int, entryType string) (time.Time, time.Time, error) {
	requestTimestamp, err := timeFromUnixBytes(split[startIndex])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("Failed to parse %w requestTimestamp in line as time: %w", entryType, err)
	}
	responseTimestamp, err := timeFromUnixBytes(split[endIndex])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("Failed to parse %w responseTimestamp in line as time: %w", entryType, err)
	}
	return requestTimestamp, responseTimestamp, nil
}

func getGlobalTags() map[string]string {
	return map[string]string{
		"nodeName":        nodeName,
		"simulation":      simulationName,
		"systemUnderTest": systemUnderTest,
		"testEnvironment": testEnvironment,
	}
}

func parseGatlingVersion(split [][]byte, index int) string {
	parsedVersion := string(split[runPattern.gatlingVersion])
	parsedVersion = strings.Trim(parsedVersion, "\r\n")

	return parsedVersion
}
