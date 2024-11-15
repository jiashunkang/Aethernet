package main

import "testing"

func TestIOhelperConsistency(t *testing.T) {
	io := NewIOHelper()
	for i := 0; i < 100; i++ {
		data := io.ReadData(500)
		end := io.WriteData(data)
		if end {
			io.WriteDataToFile()
		}
	}
	if io.hasData {
		t.Log("IOHelper has data")
	}

	CompareBin()

}
