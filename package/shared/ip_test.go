package shared

import (
	"fmt"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestCreateAndModifyPacket(t *testing.T) {
	hostIP := "172.172.8.233"
	destIP := "172.172.8.128"
	seqNum := 1
	id := 1
	sendBuffer, err := CreateICMPv4Packet(hostIP, destIP, id, seqNum)
	if err != nil {
		t.Fatalf("Error creating ICMPv4 packet: %v", err)
	}
	if true {
		sendPacket := gopacket.NewPacket(sendBuffer, layers.LayerTypeIPv4, gopacket.Default)
		icmpLayer := sendPacket.Layer(layers.LayerTypeICMPv4)
		ipv4Layer := sendPacket.Layer(layers.LayerTypeIPv4)
		if icmpLayer != nil {
			icmp, _ := icmpLayer.(*layers.ICMPv4)
			ipv4, _ := ipv4Layer.(*layers.IPv4)
			fmt.Println("icmp.TypeCode.Type():", icmp.TypeCode.Type())
			fmt.Println("icmp.TypeCode.Code():", icmp.TypeCode.Code())
			if icmp.TypeCode.Type() == layers.ICMPv4TypeEchoReply {
				// Access IPv4 fields
				fmt.Println("Version:", ipv4.Version)
				fmt.Println("IHL:", ipv4.IHL)
				fmt.Println("TOS:", ipv4.TOS)
				fmt.Println("Length:", ipv4.Length)
				fmt.Println("Id:", ipv4.Id)
				fmt.Println("Flags:", ipv4.Flags)
				fmt.Println("FragOffset:", ipv4.FragOffset)
				fmt.Println("TTL:", ipv4.TTL)
				fmt.Println("Protocol:", ipv4.Protocol)
				fmt.Println("Checksum:", ipv4.Checksum)
				fmt.Println("Source IP:", ipv4.SrcIP)
				fmt.Println("Destination IP:", ipv4.DstIP)
				fmt.Println("Options:", ipv4.Options)
				fmt.Println("Padding:", ipv4.Padding)
				// Access ICMP fields
				fmt.Println("---`Accessing ICMP fields")
				fmt.Println("Type:", icmp.TypeCode.Type())
				fmt.Println("Code:", icmp.TypeCode.Code())
				fmt.Println("Checksum:", icmp.Checksum)
				fmt.Println("Id:", icmp.Id)
				fmt.Println("Seq:", icmp.Seq)
			}
		}
	}
	receiveBuffer, _ := ModifyPacket(sendBuffer)
	if true {
		receivePacket := gopacket.NewPacket(receiveBuffer, layers.LayerTypeIPv4, gopacket.Default)
		icmpLayer := receivePacket.Layer(layers.LayerTypeICMPv4)
		ipv4Layer := receivePacket.Layer(layers.LayerTypeIPv4)
		if icmpLayer != nil {
			icmp, _ := icmpLayer.(*layers.ICMPv4)
			ipv4, _ := ipv4Layer.(*layers.IPv4)
			if icmp.TypeCode.Type() == layers.ICMPv4TypeEchoReply {
				// Access IPv4 fields
				fmt.Println("Version:", ipv4.Version)
				fmt.Println("IHL:", ipv4.IHL)
				fmt.Println("TOS:", ipv4.TOS)
				fmt.Println("Length:", ipv4.Length)
				fmt.Println("Id:", ipv4.Id)
				fmt.Println("Flags:", ipv4.Flags)
				fmt.Println("FragOffset:", ipv4.FragOffset)
				fmt.Println("TTL:", ipv4.TTL)
				fmt.Println("Protocol:", ipv4.Protocol)
				fmt.Println("Checksum:", ipv4.Checksum)
				fmt.Println("Source IP:", ipv4.SrcIP)
				fmt.Println("Destination IP:", ipv4.DstIP)
				fmt.Println("Options:", ipv4.Options)
				fmt.Println("Padding:", ipv4.Padding)
				// Access ICMP fields
				fmt.Println("---`Accessing ICMP fields")
				fmt.Println("Type:", icmp.TypeCode.Type())
				fmt.Println("Code:", icmp.TypeCode.Code())
				fmt.Println("Checksum:", icmp.Checksum)
				fmt.Println("Id:", icmp.Id)
				fmt.Println("Seq:", icmp.Seq)
			}
		}
	}
}
