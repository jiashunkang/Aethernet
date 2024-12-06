package main

import (
	"acoustic_link/package/shared"
	"bufio"
	"fmt"
	"os"

	"github.com/xthexder/go-jack"
)

func main() {
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
	byteChan := make(chan []byte, 8)
	io := shared.NewIOHelper(byteChan)
	mac := shared.NewMAC(0, 1, outputChannel, inputChannel, io)
	Router := shared.NewRouter("172.182.3.1", io, byteChan)
	// transmitter.GenerateInputTxt()
	process := func(nframes uint32) int {
		inBuffer := inPort.GetBuffer(nframes)
		outBuffer := outPort.GetBuffer(nframes)

		for i := range outBuffer {
			select {
			case sample := <-outputChannel:
				outBuffer[i] = sample
			default:
				outBuffer[i] = 0.0
			}
		}

		for _, sample := range inBuffer {
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
	// Start the MAC and IP layer threads
	go mac.Start()
	go Router.Start()
	// Main thread to read input from user
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("enter exit to quit...")
	for {
		scanner.Scan() // 读取一行输入
		command := scanner.Text()
		if command == "exit" {
			fmt.Println("exit")
			break
		}
	}

	fmt.Println("Done.")

}
