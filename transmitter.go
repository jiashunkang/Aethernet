package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/xthexder/go-jack"
)

type Transmitter struct {
	outputChannel chan jack.AudioSample
	preamble      []jack.AudioSample
	data          []int
}

func NewTransmitter(outputChannel chan jack.AudioSample) *Transmitter {
	t := &Transmitter{
		outputChannel: outputChannel,
	}
	t.preamble = GenerateChirpPreamble(ChirpStartFreq, ChirpEndFreq, FS, PreambleLength)
	SavePreambleToFile("matlab/preamble.csv", t.preamble)
	t.readFromFile("INPUT.txt")
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
	exit_flag := false
	// Separate the data into 100 frames
	for i := 0; i < 101; i++ {
		if i == 100 {
			// last frame always goes wrong and I dont know why
			i = 99
			exit_flag = true
		}
		trivial_frame := make([]int, 8, 108)
		// Get the next frame
		frame := append(trivial_frame, t.data[i*100:(i+1)*100]...)
		// Add CRC redundancy bits
		frameCRC := append(frame, CRC8(frame)...)
		// Add Error correction redundancy bits
		// frameEEC = ;
		// Modulate the frame
		frameWave := modulate(frameCRC)
		// frameWave = append(frameWave, make([]jack.AudioSample, randomSpace)...)

		// Randomly add silence between frames
		randomSpace := rand.Intn(100)
		for i := 0; i < randomSpace; i++ {
			t.outputChannel <- 0.0
		}
		// Play Preamble
		for _, sample := range t.preamble {
			t.outputChannel <- jack.AudioSample(sample)
		}
		// Play the audio
		for _, sample := range frameWave {
			t.outputChannel <- jack.AudioSample(sample)
		}
		for i := 0; i < randomSpace; i++ {
			t.outputChannel <- 0.0
		}
		if exit_flag {
			break
		}
	}
	fmt.Println("End transmitting ...")
}

func modulate(frameCRC []int) []jack.AudioSample {
	frameWave := make([]jack.AudioSample, len(frameCRC)*48)
	// Use PSK modulation with carrier frequency of 10 kHz
	f := float64(10000) // Carrier frequency
	for i, bit := range frameCRC {
		// Define phase shift for PSK: 0 -> phase 0, 1 -> phase π
		phase := 0.0
		if bit == 0 {
			phase = math.Pi
		}
		for j := 0; j < 48; j++ {
			// PSK modulation with phase shift
			frameWave[i*48+j] = jack.AudioSample(math.Sin(2*math.Pi*float64(i*48+j)*f/float64(FS) + phase))
		}
	}
	return frameWave
}
