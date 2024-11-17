package main

import (
	"fmt"
	"log"
	"os"
)

// Enable easy read/write data for MAC layer
type IOHelper struct {
	data_out    []int // all data read in from file
	data_in     []int // all data write to file
	bitOutStart int   // number of frames sent or received
	bitInStart  int   // number of frames sent or received
	hasData     bool  // whether there is data to transfer or not
}

// Create a new IOHelper
func NewIOHelper() *IOHelper {
	ioHelper := &IOHelper{}
	ioHelper.bitOutStart = 0
	ioHelper.bitInStart = 0
	ioHelper.data_in = make([]int, 0, 1000000)
	// Try to read from file
	ioHelper.hasData = true
	filePath := "compare/INPUT.bin"
	data, err := os.ReadFile(filePath)
	if err != nil {
		ioHelper.hasData = false
	}
	if ioHelper.hasData {
		ioHelper.data_out = ConvertBitArrayToIntArray(data, 8*len(data))
	}
	return ioHelper
}

func (io *IOHelper) ReadData(size int) (data []int) {
	if io.bitOutStart+size > len(io.data_out) {
		io.hasData = false
		return io.data_out[io.bitOutStart:]
	}
	data = io.data_out[io.bitOutStart : io.bitOutStart+size]
	io.bitOutStart += size
	io.hasData = io.bitOutStart < len(io.data_out)
	return
}

func (io *IOHelper) WriteData(data []int) bool {
	io.data_in = append(io.data_in, data...)
	io.bitInStart += len(data)
	fmt.Println("Write: ", io.bitInStart)
	return io.bitInStart == 6250*8 // collect enough data then return true
}

func (io *IOHelper) WriteDataToFile() {
	byteData := ConvertIntArrayToBitArray(io.data_in)
	file, err := os.Create("compare/OUTPUT.bin")
	if err != nil {
		log.Fatalf("Error creating OUTPUT.bin: %v", err)
		return
	}
	defer file.Close()
	_, err = file.Write(byteData)
	if err != nil {
		log.Fatalf("Error writing OUTPUT.bin: %v", err)
		return
	}
	fmt.Println("OUTPUT.bin written")
}
