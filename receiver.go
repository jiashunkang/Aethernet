package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/xthexder/go-jack"
)

type Receiver struct {
	inputChannel chan jack.AudioSample
	preamble     []jack.AudioSample
	decode_data  []int
	carrier      []float64
}

func NewReceiver(inputChannel chan jack.AudioSample) *Receiver {
	r := &Receiver{
		inputChannel: inputChannel,
	}
	r.preamble = GenerateChirpPreamble(ChirpStartFreq, ChirpEndFreq, FS, PreambleLength)
	carrier := make([]float64, 12000)

	for i := 0; i < 12000; i++ {
		t := float64(i) / 48000 // Time at each sample
		carrier[i] = math.Sin(2 * math.Pi * FC * t)
	}
	r.carrier = carrier
	r.decode_data = make([]int, 0, 50000)
	return r
}

func (r *Receiver) Start() {
	fmt.Println("Start Receiving ...")
	// Sample variables (replace with actual data)
	RxFIFO := make([]float64, 0, 1000000)

	var power, syncPowerLocalMax float64
	var state, startIndex int
	syncFIFO := make([]float64, PreambleLength)
	powerDebug := make([]float64, 0, 1000000)
	syncPowerDebug := make([]float64, 1000000)
	var decodeFIFO []float64
	var totalFrame, correctFrameNum int
	frameSize := 100 // id of syncpower debug
	i := -1
	for {
		jacksample := <-r.inputChannel
		i++
		currentSample := float64(jacksample)
		RxFIFO = append(RxFIFO, currentSample)
		power = power*(1-1.0/64) + math.Pow(currentSample, 2)/64
		powerDebug = append(powerDebug, power)
		if state == 0 {
			syncFIFO = append(syncFIFO[1:], currentSample)
			// syncPowerDebug = append(syncPowerDebug, sumProduct(syncFIFO, r.preamble)/200)
			syncPowerDebug[i] = sumProduct(syncFIFO, r.preamble) / 20
			if syncPowerDebug[i] > power*2 && syncPowerDebug[i] > syncPowerLocalMax && syncPowerDebug[i] > SYNC_PARA {
				syncPowerLocalMax = syncPowerDebug[i]
				startIndex = i
			} else if (i-startIndex > 240) && (startIndex != 0) {
				syncPowerLocalMax = 0
				syncFIFO = make([]float64, len(syncFIFO))
				state = 1
				tempBuffer := RxFIFO[startIndex+1 : i+1]
				decodeFIFO = tempBuffer
			}
		} else if state == 1 {
			decodeFIFO = append(decodeFIFO, currentSample)
			if len(decodeFIFO) == 4*(frameSize+8) {
				decodeFIFOPowerBit := make([]int, frameSize+8)

				for j := 0; j < frameSize+8; j++ {
					if sum(decodeFIFO[1+j*4:2+j*4]) > 0 {
						decodeFIFOPowerBit[j] = 1
					} else {
						decodeFIFOPowerBit[j] = 0
					}
				}

				crcCheck := CRC8(decodeFIFOPowerBit[:frameSize])
				r.decode_data = append(r.decode_data, decodeFIFOPowerBit[:frameSize]...)

				if !isEqual(crcCheck, decodeFIFOPowerBit[frameSize:]) {
					totalFrame++
					fmt.Println("CRC Error ", totalFrame)
				} else {
					correctFrameNum++
					totalFrame++
				}
				startIndex = 0
				decodeFIFO = nil
				state = 0
			}
		}
		if totalFrame == 500 {
			break
		}
		if i > 990000 {
			break
		}
	}
	fmt.Println("Total Frame:", totalFrame)
	fmt.Println("Correct Frame:", correctFrameNum)
	// Save received data to OUTPUT.bin
	byteData := ConvertIntArrayToBitArray(r.decode_data)
	file, err := os.Create("compare/OUTPUT.bin")
	if err != nil {
		log.Fatalf("Error creating OUTPUT.bin: %v", err)
	}
	defer file.Close()
	_, err = file.Write(byteData)
	if err != nil {
		log.Fatalf("Error writing OUTPUT.bin: %v", err)
	}
	fmt.Println("End receiving ...")
	for {
		_ = r.inputChannel
	}
}

func sumProduct(a []float64, b []jack.AudioSample) float64 {
	var sum float64
	for i := range a {
		sum += float64(b[i]) * a[i]
	}
	return sum
}

func smooth(data []float64, windowSize int) []float64 {
	if windowSize <= 1 {
		return data // 如果窗口大小小于等于1，直接返回原数据
	}

	smoothed := make([]float64, len(data))
	windowSum := 0.0

	// 初始化窗口的和
	for i := 0; i < windowSize && i < len(data); i++ {
		windowSum += data[i]
		smoothed[i] = windowSum / float64(i+1) // 初始部分，窗口不完整时的平均值
	}

	// 滑动窗口计算平滑值
	for i := windowSize; i < len(data); i++ {
		windowSum += data[i] - data[i-windowSize] // 加入新的数据点并移除最旧的数据点
		smoothed[i] = windowSum / float64(windowSize)
	}

	return smoothed
}

func multiply(a, b []float64) []float64 {
	result := make([]float64, len(a))
	for i := range a {
		result[i] = a[i] * b[i]
	}
	return result
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

func sum(data []float64) float64 {
	var sum float64
	for _, v := range data {
		sum += v
	}
	return sum
}
