package shared

import (
	"crypto/rand"
	"fmt"
	"log"
	"testing"
	"time"
)

// Test IOHelper creation
func TestNewIOHelper(t *testing.T) {
	ipByteChan := make(chan []byte, 1)
	io := NewIOHelper(ipByteChan)
	if io == nil {
		t.Fatal("Failed to create IOHelper")
	}
}

// Test Send_IP
func TestSendIP(t *testing.T) {
	for i := 25; i <= 200; i += 25 {
		time.Sleep(time.Second)
		fmt.Println("TestSendIP with", i, "bytes")
		ipByteChan := make(chan []byte, 4)
		io := NewIOHelper(ipByteChan)
		// Test data with 40,60,80,100,120,140,160,180 byte cases
		arr := make([]byte, i)
		for i := range arr {
			arr[i] = 0x32
		}
		io.IPWriteBuffer(arr)
		// Check state
		if !io.hasDataBool {
			t.Fatal("hasDataBool should be true")
		}
		slot := io.sendBuffer.Front().Value.(*bufferSlot)
		if slot.size != len(arr)*8+(1+(len(arr)-1)/50)*2 {
			t.Fatalf("slot size should be %d, got %d", len(arr)*8+(1+(len(arr)-1)/50)*2, slot.size)
		}
		if len(slot.Data) != len(arr)*8+(1+(len(arr)-1)/50)*2 {
			t.Fatalf("sendBuffer length should be %d, got %d", len(arr)*8+(1+(len(arr)-1)/50)*2, len(slot.Data))
		}
		// Check data
		if len(slot.Data) <= 402 {
			if slot.Data[0] != 0 || slot.Data[1] != 1 {
				t.Fatal("Only one slot should be 0,1")
			}
		} else if len(slot.Data) > 402 && len(slot.Data) <= 804 {
			if slot.Data[0] != 0 || slot.Data[1] != 0 {
				t.Fatal("First slot should be 0,2")
			}
			if slot.Data[402] != 1 || slot.Data[403] != 1 {
				t.Fatal("Final slot should be 1,1")
			}
		} else if len(slot.Data) > 804 && len(slot.Data) <= 402*3 {
			if slot.Data[0] != 0 || slot.Data[1] != 0 {
				t.Fatal("First slot should be 0,2")
			}
			if slot.Data[402] != 1 || slot.Data[403] != 0 {
				fmt.Println(slot.Data[402], slot.Data[403])
				t.Fatal("Second slot should be 1,0")
			}
			if slot.Data[804] != 1 || slot.Data[805] != 1 {
				t.Fatal("Final slot should be 1,1")
			}
		} else if len(slot.Data) > 402*3 && len(slot.Data) <= 402*4 {
			if slot.Data[0] != 0 || slot.Data[1] != 0 {
				t.Fatal("First slot should be 0,2")
			}
			if slot.Data[402] != 1 || slot.Data[403] != 0 {
				t.Fatal("Second slot should be 1,0")
			}
			if slot.Data[804] != 1 || slot.Data[805] != 0 {
				t.Fatal("Third slot should be 1,0")
			}
			if slot.Data[402*3] != 1 || slot.Data[402*3+1] != 1 {
				t.Fatal("Final slot should be 1,1")
			}
		}

	}
}

func TestMultipleSendIP(t *testing.T) {
	ipByteChan := make(chan []byte, 4)
	io := NewIOHelper(ipByteChan)
	for i := 25; i <= 200; i += 25 {
		fmt.Println("TestSendIP with", i, "bytes")
		// Test data with 40,60,80,100,120,140,160,180 byte cases
		arr := make([]byte, i)
		for j := range arr {
			arr[j] = 0x32
		}
		io.IPWriteBuffer(arr)
		// Check state
		if !io.hasDataBool {
			t.Fatal("hasDataBool should be true")
		}
	}
	if io.sendBuffer.Len() != 8 {
		t.Fatalf("sendBuffer length should be 8, got %d", io.sendBuffer.Len())
	}
}

// Test ReadData
func TestSingleReadData(t *testing.T) {
	for i := 23; i <= 200; i += 25 {
		ipByteChan := make(chan []byte, 1)
		io := NewIOHelper(ipByteChan)
		arr := make([]byte, i)
		for i := range arr {
			arr[i] = 0x32
		}
		arr_size := len(arr)*8 + (1+(len(arr)-1)/50)*2
		io.IPWriteBuffer(arr)
		count := 0
		size := 0
		for io.HasData() {
			buf := io.ReadData(DATA_SIZE)
			size += len(buf)
			count++
		}
		log.Println("count", count, "bytes", len(arr))
		if size != arr_size {
			t.Fatalf("size should be %d, got %d", arr_size, size)
		}
	}
}
func TestMultipleReadData(t *testing.T) {
	byteCount := 0
	ipByteChan := make(chan []byte, 1)
	io := NewIOHelper(ipByteChan)
	for i := 25; i <= 200; i += 25 {
		byteCount += i
		arr := make([]byte, i)
		for i := range arr {
			arr[i] = 0x32
		}
		io.IPWriteBuffer(arr)
	}
	size := 0
	for io.HasData() {
		buf := io.ReadData(DATA_SIZE)
		fmt.Println("buf", len(buf))
		size += len(buf)
		size -= 2
	}

	if size != byteCount*8 {
		t.Fatalf("size should be %d, got %d", byteCount*8, size)
	}

}

func TestWriteData(t *testing.T) {
	for i := 25; i <= 200; i += 25 {
		ipByteChan := make(chan []byte, 4)
		io := NewIOHelper(ipByteChan)
		arr := make([]byte, 200)
		// Generate random byte
		_, err := rand.Read(arr)
		if err != nil {
			t.Fatal("Failed to generate random byte")
		}
		io.IPWriteBuffer(arr)
		for io.HasData() {
			buf := io.ReadData(DATA_SIZE)
			io.WriteData(buf)
		}
		ipByteStream := <-ipByteChan
		// Compare ipByteStream and arr
		if len(ipByteStream) != len(arr) {
			t.Fatalf("ipByteStream length should be %d, got %d", len(arr), len(ipByteStream))
		}
		for i := range arr {
			if arr[i] != ipByteStream[i] {
				t.Fatalf("arr[%d] should be %d, got %d", i, arr[i], ipByteStream[i])
			}
		}
	}
}
