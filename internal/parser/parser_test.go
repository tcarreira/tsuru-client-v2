package parser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDurationFromTimeWithoutSeconds(t *testing.T) {
	oldTimeSince := timeSince
	now, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	timeSince = now.Sub                         // mocking time.Since via global variable
	defer func() { timeSince = oldTimeSince }() // restore time.Since

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
func TestDurationFromTimeWithoutSeconds_OnError(t *testing.T) {
	for _, test := range []struct {
		timeStr    string
		defaultVal string
	}{
		{"2019", "not time.RFC3339_1"},
		{"2019-12-31", "not time.RFC3339_2"},
		{"2019-12-30T23:51:00", "not time.RFC3339_3"},
		{"invalid-string", "any string"},
		{"2019-13-30T23:51:00Z", "wrong month"},
		{"2019-02-29T23:51:00Z", "wrong day"},
		{"2019-12-30T24:51:00Z", "wrong hour"},
		{"2019-12-30T23:71:00Z", "wrong minute"},
		{"2019-12-30T23:51:90Z", "wrong second"},
		{"2019-12-30T23:51:00L", "wrong tz"},
	} {
		got := DurationFromTimeWithoutSeconds(test.timeStr, test.defaultVal)
		assert.Equal(t, test.defaultVal, got)
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
