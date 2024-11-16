package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

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
	backoffChan chan bool
	powerChan   chan float64
	curPower    float64
	bkoffCount  int // count backoff times, select random number from 2^0 to 2^backoffCount
}

type ACK struct {
	destId int
	srcId  int
	seqNum int
}

type Data struct {
	destId int
	srcId  int
	seqNum int
	data   []int
}

type SenderWindowSlot struct {
	macframe        []int
	timeOutChan     chan bool
	freeTimeOutChan chan bool
	resend          int
	seqNum          int
}

func NewMAC(id, targetId int, outputChannel, inputChannel chan jack.AudioSample) *MAC {
	mac := &MAC{}
	mac.ackChan = make(chan ACK, 10)
	mac.dataChan = make(chan Data, 10)
	mac.backoffChan = make(chan bool, 10)
	mac.powerChan = make(chan float64, 10000000)
	mac.ioHelper = NewIOHelper()
	mac.transmitter = NewTransmitter(outputChannel)
	mac.receiver = NewReceiver(inputChannel, mac.ackChan, mac.dataChan, mac.powerChan)
	mac.macId = id
	mac.targetId = targetId
	mac.curPower = 0
	mac.bkoffCount = 0
	return mac
}

func (m *MAC) Start() {
	defer m.ioHelper.WriteDataToFile()
	lastAckReceived := 2                    // Sender Protocol: LAR
	lastFrameSent := 2                      // Sender Protocol: LFS
	lastFrameReceived := 2                  // Receiver Protocol: LFR
	largestAcceptFrame := 2 + S_WINDOW_SIZE // Receiver Protocol: LAF
	ackFrame := make([]int, 7)
	ackFrame[0] = m.macId
	ackFrame[1] = m.targetId
	ackFrame[2] = 1                                           // 0 means data frame, 1 means ack frame
	receiveWindow := make([]*Data, 10000)                     // sliding window big enough so that do not need to worry about overflow
	sendWindow := make([]*SenderWindowSlot, 0, S_WINDOW_SIZE) // sliding window
	// if you have INPUT.bin, then you are a transimtter and you are a receiver in any case.
	receiveEnd := false
	transmitEnd := !m.ioHelper.hasData
	// If mac is during back off
	isBackoff := false
	// Start receiver
	go m.receiver.Start()
	for {
		// Satisfy create new transimission condition
		if m.ioHelper.hasData && Minus(lastFrameSent, lastAckReceived) < S_WINDOW_SIZE {
			// Sense Medium & Backoff
			if isBackoff {
				// Do nothing
			} else if m.senseSignal() {
				isBackoff = true
				go m.backoff(RTT)
			} else {
				m.bkoffCount = 0 // count backoff times, select random number from 2^0 to 2^backoffCount
				// Create window slot
				slot := &SenderWindowSlot{}
				slot.timeOutChan = make(chan bool)
				slot.freeTimeOutChan = make(chan bool, 2)
				slot.resend = 0
				lastFrameSent++
				lastFrameSent %= 16
				slot.seqNum = lastFrameSent
				// Send data
				slot.macframe = make([]int, DATA_SIZE+3+4)
				slot.macframe[0] = m.macId    // this mac id
				slot.macframe[1] = m.targetId // receiver mac id
				slot.macframe[2] = 0          // 0 means data frame, 1 means ack frame
				// bit 3,4,5,6 represent frame id
				copy(slot.macframe[3:7], IntToBinaryArray(slot.seqNum)[5:9])
				// bit 7 - end represent data
				copy(slot.macframe[7:], m.ioHelper.ReadData(DATA_SIZE))
				// Add to window
				sendWindow = append(sendWindow, slot)
				go m.transmitter.Send(slot.macframe, slot.timeOutChan, slot.freeTimeOutChan, false)
			}
		}
		// Timeout waiting ACK
		for _, slot := range sendWindow {
			select {
			case <-slot.timeOutChan:
				if isBackoff {
					// Do nothing
					slot.timeOutChan <- true
				} else if m.senseSignal() {
					slot.timeOutChan <- true
					isBackoff = true
					go m.backoff(RTT)
				} else {
					m.bkoffCount = 0 // count backoff times, select random number from 2^0 to 2^backoffCount
					if slot.resend < MAX_RESEND {
						// Sense Medium & Backoff
						slot.resend++
						go m.transmitter.Send(slot.macframe, slot.timeOutChan, slot.freeTimeOutChan, false)
						fmt.Println("Resend", slot.resend, "SeqNum", slot.seqNum)
					} else {
						// Report error
						slot.resend = 0
						fmt.Println("Error: Link Error")
						return
					}
				}
			default:
				// Do nothing
			}
		}
		// Process ACK
		select {
		case ack := <-m.ackChan:
			fmt.Println("Ack ", ack.seqNum)
			if ack.destId == m.macId {
				if GreaterThan(ack.seqNum, lastAckReceived) && LessEqual(ack.seqNum, lastFrameSent) {
					lastAckReceived = ack.seqNum
					// Clear window slot with seqNum <= ack.seqNum
					for _, slot := range sendWindow {
						if LessEqual(slot.seqNum, ack.seqNum) {
							slot.freeTimeOutChan <- true // this should be a buffered channel (no waiting)
							if len(sendWindow) > 0 {
								sendWindow = sendWindow[1:]
							} else {
								sendWindow = make([]*SenderWindowSlot, 0, S_WINDOW_SIZE)
							}
						}
					}
				}
			}
		default:
			// Do nothing
		}
		// Process Receiver Data
		select {
		case data := <-m.dataChan:
			fmt.Println("Data ", data.seqNum)
			if data.destId == m.macId {
				if GreaterThan(data.seqNum, lastFrameReceived) && LessEqual(data.seqNum, largestAcceptFrame) {
					slotid := Minus(data.seqNum, (lastFrameReceived+1)%16)
					fmt.Println("Slotid", slotid)
					receiveWindow[slotid] = &data
					// Update LFR
					slide := 0
					for _, d := range receiveWindow[0:S_WINDOW_SIZE] {
						if d != nil {
							lastFrameReceived = d.seqNum
							largestAcceptFrame = lastFrameReceived + S_WINDOW_SIZE
							fmt.Println("Data length ", len(d.data))
							receiveEnd = m.ioHelper.WriteData(d.data)
							if receiveEnd {
								m.ioHelper.WriteDataToFile()
							}
							slide++
						} else {
							break
						}
					}
					// Slide window
					receiveWindow = receiveWindow[slide:]

				}
				// Send Accumulative Ack
				copy(ackFrame[3:], IntToBinaryArray(lastFrameReceived)[5:9])
				go m.transmitter.Send(ackFrame, nil, nil, true)
			}
		default:
			// Do nothing
		}
		if !transmitEnd {
			if !m.ioHelper.hasData && len(sendWindow) == 0 {
				transmitEnd = true
				fmt.Println("Transmission End")
			}
		}
		if transmitEnd && receiveEnd {
			break
		}
		// Finish backoff
		select {
		case <-m.backoffChan:
			isBackoff = false
		default:
			// Do nothing
		}

	}
}

func (m *MAC) backoff(milisecond int) {
	// backoff
	m.bkoffCount++
	if (m.bkoffCount) > 4 {
		m.bkoffCount = 0
	}
	num := math.Pow(2, float64(rand.Intn(m.bkoffCount)))
	time.Sleep(time.Duration(num) * time.Duration(milisecond) * time.Millisecond)
	m.backoffChan <- true
}

func (m *MAC) senseSignal() bool {
	exitLoop := false
	for !exitLoop {
		select {
		case m.curPower = <-m.powerChan:
			continue
		default:
			exitLoop = true
		}
	}
	return m.curPower > POWER_SIGNAL
}
