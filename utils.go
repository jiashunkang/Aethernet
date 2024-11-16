package main

import (
	"bufio"
	"crypto/rand"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/sigurn/crc8"
	"github.com/xthexder/go-jack"
)

func GenerateInputBin() {
	file, err := os.Create("compare/INPUT.bin")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 定义要生成的字节数
	const byteCount = 6250

	// 创建一个切片用于存储随机字节
	data := make([]byte, byteCount)

	// 使用 crypto/rand 生成随机字节
	_, err = rand.Read(data)
	if err != nil {
		panic(err) // 处理随机数生成错误
	}
	// 将随机字节写入文件
	_, err = file.Write(data)
	if err != nil {
		panic(err) // 处理写入错误
	}
	fmt.Println("6250 bytes of random data written to INPUT.bin")
}

func WriteOutputTxt(data []int) {
	file, err := os.Create("compare/OUTPUT.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for _, bit := range data {
		file.WriteString(fmt.Sprint(bit))
		// add a space between every bit
		file.WriteString(" ")
	}
}

func ReadFromCsvFile(filename string) ([]jack.AudioSample, error) {
	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 创建 CSV reader
	reader := csv.NewReader(file)

	// 读取所有行
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// 创建一个 float32 切片来存储结果
	var result []jack.AudioSample

	// 遍历每一行的数据
	for _, record := range records {
		for _, value := range record {
			// 将字符串转换为 float32
			floatValue, err := strconv.ParseFloat(value, 32)
			if err != nil {
				return nil, err
			}
			result = append(result, jack.AudioSample(floatValue))
		}
	}

	return result, nil
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
	table := crc8.MakeTable(crc8.CRC8_MAXIM)
	byteData := ConvertIntArrayToBitArray(data)
	crc := crc8.Checksum(byteData, table)
	intCRC := ConvertBitArrayToIntArray([]byte{crc}, 8)
	return intCRC
}
func GenerateChirpPreamble(fstart, fend, fs, length int) []jack.AudioSample {
	// make a preamble array
	preamble := make([]jack.AudioSample, length)
	// Define the number of samples
	n := length
	time := make([]float64, n)
	dt := 1.0 / 48000.0 // Assuming a 48 kHz sample rate
	// Create the time vector t
	for i := range time {
		time[i] = float64(i) * dt
	}
	// Create the frequency profile f_p
	f_p := make([]float64, n)
	for i := 0; i < PreambleLength/2; i++ {
		f_p[i] = 4e3 + 8e3*float64(i)/240
		f_p[PreambleLength-1-i] = 4e3 + 8e3*float64(i)/240
	}
	// Compute the cumulative integral (omega) using the trapezoidal rule
	omega := make([]float64, n)
	omega[0] = 0
	for i := 1; i < n; i++ {
		omega[i] = omega[i-1] + 0.5*(f_p[i]+f_p[i-1])*2*math.Pi*dt
	}
	for i := range omega {
		preamble[i] = jack.AudioSample(math.Sin(omega[i]))
	}
	// // save preamble to file for matlab debugging
	// err := SavePreambleToFile("matlab/preamble.csv", preamble)
	// if err != nil {
	// 	fmt.Println("Error saving preamble:", err)
	// } else {
	// 	fmt.Println("Preamble saved to preamble.csv")
	// }
	return preamble
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
			fmt.Printf("%d ", i)
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
	fmt.Printf("Error rate is: %.2f\n", float32(error_count)/10000)
}

func CompareBin() {
	inputData, err := os.ReadFile("compare/INPUT2.bin")
	if err != nil {
		fmt.Println("Error reading file:", err)
	}
	outputData, err := os.ReadFile("compare/OUTPUT.bin")
	if err != nil {
		fmt.Println("Error reading file:", err)
	}
	byteError := 0
	for i, sample := range outputData {
		if inputData[i] != sample {
			byteError++
		}
	}
	bitError := 0
	bitInput := ConvertBitArrayToIntArray(inputData, 8*len(inputData))
	bitOutput := ConvertBitArrayToIntArray(outputData, 8*len(outputData))
	for i, sample := range bitOutput {
		if bitInput[i] != sample {
			bitError++
		}
	}
	fmt.Println("bitError is ", bitError)
	fmt.Println("byteError is ", byteError)
	fmt.Println("input length, output length", len(inputData), len(outputData))
}

// Input an int(0-500), return a 9-bit binary array
func IntToBinaryArray(num int) []int {
	binary := make([]int, 9)
	for i := 8; i >= 0; i-- {
		binary[i] = num % 2
		num /= 2
	}
	return binary
}

// Input a 9-bit binary array, return an int
func BinaryArrayToInt(binary []int) int {
	num := 0
	for i := 0; i < len(binary); i++ {
		num += binary[i] * int(math.Pow(2, float64(len(binary)-1-i)))
	}
	return num
}

func LessThan(a, b int) bool {
	// set a as a mid point and extend to find b
	// [a+1,a+8]
	for i := a + 1; i <= a+8; i++ {
		if i%16 == b {
			return true
		}
	}
	// (a-8,a]
	for i := a; i > a-8; i-- {
		if (i+16)%16 == b {
			return false
		}
	}
	return false
}

func GreaterThan(a, b int) bool {
	// set a as a mid point and extend to find b
	// [a,a+8)
	for i := a; i < a+8; i++ {
		if i%16 == b {
			return false
		}
	}
	// [a-8,a)
	for i := a; i >= a-8; i-- {
		if (i+16)%16 == b {
			return true
		}
	}
	return false
}

func GreaterEqual(a, b int) bool {
	return !LessThan(a, b)
}

func LessEqual(a, b int) bool {
	return !GreaterThan(a, b)
}

func Minus(a, b int) int {
	if GreaterThan(a, b) {
		return (a + 16 - b) % 16
	} else {
		return (b + 16 - a) % 16
	}
}
