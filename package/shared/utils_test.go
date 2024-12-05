// main_test.go
package shared

import (
	"bytes"
	"fmt"
	"testing"
)

// TestConvertIntArrayToBitArray 测试 ConvertIntArrayToBitArray 函数
func TestConvertIntArrayToBitArray(t *testing.T) {
	tests := []struct {
		name     string
		intArray []int
		expected []byte
	}{
		{
			name:     "Empty array",
			intArray: []int{},
			expected: []byte{},
		},
		{
			name:     "All zeros",
			intArray: []int{0, 0, 0, 0, 0, 0, 0, 0},
			expected: []byte{0x00},
		},
		{
			name:     "All ones",
			intArray: []int{1, 1, 1, 1, 1, 1, 1, 1},
			expected: []byte{0xFF},
		},
		{
			name:     "Mixed bits less than a byte",
			intArray: []int{1, 0, 1},
			expected: []byte{0x05}, // 00000101
		},
		{
			name:     "Mixed bits more than a byte",
			intArray: []int{1, 0, 1, 1, 0, 0, 1, 0},
			expected: []byte{0xB2}, // 10110010
		},
		{
			name:     "Multiple bytes",
			intArray: []int{1, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0, 0, 1, 0, 1, 1},
			expected: []byte{0xB2, 0xCB}, // 10110010 00110101
		},
		{
			name:     "Remainder bits set",
			intArray: []int{1, 1, 1}, // remainder = 3
			expected: []byte{0x07},   // 00000111
		},
		{
			name:     "Remainder and full bytes",
			intArray: []int{1, 1, 0, 1, 1, 1, 0, 0, 1, 0, 1},
			expected: []byte{0x06, 0xE5}, // 00000110 11100101
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertIntArrayToBitArray(tt.intArray)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("ConvertIntArrayToBitArray(%v) = %v; want %v", tt.intArray, result, tt.expected)
			}
		})
	}
}

func TestCRC8(t *testing.T) {
	//data := []int{1, 0, 0, 0, 1, 1, 1, 1, 0, 1, 1, 0, 0, 1, 0, 0, 0, 1, 1, 1, 0, 0, 1, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 1, 1, 1, 1, 0, 1, 0, 0, 0, 1, 1, 1, 1, 0, 0, 1, 1, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0, 1, 1, 0, 0, 1, 0, 0, 1, 1, 0, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 1, 1, 1, 1, 0, 1, 0, 0, 0, 1, 1, 1, 1, 0, 0, 1, 1, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0, 1, 1, 0, 1, 0, 1, 0, 0, 1, 0, 0, 1, 1, 0, 1, 0, 0, 1, 1, 1, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 1, 0, 1, 1, 1, 1, 1, 0, 1, 1, 0, 0, 1, 0, 0, 0, 0, 0, 1, 1}
	// check if crc is correct
	// 1,1,1,0,1,0,0,0
	//result := CRC8(data)

	//fmt.Printf("CRC8 result for data %v: %v\n", data, result)
}

// Test: Compare INPUT.txt and matlab/decode_output.txt
func TestDecode(t *testing.T) {
	CompareBin()
}

func TestCRCAffectData(t *testing.T) {
	// data = 1,1,1,0,1,0,0,0
	data := []int{1, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 1, 1}
	data_copy := make([]int, len(data))
	copy(data_copy, data)
	_ = CRC8(data)
	if !isEqual(data, data_copy) {
		t.Errorf("CRC8 function should not affect the data")
	}
	fmt.Println("Data after CRC8:", data)

}

func TestIntToBinaryArray1(t *testing.T) {
	num := 3
	result := IntToBinaryArray(num)
	expected := []int{0, 0, 0, 0, 0, 0, 0, 1, 1}
	if !isEqual(result, expected) {
		t.Errorf("IntToBinaryArray(%d) = %v; want %v", num, result, expected)
	}
}

func TestIntToBinaryArray2(t *testing.T) {
	num := 104
	result := IntToBinaryArray(num)
	expected := []int{0, 0, 1, 1, 0, 1, 0, 0, 0}
	if !isEqual(result, expected) {
		t.Errorf("IntToBinaryArray(%d) = %v; want %v", num, result, expected)
	}
}

func TestBinaryToInt1(t *testing.T) {
	binary := []int{1, 0, 0, 0, 0, 0, 0, 0, 1}
	result := BinaryArrayToInt(binary)
	expected := 256 + 1
	if result != expected {
		t.Errorf("BinaryToInt(%v) = %d; want %d", binary, result, expected)
	}
}

func TestLessThan(t *testing.T) {
	fmt.Println(LessThan(11, 2) == true)
	fmt.Println(LessThan(11, 1) == true)
	fmt.Println(LessThan(11, 0) == true)
	fmt.Println(LessThan(11, 15) == true)
	fmt.Println(LessThan(11, 14) == true)
	fmt.Println(LessThan(11, 13) == true)
	fmt.Println(LessThan(11, 12) == true)
	fmt.Println(LessThan(11, 11) == false)
	fmt.Println(LessThan(11, 10) == false)
	fmt.Println(LessThan(11, 9) == false)
	fmt.Println(LessThan(11, 8) == false)
	fmt.Println(LessThan(11, 7) == false)
	fmt.Println(LessThan(11, 6) == false)
	fmt.Println(LessThan(11, 5) == false)
	fmt.Println(LessThan(11, 4) == false)
	fmt.Println(LessThan(11, 3) == true)
}

func TestGreaterThan(t *testing.T) {
	fmt.Println(GreaterThan(11, 2) == true)
	fmt.Println(GreaterThan(11, 1) == true)
	fmt.Println(GreaterThan(11, 0) == true)
	fmt.Println(GreaterThan(11, 15) == true)
	fmt.Println(GreaterThan(11, 14) == true)
	fmt.Println(GreaterThan(11, 13) == true)
	fmt.Println(GreaterThan(11, 12) == true)
	fmt.Println(GreaterThan(11, 11) == true)
	fmt.Println(GreaterThan(11, 10) == false)
	fmt.Println(GreaterThan(11, 9) == false)
	fmt.Println(GreaterThan(11, 8) == false)
	fmt.Println(GreaterThan(11, 7) == false)
	fmt.Println(GreaterThan(11, 6) == false)
	fmt.Println(GreaterThan(11, 5) == false)
	fmt.Println(GreaterThan(11, 4) == false)
	fmt.Println(GreaterThan(11, 3) == true)
}

func TestGenerateInputBin(t *testing.T) {
	GenerateInputBin()
}