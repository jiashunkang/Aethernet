package shared

import (
	"container/list"
	"fmt"
	"log"
	"sync"
)

// Enable easy read/write data for MAC layer
type IOHelper struct {
	sendBuffer     list.List
	recvBuffer     list.List
	sendBufferLock sync.Mutex
	recvBufferLock sync.Mutex
	hasDataBool    bool
	hasDataLock    sync.Mutex
	ipByteChan     chan []byte
}

type bufferSlot struct {
	start int
	size  int
	Data  []int
}

// Create a new IOHelper
func NewIOHelper(ipByteChan chan []byte) *IOHelper {
	return &IOHelper{
		sendBuffer:     list.List{},
		recvBuffer:     list.List{},
		sendBufferLock: sync.Mutex{},
		recvBufferLock: sync.Mutex{},
		hasDataBool:    false,
		hasDataLock:    sync.Mutex{},
		ipByteChan:     ipByteChan,
	}
}

func (io *IOHelper) ReadData(size int) (data []int) {
	io.sendBufferLock.Lock()
	defer io.sendBufferLock.Unlock()

	if io.sendBuffer.Len() == 0 {
		fmt.Println("Error: No data in buffer")
		return nil
	}
	slot := io.sendBuffer.Front().Value.(*bufferSlot)
	if slot.start+size > slot.size {
		io.sendBuffer.Remove(io.sendBuffer.Front())
		if io.sendBuffer.Len() == 0 {
			io.hasDataLock.Lock()
			io.hasDataBool = false
			io.hasDataLock.Unlock()
		}
		return slot.Data[slot.start:]
	}
	data = slot.Data[slot.start : slot.start+size]
	slot.start += size
	if slot.start == slot.size {
		io.sendBuffer.Remove(io.sendBuffer.Front())
		if io.sendBuffer.Len() == 0 {
			io.hasDataLock.Lock()
			io.hasDataBool = false
			io.hasDataLock.Unlock()
		}
	}
	return
}

// This is for mac layer to write data
func (io *IOHelper) WriteData(data []int) {
	// Examine the first 2 bit of the data
	// 00,  the start of a IP frame
	// 10,  the middle of a IP frame
	// 11,  end of the last IP frame
	// 01,  a frame by itself

	if len(data) < 2 {
		fmt.Println("Error: Data size is less than 2")
		return
	}
	if data[0] == 0 && data[1] == 0 {
		// Start of a IP frame
		slot_data := make([]int, 0, 8*DATA_SIZE)
		slot_data = append(slot_data, data[2:]...)
		io.recvBufferLock.Lock()
		io.recvBuffer.PushBack(&bufferSlot{0, len(data), slot_data})
		io.recvBufferLock.Unlock()
	} else if data[0] == 1 && data[1] == 0 {
		// Middle of a IP frame
		io.recvBufferLock.Lock()
		slot := io.recvBuffer.Back().Value.(*bufferSlot)
		slot.Data = append(slot.Data, data[2:]...)
		io.recvBufferLock.Unlock()
	} else if data[0] == 1 && data[1] == 1 {
		// End of the last IP frame
		io.recvBufferLock.Lock()
		slot := io.recvBuffer.Back().Value.(*bufferSlot)
		slot.Data = append(slot.Data, data[2:]...)
		io.recvBuffer.Remove(io.recvBuffer.Front())
		io.recvBufferLock.Unlock()
		io.ipByteChan <- ConvertIntArrayToBitArray(slot.Data)
	} else {
		// A frame by itself
		io.ipByteChan <- ConvertIntArrayToBitArray(data[2:])
	}
}

// Write IP data to buffer, Mac will fetch and send it later
func (io *IOHelper) IPWriteBuffer(data []byte) {
	if len(data) > MTU {
		// throw error
		log.Fatalf("Error: IP data size exceeds MTU")
	}
	data_out := make([]int, 0, 8*len(data)+8)
	FrameSize := (DATA_SIZE - 2)
	// separate to (DATA_SIZE-2) bits per mac frame
	dataBits := ConvertBitArrayToIntArray(data, 8*len(data))
	if len(dataBits) <= DATA_SIZE-2 {
		data_out = append(data_out, []int{0, 1}...)
		data_out = append(data_out, dataBits...)
	} else {
		// calculate how many frames needed
		frameNum := (len(dataBits)-1)/(DATA_SIZE-2) + 1
		// First frame
		data_out = append(data_out, []int{0, 0}...)
		data_out = append(data_out, dataBits[0:FrameSize]...)
		// Middle frames
		for i := 1; i < frameNum-1; i++ {
			data_out = append(data_out, []int{1, 0}...)
			data_out = append(data_out, dataBits[i*FrameSize:(i+1)*FrameSize]...)
		}
		// Last frame
		data_out = append(data_out, []int{1, 1}...)
		data_out = append(data_out, dataBits[(frameNum-1)*FrameSize:]...)
	}
	io.sendBufferLock.Lock()
	slot := &bufferSlot{0, len(data_out), data_out}
	io.sendBuffer.PushBack(slot)
	io.sendBufferLock.Unlock()

	io.hasDataLock.Lock()
	if !io.hasDataBool {
		io.hasDataBool = true
	}
	io.hasDataLock.Unlock()
}

func (io *IOHelper) HasData() bool {
	io.hasDataLock.Lock()
	defer io.hasDataLock.Unlock()
	return io.hasDataBool
}
