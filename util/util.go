package util

import (
	"github.com/prometheus/common/log"
	"strconv"
)

// Str2float64 converts a string to float64
func Str2float64(str string) float64 {
	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		log.Fatalf("Could not parse '%s' as float!", str)
		return -1
	}
	return value
}
