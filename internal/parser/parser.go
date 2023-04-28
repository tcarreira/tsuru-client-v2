// Copyright © 2023 tsuru-client authors
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parser

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	timeSince = time.Since // for mocking time.Since in tests
)

// DurationFromTimeStrWithoutSeconds returns a string representing the duration,
// without seconds, between the given time and now.
// eg: 1h2m
func DurationFromTimeStrWithoutSeconds(timeStr string, defaultOnError string) string {
	createdAt, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return defaultOnError
	}
	return DurationFromTimeWithoutSeconds(createdAt, defaultOnError)
}

// DurationFromTimeWithoutSeconds returns a string representing the duration,
// without seconds, between the given time and now.
// eg: 1h2m
func DurationFromTimeWithoutSeconds(createdAt time.Time, defaultOnError string) string {
	age := timeSince(createdAt).Truncate(1 * time.Minute)
	if age >= 1*time.Hour {
		return fmt.Sprintf("%dh%02dm", age/time.Hour, age%time.Hour/time.Minute)
	}
	return fmt.Sprintf("%dm", age/time.Minute)
}

// CPUValue parses CPU as Quantity and returns a the percentage of a CPU core (as string).
func CPUValue(q string) string {
	qt, err := resource.ParseQuantity(q)
	if err == nil {
		return fmt.Sprintf("%d%%", qt.MilliValue()/10)
	}
	return ""
}

// MemoryValue parses Memory as Quantity and returns a human readable quantity.
func MemoryValue(q string) string {
	qt, err := resource.ParseQuantity(q)
	if err != nil {
		return ""
	}
	m := qt.Value()
	if m >= 1024*1024*1024 {
		return fmt.Sprintf("%dGi", m/1024/1024/1024)
	} else if m >= 1024*1024 {
		return fmt.Sprintf("%dMi", m/1024/1024)
	} else if m >= 1024 {
		return fmt.Sprintf("%dKi", m/1024)
	}
	return fmt.Sprintf("%d", m)
}

// CPUMilliToPercent returns a string representing the percentage of a CPU core.
// eg: 1000 = 100%, 200 = 20%
func CPUMilliToPercent(milli int32) string {
	return fmt.Sprintf("%d%%", milli/10)
}

// MemoryToHuman returns a string representing the memory in a human readable format.
// eg: 1024 = 1Ki, 1024*1024 = 1Mi, 1024*1024*1024 = 1Gi
func MemoryToHuman(memory int64) string {
	return resource.NewQuantity(memory, resource.BinarySI).String()
}
