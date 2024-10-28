package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/xthexder/go-jack"
)

type Transmitter struct {
	MAC_frame_channel chan []int
	outputChannel     chan jack.AudioSample
	preamble          []jack.AudioSample
	data              []int
}

func NewTransmitter(outputChannel chan jack.AudioSample) *Transmitter {
	t := &Transmitter{
		outputChannel: outputChannel,
	}
	t.preamble = GenerateChirpPreamble(ChirpStartFreq, ChirpEndFreq, FS, PreambleLength)
	SavePreambleToFile("matlab/preamble.csv", t.preamble)
	filePath := "compare/INPUT.bin"
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
	}
	t.data = ConvertBitArrayToIntArray(data, 8*len(data))
	return t
}

func (t *Transmitter) readFromFile(fileName string) {
	// Open the file
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	t.data = make([]int, 0, 10000)
	// Read the file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		numbers := strings.Fields(line)
		for _, num := range numbers {
			// 将字符串转换为整数
			value, err := strconv.Atoi(num)
			if err != nil {
				fmt.Println("Error converting string to int:", err)
				continue
			}
			t.data = append(t.data, value)
		}
	}
	fmt.Println("Data Length:", len(t.data))
	fmt.Println("First ten bits:", t.data[:10])
}
func (t *Transmitter) Start() {
	fmt.Println("Start transmitting ...")
	// Separate the data into 100 frames
	frameNum := len(t.data) / 100
	for i := 0; i < frameNum; i++ {
		// Get the next frame
		frame := t.data[i*100 : (i+1)*100]
		// Add CRC redundancy bits
		frameCRC := make([]int, len(frame), 108)
		copy(frameCRC, frame)
		frameCRC = append(frameCRC, CRC8(frame)...)
		// Add Error correction redundancy bits
		// frameEEC = ;
		// Modulate the frame
		frameWave := modulate(frameCRC)
		// Play Preamble
		for _, sample := range t.preamble {
			t.outputChannel <- jack.AudioSample(sample)
		}
		// Play the audio
		for _, sample := range frameWave {
			t.outputChannel <- jack.AudioSample(sample)
		}
	}
	fmt.Println("End transmitting ...")
}

func (t *Transmitter) Send() {
	mframe := <-t.MAC_frame_channel
	mframeCRC := make([]int, len(mframe), 108)
	copy(mframeCRC, mframe)
	// Add CRC redundancy bits
	crc := CRC8(mframe)
	// Modeulate the frame
	for _, sample := range t.preamble {
		t.outputChannel <- jack.AudioSample(sample)
	}
	for _, sample := range modulate(mframe) {
		t.outputChannel <- jack.AudioSample(sample)
	}
	for _, sample := range modulate(crc) {
		t.outputChannel <- jack.AudioSample(sample)
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
