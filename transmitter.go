package main

import (
	"sync"
	"time"

	"github.com/xthexder/go-jack"
)

type Transmitter struct {
	outputChannel chan jack.AudioSample
	preamble      []jack.AudioSample
	channelLock   sync.Mutex
}

func NewTransmitter(outputChannel chan jack.AudioSample) *Transmitter {
	t := &Transmitter{
		outputChannel: outputChannel,
	}
	t.preamble = GenerateChirpPreamble(ChirpStartFreq, ChirpEndFreq, FS, PreambleLength)
	SavePreambleToFile("matlab/preamble.csv", t.preamble)
	return t
}

func (t *Transmitter) Send(mframe []int, timeoutChan, freeTimeOutChan chan bool, isACK bool) {
	mframePhy := make([]int, 9+len(mframe))
	copy(mframePhy[0:9], IntToBinaryArray(len(mframe)))
	copy(mframePhy[9:], mframe)
	// Add CRC redundancy bits
	crc := CRC8(mframePhy)
	// Acquire the lock
	t.channelLock.Lock()
	// Modeulate the frame
	for _, sample := range t.preamble {
		t.outputChannel <- jack.AudioSample(sample)
	}
	for _, sample := range modulate(mframePhy) {
		t.outputChannel <- jack.AudioSample(sample)
	}
	for _, sample := range modulate(crc) {
		t.outputChannel <- jack.AudioSample(sample)
	}
	// Release the lock
	t.channelLock.Unlock()
	if isACK {
		return
	}
	counter := 0
	for {
		select {
		case <-freeTimeOutChan:
			return
		default:
			time.Sleep(1 * time.Millisecond)
			counter++
			if counter > 1000 {
				timeoutChan <- true
				return
			}
		}
	}
}

func modulate(frameCRC []int) []jack.AudioSample {
	frameWave := make([]jack.AudioSample, 4*len(frameCRC))
	// Use Line coding  {1,1,1,1} is 0, {-1,-1,-1,-1} is 1
	for i, bit := range frameCRC {
		if bit == 0 {
			frameWave[i*4] = 1
			frameWave[i*4+1] = 1
			frameWave[i*4+2] = 1
			frameWave[i*4+3] = 1
		} else {
			frameWave[i*4] = -1
			frameWave[i*4+1] = -1
			frameWave[i*4+2] = -1
			frameWave[i*4+3] = -1
		}
	}
	return frameWave
}
