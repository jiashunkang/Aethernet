package main

import (
	"math/rand"
	"testing"
	"time"

	"github.com/xthexder/go-jack"
)

func TestReceiver(t *testing.T) {
	// Read data from matlab/output_track.csv
	data, _ := ReadFromCsvFile("matlab/input_track.csv")
	// Declare input channel to simulate input data
	inputChannel := make(chan jack.AudioSample, 1024)
	// Create a new receiver
	receiver := NewReceiver(inputChannel)
	go func() {
		// Create a go routine to transmit data to the channel
		// Sleep a while
		for _, sample := range data {
			inputChannel <- sample
		}
		for {
			inputChannel <- jack.AudioSample(rand.Float64() - 0.5)
		}
	}()
	// Start the receiver
	go receiver.Start()
	// wait for the receiver to finish
	time.Sleep(15 * time.Second)
	Compare()
}
