package main

import (
	"fmt"
	"time"

	"github.com/xthexder/go-jack"
)

func main() {
	// Copy the whole track into data
	data_in := make([]jack.AudioSample, 0, 2000000)
	data_out := make([]jack.AudioSample, 0, 2000000)

	client, _ := jack.ClientOpen("AcousticLink", jack.NoStartServer)
	if client == nil {
		fmt.Println("Could not connect to jack server.")
		return
	}
	defer client.Close()

	inPort := client.PortRegister("input", jack.DEFAULT_AUDIO_TYPE, jack.PortIsInput, 0)
	outPort := client.PortRegister("output", jack.DEFAULT_AUDIO_TYPE, jack.PortIsOutput, 0)

	systemInPort := client.GetPortByName("system:capture_1")
	systemOutPort := client.GetPortByName("system:playback_2")

	inputChannel := make(chan jack.AudioSample, 4096)
	outputChannel := make(chan jack.AudioSample, 10000000)
	endChan := make(chan bool)
	mac := NewMAC(1, 0, outputChannel, inputChannel, endChan)
	// transmitter.GenerateInputTxt()

	process := func(nframes uint32) int {
		inBuffer := inPort.GetBuffer(nframes)
		outBuffer := outPort.GetBuffer(nframes)

		for i := range outBuffer {
			select {
			case sample := <-outputChannel:
				outBuffer[i] = sample
				data_out = append(data_out, sample)
			default:
				data_out = append(data_out, jack.AudioSample(0.0))
				outBuffer[i] = 0.0
			}
		}

		for _, sample := range inBuffer {
			data_in = append(data_in, sample)
			inputChannel <- sample

		}

		return 0
	}

	if code := client.SetProcessCallback(process); code != 0 {
		fmt.Println("Failed to set process callback.")
		return
	}

	if code := client.Activate(); code != 0 {
		fmt.Println("Failed to activate client.")
		return
	}

	client.ConnectPorts(systemInPort, inPort)
	client.ConnectPorts(outPort, systemOutPort)
	startTime := time.Now()
	fmt.Println("Start at ", startTime)
	go mac.Start()
	// fmt.Println("Press enter or return to quit...")
	// bufio.NewReader(os.Stdin).ReadString('\n')
	<-endChan
	endTime := time.Now()
	fmt.Println("End at ", endTime)
	fmt.Println("Duration: ", endTime.Sub(startTime))
	err := SavePreambleToFile("track/input_track.csv", data_in)
	if err != nil {
		fmt.Println("Error saving preamble:", err)
	} else {
		fmt.Println("Output saved to matlab/input_track.csv")
	}
	// Write the data to a file, reuse function from utils
	err_out := SavePreambleToFile("track/output_track.csv", data_out)
	if err_out != nil {
		fmt.Println("Error saving preamble:", err)
	} else {
		fmt.Println("Output saved to matlab/output_track.csv")
	}
	fmt.Println("Done.")

}
