package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"

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

// Test: Compare INPUT.txt and matlab/decode_output.txt
func Compare() {
	// Open the  input file
	input_file, err := os.Open("compare/INPUT.txt")
	if err != nil {
		panic(err)
	}
	defer input_file.Close()

	output_file, err := os.Open("compare/OUTPUT.txt")
	if err != nil {
		panic(err)
	}
	defer output_file.Close()

	data_input := make([]int, 0, 10000)
	// Read the file
	scanner := bufio.NewScanner(input_file)
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
			data_input = append(data_input, value)
		}
	}

	data_output := make([]int, 0, 10000)
	// Read the file
	out_scanner := bufio.NewScanner(output_file)
	for out_scanner.Scan() {
		line := out_scanner.Text()
		numbers := strings.Fields(line)
		for _, num := range numbers {
			// 将字符串转换为整数
			value, err := strconv.Atoi(num)
			if err != nil {
				fmt.Println("Error converting string to int:", err)
				continue
			}
			data_output = append(data_output, value)
		}
	}
	error_count := 0
	for i := 0; i < len(data_input); i++ {
		if data_input[i] != data_output[i] {
			error_count++
		}
		// } else {
		// 	if error_count != 0 {
		// 		fmt.Printf("%d ", error_count)
		// 	}
		// 	error_count = 0
		// }
	}
	// Calculate error rate
	fmt.Println("Total error is: ", error_count, " bit")
	fmt.Println("Error rate is: ", float32(error_count/10000))
}
