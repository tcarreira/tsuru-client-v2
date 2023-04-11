package parser

import (
	"testing"
	"time"
	_ "unsafe"

	"github.com/stretchr/testify/assert"
)

func TestDurationFromTimeWithoutSeconds(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	timeSince = now.Sub // mocking time.Since via global variable

	for _, test := range []struct {
		timeStr  string
		expected string
	}{
		{"2019-12-31T23:51:00Z", "9m"},
		{"2019-12-31T23:00:00Z", "1h00m"},
		{"2019-12-30T23:51:00Z", "24h09m"},
		{"2019-12-30T12:21:00Z", "35h39m"},
	} {
		got := DurationFromTimeWithoutSeconds(test.timeStr, "default")
		assert.Equal(t, test.expected, got)
	}
}

func TestCPUMillisToPercent(t *testing.T) {
	for _, test := range []struct {
		milli    int32
		expected string
	}{
		{0, "0%"},
		{1, "0%"},
		{5, "0%"},
		{10, "1%"},
		{99, "9%"},
		{100, "10%"},
		{1000, "100%"},
		{2000, "200%"},
		{10000, "1000%"},
	} {
		got := CPUMilliToPercent(test.milli)
		assert.Equal(t, test.expected, got)
	}
}

func TestMemoryToHuman(t *testing.T) {
	for _, test := range []struct {
		memory   int64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{1023, "1023"},
		{1024, "1Ki"},
		{1025, "1025"},
		{1024 * 1024, "1Mi"},
		{1024 * 1024 * 1024, "1Gi"},
		{2 * 1024 * 1024 * 1024, "2Gi"},
		{1024 * 1024 * 1024 * 1024, "1Ti"},
	} {
		got := MemoryToHuman(test.memory)
		assert.Equal(t, test.expected, got)
	}
}
