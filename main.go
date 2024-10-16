package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/xthexder/go-jack"
)

func main() {
	// Copy the whole track into data
	data := make([]jack.AudioSample, 0, 1000000)

	client, _ := jack.ClientOpen("AcousticLink", jack.NoStartServer)
	if client == nil {
		fmt.Println("Could not connect to jack server.")
		return
	}
	defer client.Close()

	inPort := client.PortRegister("input", jack.DEFAULT_AUDIO_TYPE, jack.PortIsInput, 0)
	outPort := client.PortRegister("output", jack.DEFAULT_AUDIO_TYPE, jack.PortIsOutput, 0)

	systemInPort := client.GetPortByName("system:capture_1")
	systemOutPort := client.GetPortByName("system:playback_1")

	// inputChannel := make(chan jack.AudioSample, 1024)
	outputChannel := make(chan jack.AudioSample, 1024)

	transmitter := NewTransmitter(outputChannel)
	// transmitter.GenerateInputTxt()

	process := func(nframes uint32) int {
		// inBuffer := inPort.GetBuffer(nframes)
		outBuffer := outPort.GetBuffer(nframes)

		// for _, sample := range inBuffer {
		// 	select {
		// 	// case inputChannel <- sample:
		// 	default:
		// 	}
		// }

		for i := range outBuffer {
			select {
			case sample := <-outputChannel:
				outBuffer[i] = sample
				data = append(data, sample)
			default:
				outBuffer[i] = 0.0
			}
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

	// Start Transmitting
	go transmitter.Start()

	fmt.Println("Press enter or return to quit...")
	bufio.NewReader(os.Stdin).ReadString('\n')

	// Write the data to a file, reuse function from utils
	err := SavePreambleToFile("matlab/output_track.csv", data)
	if err != nil {
		fmt.Println("Error saving preamble:", err)
	} else {
		fmt.Println("Output saved to matlab/output_track.csv")
	}

}
