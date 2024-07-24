package wal

import (
	"fmt"
	"testing"
)

func TestSomething(t *testing.T) {
	buf := make([]byte, 1)

	buf[0] |= 1          // Is Monotonic - Yes
	buf[0] = buf[0] << 2 // Make room for aggregation type
	buf[0] |= 2          // Aggregation type - Cumulative
	buf[0] = buf[0] << 3 // Make room for metric type
	buf[0] |= 3          // Metric type - Exponential histogram

	fmt.Printf("%b\n", buf[0])

	if buf[0]&0x4 == 0x4 {
		fmt.Println("Summary")
	} else if buf[0]&0x3 == 0x3 {
		fmt.Println("Exponential Histogram")
	} else if buf[0]&0x2 == 0x2 {
		fmt.Println("Histogram")
	} else if buf[0]&0x1 == 0x1 {
		fmt.Println("Sum")
	} else if buf[0]&0x0 == 0x0 {
		fmt.Println("Gauge")
	} else {
		t.Fatal("Unsupported type")
	}

	buf[0] = buf[0] >> 3 // Remove metric type

	if buf[0]&0x2 == 0x2 {
		fmt.Println("Cumulative")
	} else if buf[0]&0x1 == 0x1 {
		fmt.Println("Delta")
	} else if buf[0]&0x0 == 0x0 {
		fmt.Println("Unknown")
	} else {
		t.Fatal("Unsupported type")
	}

	buf[0] = buf[0] >> 2 // Remove metric type

	if buf[0]&0x1 == 0x1 {
		fmt.Println("Monotonic")
	} else if buf[0]&0x0 == 0x0 {
		fmt.Println("Not monotonic")
	} else {
		t.Fatal("Unsupported type")
	}

}
