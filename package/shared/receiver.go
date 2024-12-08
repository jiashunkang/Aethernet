package shared

import (
	"fmt"
	"math"
	"os"

	"github.com/xthexder/go-jack"
)

type Receiver struct {
	inputChannel chan jack.AudioSample
	ackChan      chan ACK
	dataChan     chan Data
	powerChan    chan float64
	preamble     []jack.AudioSample
}

type DebugLog struct {
	AckReceived  int
	DataReceived int
	CRCError     int
}

func (debug *DebugLog) Log() {
	file, err := os.Create("log/receiver.log")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	str := fmt.Sprintf("AckReceived: %d\n", debug.AckReceived)
	str += fmt.Sprintf("DataReceived: %d\n", debug.DataReceived)
	str += fmt.Sprintf("CRCError: %d\n", debug.CRCError)
	_, err = fmt.Fprintf(file, str, debug.AckReceived, debug.DataReceived)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}

func NewReceiver(inputChannel chan jack.AudioSample, ackChan chan ACK, dataChan chan Data, powerChan chan float64) *Receiver {
	r := &Receiver{
		inputChannel: inputChannel,
		ackChan:      ackChan,
		dataChan:     dataChan,
		powerChan:    powerChan,
	}
	r.preamble = GenerateChirpPreamble(ChirpStartFreq, ChirpEndFreq, FS, PreambleLength)
	return r
}

func (r *Receiver) Start() {
	fmt.Println("Start Receiving ...")
	debug := DebugLog{}
	defer debug.Log()
	// Sample variables (replace with actual data)
	RxFIFO := make([]float64, 0, 1000000)

	var power, syncPowerLocalMax float64
	var state, startIndex int
	syncFIFO := make([]float64, PreambleLength)
	var decodeFIFO []float64
	mframeSize := 104 // mac frame size, read from the header
	hasRecordMframeSize := false
	i := -1
	for {
		jacksample := <-r.inputChannel
		i++
		currentSample := float64(jacksample)
		RxFIFO = append(RxFIFO, currentSample)
		power = power*(1-1.0/64) + math.Pow(currentSample, 2)/64
		r.powerChan <- power
		if state == 0 {
			syncFIFO = append(syncFIFO[1:], currentSample)
			syncPowerDebug := sumProduct(syncFIFO, r.preamble) / 20
			if syncPowerDebug > power*2 && syncPowerDebug > syncPowerLocalMax && syncPowerDebug > SYNC_PARA {
				syncPowerLocalMax = syncPowerDebug
				startIndex = i
			} else if (i-startIndex > 24) && (startIndex != 0) {
				syncPowerLocalMax = 0
				syncFIFO = make([]float64, len(syncFIFO))
				state = 1
				tempBuffer := RxFIFO[startIndex+1 : i+1]
				decodeFIFO = tempBuffer
			}
		} else if state == 1 {
			decodeFIFO = append(decodeFIFO, currentSample)
			if len(decodeFIFO) == 4*(9+7) && !hasRecordMframeSize {
				mframeSize = BinaryArrayToInt(demodulate(decodeFIFO[:4*9]))
				header := demodulate(decodeFIFO[4*9:])
				hasRecordMframeSize = true
				// dest src type
				if mframeSize == 7 && header[2] == 1 {
					ack := ACK{
						destId: header[1],
						srcId:  header[0],
						seqNum: BinaryArrayToInt(header[3:])}
					r.ackChan <- ack
					startIndex = 0
					decodeFIFO = nil
					state = 0
					hasRecordMframeSize = false
					debug.AckReceived++
					mframeSize = DATA_SIZE + 7
				}
			}
			if len(decodeFIFO) == 4*(9+mframeSize+8) {
				totalFrame := demodulate(decodeFIFO)
				mframeCRC := totalFrame[9:]
				crcCheck := CRC8(totalFrame[:9+mframeSize])
				if !isEqual(crcCheck, mframeCRC[mframeSize:]) {
					debug.CRCError++
					// fmt.Println("Receiver: CRC Error")
				} else {
					data := Data{
						destId: mframeCRC[1],
						srcId:  mframeCRC[0],
						seqNum: BinaryArrayToInt(mframeCRC[3:7]),
						data:   mframeCRC[7:mframeSize]}
					r.dataChan <- data
					debug.DataReceived++
				}
				startIndex = 0
				decodeFIFO = nil
				state = 0
				hasRecordMframeSize = false
			}
		}
		if i%500000 == 0 {
			// fmt.Println("debuglog")
			debug.Log()
		}
		if i%1000000 == 0 && i != 0 {
			fmt.Println("Receiver: ", i)
			RxFIFO = make([]float64, 0, 1000000)
			power, syncPowerLocalMax = 0, 0
			state, startIndex = 0, 0
			syncFIFO = make([]float64, PreambleLength)
			decodeFIFO = []float64{}
			hasRecordMframeSize = false
			i = -1
		}
	}
}

func demodulate(frameWave []float64) []int {
	res := make([]int, len(frameWave)/4)
	for i := 0; i < len(frameWave)/4; i++ {
		if frameWave[1+i*4] > frameWave[2+i*4] {
			res[i] = 0
		} else {
			res[i] = 1
		}
	}
	return res
}

func sumProduct(a []float64, b []jack.AudioSample) float64 {
	var sum float64
	for i := range a {
		sum += float64(b[i]) * a[i]
	}
	return sum
}

func isEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
