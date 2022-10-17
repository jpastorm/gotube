package gotube

import (
	"strings"
)

func extractValue(source, startSource, endSource string) string {
	var start, end int
	if strings.Contains(source, startSource) && strings.Contains(source, endSource) {
		start = strings.Index(source, startSource) + len(startSource)
		end = index(source, endSource, start)
		return source[start:end]
	} else {
		return " "
	}
}

func index(s, substr string, offset int) int {
	if len(s) < offset {
		return -1
	}
	if idx := strings.Index(s[offset:], substr); idx >= 0 {
		return offset + idx
	}
	return -1
}
