package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"

	"github.com/sigurn/crc8"
	"github.com/xthexder/go-jack"
)

func GenerateInputTxt() {
	file, err := os.Create("INPUT.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for i := 0; i < 10000; i++ {
		bit := rand.Intn(2)
		file.WriteString(fmt.Sprint(bit))
		// add a space between every bit
		file.WriteString(" ")
	}
}
func ConvertIntArrayToBitArray(intArray []int) []byte {
	byteArrayLength := (len(intArray) + 7) / 8
	remainder := len(intArray) % 8
	if remainder == 0 && len(intArray) > 0 {
		remainder = 8
	}
	bitArray := make([]byte, byteArrayLength)
	// Deal with higher bits
	for i := 0; i < remainder; i++ {
		if intArray[i] == 1 {
			bitArray[0] |= (1 << (remainder - 1 - i))
		}
	}
	// Deal with lower bits
	tempIntArray := intArray[remainder:]
	for i, val := range tempIntArray {
		byteIndex := i/8 + 1
		bitIndex := 7 - (i % 8)
		if val == 1 {
			bitArray[byteIndex] |= (1 << bitIndex)
		}
	}
	return bitArray
}

func ConvertBitArrayToIntArray(bitArray []byte, totalBits int) []int {
	intArray := make([]int, totalBits)

	for i := 0; i < totalBits; i++ {
		byteIndex := i / 8
		bitIndex := 7 - (i % 8)
		if byteIndex < len(bitArray) && (bitArray[byteIndex]&(1<<bitIndex)) != 0 {
			intArray[i] = 1
		} else {
			intArray[i] = 0
		}
	}
	return intArray
}

func CRC8(data []int) []int {
	table := crc8.MakeTable(crc8.CRC8)
	byteData := ConvertIntArrayToBitArray(data)
	crc := crc8.Checksum(byteData, table)
	intCRC := ConvertBitArrayToIntArray([]byte{crc}, 8)
	return intCRC
}

func SavePreambleToFile(filename string, preamble []jack.AudioSample) error {
	// 打开或创建文件
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 将preamble数组写入文件
	for _, sample := range preamble {
		// 将AudioSample类型转换为字符串
		err := writer.Write([]string{strconv.FormatFloat(float64(sample), 'f', -1, 64)})
		if err != nil {
			return err
		}
	}

	return nil
}
