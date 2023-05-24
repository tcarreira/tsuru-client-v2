// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser

import (
	"fmt"
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
		got := DurationFromTimeStrWithoutSeconds(test.timeStr, "default")
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
		got := DurationFromTimeStrWithoutSeconds(test.timeStr, test.defaultVal)
		assert.Equal(t, test.defaultVal, got)
	}
}

func TestSliceToMapFlags(t *testing.T) {
	invalidErrorMsg := "invalid flag %q. Must be on the form \"key=value\""
	for _, test := range []struct {
		in  []string
		out map[string]string
		err error
	}{
		{nil, map[string]string{}, nil},
		{[]string{}, map[string]string{}, nil},
		{[]string{"hey"}, nil, fmt.Errorf(invalidErrorMsg, "hey")},
		{[]string{"hey=hoy"}, map[string]string{"hey": "hoy"}, nil},
		{[]string{"hey=hoy=whaa"}, map[string]string{"hey": "hoy=whaa"}, nil},
		{[]string{"hey=hoy", "two=2"}, map[string]string{"hey": "hoy", "two": "2"}, nil},
		{[]string{"hey=hoy", "two=2", "hey=over"}, map[string]string{"hey": "over", "two": "2"}, nil},
		{[]string{"hey=hoy", "two=2", "all", "must_be_map"}, nil, fmt.Errorf(invalidErrorMsg, "all")},
	} {
		got, err := SliceToMapFlags(test.in)
		if test.err != nil {
			assert.EqualError(t, err, test.err.Error())
		}
		assert.Equal(t, test.out, got)
	}
}

func TestTranslateTimestampSince(t *testing.T) {
	t.Parallel()
	now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, test := range []struct {
		given    time.Time
		expected string
	}{
		{time.Time{}, ""},
		{time.Date(2023, 12, 31, 23, 59, 59, 999, time.UTC), ""},
		{time.Date(2022, 12, 31, 23, 0, 0, 0, time.UTC), "60m"},
		{time.Date(2022, 12, 31, 22, 0, 0, 0, time.UTC), "120m"},
		// {time.Date(2022, 12, 31, 22, 0, 0, 0, time.UTC), "2h"}, // fails
		{time.Date(2022, 12, 31, 20, 0, 0, 0, time.UTC), "4h"},
		{time.Date(2022, 12, 31, 12, 0, 0, 0, time.UTC), "12h"},
		{time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC), "24h"},
		{time.Date(2022, 12, 30, 0, 0, 0, 0, time.UTC), "2d"},
	} {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			got := TranslateDuration(test.given, now)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestCPUValue(t *testing.T) {
	t.Parallel()
	for i, test := range []struct {
		given    string
		expected string
	}{
		{"10m", "1%"},
		{"100m", "10%"},
		{"1000m", "100%"},
		{"2000m", "200%"},
	} {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			assert.Equal(t, test.expected, CPUValue(test.given))
		})
	}
}

func TestMemoryValue(t *testing.T) {
	t.Parallel()
	for i, test := range []struct {
		given    string
		expected string
	}{
		{"256", "256"},
		{"1024", "1Ki"},
		{"16384", "16Ki"},
		{"134217728", "128Mi"},
		{"2147483648", "2Gi"},
		{"2684354560", "2Gi"}, // 2.5Gi shown as 2Gi
		{"3113851290", "2Gi"}, // 2.9Gi shown as 2Gi
	} {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			assert.Equal(t, test.expected, MemoryValue(test.given))
		})
	}
}

func TestIntValue(t *testing.T) {
	assert.Equal(t, "", IntValue(nil))
	i := 1
	assert.Equal(t, "1", IntValue(&i))
}
