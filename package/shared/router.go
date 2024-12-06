package shared

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type Router struct {
	io           *IOHelper   // IOHelper
	aetherIPChan chan []byte // IP message in byte from acoustic channel
	AetherIP     net.IP      // self defined acoustic net IP
	FT           []ForwardingTableSlot
}

type ForwardingTableSlot struct {
	SubNet      *net.IPNet
	InterfaceIP net.IP
	NetType     string
}

func NewRouter(aetherIP string, io *IOHelper, aetherchan chan []byte) *Router {
	r := &Router{
		AetherIP:     net.ParseIP(aetherIP),
		io:           io,
		aetherIPChan: aetherchan,
	}
	// Construct static forwarding table
	r.FT = make([]ForwardingTableSlot, 3)
	// 1. Athernet interface
	r.FT[0].InterfaceIP = r.AetherIP
	r.FT[0].NetType = "Aethernet"
	_, r.FT[0].SubNet, _ = net.ParseCIDR(r.AetherIP.String() + "/24")
	// 2. Hotspot interface for phone
	r.FT[1] = ForwardingTableSlot{
		SubNet:      GetSubnetMaskByIP(GetHotSpotIP()),
		InterfaceIP: net.ParseIP(GetHotSpotIP()),
		NetType:     "Ethernet",
	}
	// 3. Ethernet interface for Outward Connection
	r.FT[2] = ForwardingTableSlot{
		SubNet:      GetSubnetMaskByIP(GetOutboundIP()),
		InterfaceIP: net.ParseIP(GetOutboundIP()),
		NetType:     "Ethernet",
	}
	fmt.Println("Forwarding Table:", r.FT)
	return r
}

func (r *Router) ListenEther(slot ForwardingTableSlot) {
	// Listen on the interface
	deviceName := GetDeviceNameByIp(slot.InterfaceIP.String())
	fmt.Println("Listening on interface:", slot.InterfaceIP, "  ", deviceName)
	handle, err := pcap.OpenLive(deviceName, 1600, true, pcap.BlockForever)
	if err != nil {
		fmt.Println("Error opening interface:", err)
		return
	}
	defer handle.Close()
	err = handle.SetBPFFilter("icmp")
	if err != nil {
		fmt.Println("Error setting BPF filter:", err)
		return
	}
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if ipv4Layer := packet.Layer(layers.LayerTypeIPv4); ipv4Layer != nil {
			ipv4, _ := ipv4Layer.(*layers.IPv4)
			// Check if belongs to aether
			if r.FT[0].SubNet.Contains(ipv4.DstIP) {
				// Serialize again (remove ether header, keep ipv4 layer only)
				buffer := gopacket.NewSerializeBuffer()
				serializeOptions := gopacket.SerializeOptions{}
				err := gopacket.SerializeLayers(buffer, serializeOptions, ipv4, gopacket.Payload(ipv4.Payload))
				if err != nil {
					fmt.Println("Error serializing packet:", err)
					continue
				}
				r.io.IPWriteBuffer(buffer.Bytes())
			}
		}
	}
}

func (r *Router) ListenAether() {
	for {
		data := <-r.aetherIPChan
		packet := gopacket.NewPacket(data, layers.LayerTypeIPv4, gopacket.Default)
		if ipv4Layer := packet.Layer(layers.LayerTypeIPv4); ipv4Layer != nil {
			ipv4, _ := ipv4Layer.(*layers.IPv4)
			// Check if belongs to ethernet
			founded := false
			for _, slot := range r.FT[1:] {
				if slot.SubNet.Contains(ipv4.DstIP) {
					founded = true
					// Serialize again (remove ether header, keep ipv4 layer only)
					buffer := gopacket.NewSerializeBuffer()
					serializeOptions := gopacket.SerializeOptions{}
					etherLayer := &layers.Ethernet{
						EthernetType: layers.EthernetTypeIPv4,
						SrcMAC:       net.HardwareAddr{0x00, 0x0c, 0x29, 0x57, 0xd7, 0x1c}, // self defined MAC... I dont know how to configure mac...
						DstMAC:       net.HardwareAddr{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
					}
					err := gopacket.SerializeLayers(buffer, serializeOptions, etherLayer, ipv4, gopacket.Payload(ipv4.Payload))
					if err != nil {
						fmt.Println("Error serializing packet:", err)
						continue
					}
					// Create new handle to write packet
					deviceName := GetDeviceNameByIp(slot.InterfaceIP.String())
					handle, err := pcap.OpenLive(deviceName, 1600, true, pcap.BlockForever)
					if err != nil {
						fmt.Println("Error opening interface:", err)
						return
					}
					err = handle.WritePacketData(buffer.Bytes())
					if err != nil {
						fmt.Println("Error writing packet:", err)
						handle.Close()
						continue
					}
					handle.Close()
				}
			}
			if !founded {
				fmt.Println("No forwarding table found for IP:", ipv4.DstIP)
			}
			// Check if send to this router through aether
			if r.FT[0].SubNet.Contains(ipv4.DstIP) {
				if ipv4.DstIP.String() == r.AetherIP.String() {
					// Respond ICMP Echo Reply if needed
					if icmpLayer := packet.Layer(layers.LayerTypeICMPv4); icmpLayer != nil {
						icmp, _ := icmpLayer.(*layers.ICMPv4)
						if icmp.TypeCode == layers.ICMPv4TypeEchoRequest {
							// Modify the packet
							originalSrcIP := ipv4.SrcIP
							originalDstIP := ipv4.DstIP
							ipv4.SrcIP = originalDstIP
							ipv4.DstIP = originalSrcIP
							icmp.TypeCode = layers.ICMPv4TypeEchoReply
							buffer := gopacket.NewSerializeBuffer()
							options := gopacket.SerializeOptions{}
							err := gopacket.SerializePacket(buffer, options, packet)
							if err != nil {
								fmt.Println("Failed to serialize packet:", err)
							}
							fmt.Println("Responding ICMP Echo Reply")
							r.io.IPWriteBuffer(buffer.Bytes())
						}
					}
				}
			}
		}

	}
}

func (r *Router) Start() {
	go r.ListenAether()
	for _, slot := range r.FT[1:] {
		go r.ListenEther(slot)
	}
}
