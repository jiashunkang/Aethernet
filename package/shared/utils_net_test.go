package shared

import (
	"fmt"
	"net"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestGetHotSpotIP(t *testing.T) {
	ip := GetHotSpotIP()
	if ip == "" {
		t.Error("GetHotSpotIP failed")
	}
	fmt.Println("Hotspot IP:", ip)
}

func TestGetDeviceNameByIp(t *testing.T) {
	ip := GetHotSpotIP()
	if ip == "" {
		t.Error("GetHotSpotIP failed")
	}
	name := GetDeviceNameByIp(ip)
	if name == "" {
		t.Error("GetDeviceNameByIp failed")
	}
	fmt.Println("Device name:", name)
}

func TestSubnet(t *testing.T) {
	NetIP := *GetSubnetMaskByIP("192.168.137.1")
	fmt.Println(GetSubnetMaskByIP("192.168.137.1"))
	fmt.Println(NetIP.Contains(net.ParseIP("192.168.137.123")))
}

func TestPacketDecode(t *testing.T) {
	packetData, _ := CreateICMPv4Packet("192.168.137.212", "192.168.137.1", 23451, 0)
	packet := gopacket.NewPacket(packetData, layers.LayerTypeIPv4, gopacket.Default)

	if ipv4Layer := packet.Layer(layers.LayerTypeIPv4); ipv4Layer != nil {
		ipv4, _ := ipv4Layer.(*layers.IPv4)
		// Check if belongs to aether
		if ipv4.SrcIP.String() == "192.168.137.212" && ipv4.DstIP.String() == "192.168.137.1" {
			// Serialize again (remove ether header, keep ipv4 layer only)
			buffer := gopacket.NewSerializeBuffer()
			serializeOptions := gopacket.SerializeOptions{}
			err := gopacket.SerializeLayers(buffer, serializeOptions, ipv4)
			if err != nil {
				fmt.Println("Error serializing packet:", err)
			}
			// Decode the serialize buffer
			echoPacket := gopacket.NewPacket(buffer.Bytes(), layers.LayerTypeIPv4, gopacket.Default)
			// Extract IPv4 layer
			newicmpLayer := echoPacket.Layer(layers.LayerTypeICMPv4)
			if newicmpLayer == nil {
				fmt.Println("No ICMP layer found")
			} else {
				fmt.Println("ICMP layer found")
			}
		}
	}

}
