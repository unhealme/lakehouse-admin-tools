package utils

import (
	"fmt"
	"strconv"
	"strings"
)

var sizeUnits = [...]string{
	"K", "M", "G", "T", "P", "E", "Z", "Y", "R",
}

func FormatSize(si int64) string {
	if si < 1024 {
		return strconv.FormatInt(si, 10)
	}

	sf := float64(si)
	for _, u := range sizeUnits {
		if sf /= 1024.0; sf < 1024.0 {
			return fmt.Sprintf("%3.1f%s", sf, u)
		}
	}
	return fmt.Sprintf("%.1fQ", sf)
}

func ParseSize(si string) (int64, error) {
	sizeUnit := strings.ToUpper(si[len(si)-1:])
	if _, err := strconv.ParseFloat(sizeUnit, 64); err == nil {
		return strconv.ParseInt(si, 10, 64)
	}
	size, err := strconv.ParseFloat(si[:len(si)-1], 64)
	if err != nil {
		return 0, err
	}
	mul := 1024.0
	for _, u := range sizeUnits {
		if sizeUnit == u {
			return int64(size * mul), nil
		}
		mul *= 1024.0
	}
	if sizeUnit == "Q" {
		return int64(size * mul), nil
	}
	return 0, fmt.Errorf("unable to parse size %s to int", si)
}
