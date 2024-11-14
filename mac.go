package main

import (
	"fmt"

	"github.com/xthexder/go-jack"
)

type MAC struct {
	ioHelper    *IOHelper
	transmitter *Transmitter
	receiver    *Receiver
	macId       int
	targetId    int
	ackChan     chan ACK
	dataChan    chan Data
}

type ACK struct {
	destId int
	srcId  int
}

type Data struct {
	destId int
	srcId  int
	id     int
	data   []int
}

func NewMAC(id, targetId int, outputChannel, inputChannel chan jack.AudioSample) *MAC {
	mac := &MAC{}
	mac.ackChan = make(chan ACK)
	mac.dataChan = make(chan Data)
	mac.ioHelper = NewIOHelper()
	mac.transmitter = NewTransmitter(outputChannel)
	mac.receiver = NewReceiver(inputChannel, mac.ackChan, mac.dataChan)
	mac.macId = id
	mac.targetId = targetId
	return mac
}

func (m *MAC) Start() {
	defer m.ioHelper.WriteDataToFile()

	timeOutChan := make(chan bool)
	freeTimeOutChan := make(chan bool)
	waitingACK := false // if mac has send a frame an dwaiting for ack
	ackFrame := make([]int, 3)
	ackFrame[0] = m.macId
	ackFrame[1] = m.targetId
	ackFrame[2] = 1                        // 0 means data frame, 1 means ack frame
	sendBuffer := make([]int, DATA_SIZE+4) // store temp mac frame with header
	sendBuffer[0] = m.macId
	sendBuffer[1] = m.targetId
	sendBuffer[2] = 0 // 0 means data frame, 1 means ack frame
	sendBuffer[3] = 0 // 0,1 to make difference in case of mistakenly received resending frame
	lastReceivedId := 0
	resend := 0 // resend data for 5 times and give up report error
	// if you have INPUT.bin, then you are a transimtter and you are a receiver in any case.
	receiveEnd := false
	transmitEnd := !m.ioHelper.hasData
	// Start receiver
	go m.receiver.Start()
	for {
		// Satisfy create new transimission condition
		if m.ioHelper.hasData && !waitingACK && resend == 0 {
			copy(sendBuffer[4:], m.ioHelper.ReadData(DATA_SIZE))
			// Flip the bit to indicate sending a new frame
			if sendBuffer[3] == 0 {
				sendBuffer[3] = 1
			} else {
				sendBuffer[3] = 0
			}
			go m.transmitter.Send(sendBuffer, timeOutChan, freeTimeOutChan, false)
			waitingACK = true
		}
		// Timeout waiting ACK
		select {
		case <-timeOutChan:
			resend++
			if resend < 5 {
				go m.transmitter.Send(sendBuffer, timeOutChan, freeTimeOutChan, false)
				fmt.Println("Resend", resend)
			} else {
				// Report error
				resend = 0
				fmt.Println("Error: Link Error")
				return
			}
		default:
			// Do nothing
		}
		// Process ACK
		select {
		case ack := <-m.ackChan:
			if ack.destId == m.macId {
				freeTimeOutChan <- true
				waitingACK = false
				resend = 0
			}
		default:
			// Do nothing
		}
		// Process Receiver Data
		select {
		case data := <-m.dataChan:
			if data.destId == m.macId {
				// Send ACK
				go m.transmitter.Send(ackFrame, timeOutChan, freeTimeOutChan, true)
				if data.id != lastReceivedId {
					lastReceivedId = data.id
					receiveEnd = m.ioHelper.WriteData(data.data)
					if receiveEnd {
						fmt.Println("Receive End")
						m.ioHelper.WriteDataToFile()
					}
				}
			}
		default:
			// Do nothing
		}
		if !transmitEnd {
			if !m.ioHelper.hasData && resend == 0 && !waitingACK {
				transmitEnd = true
				fmt.Println("Transmission End")
			}
		}
		if transmitEnd && receiveEnd {
			break
		}

	}
}
