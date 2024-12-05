package shared

import (
	"fmt"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type IPstruct struct {
	AetherIP     net.IP           // self defined acoustic net IP
	io           *IOHelper        // IOHelper
	mac          *MAC             // MAC
	aetherIPChan chan []byte      // IP message in byte from acoustic channel
	reqPingChan  chan RequestPing // Ping request channel
	EtherIP      net.IP           // Ethernet IP
}

type RequestPing struct {
	dstIP net.IP
	count int
}

func NewIP(aetherIP string, io *IOHelper, mac *MAC, aetherchan chan []byte) *IPstruct {
	ip := &IPstruct{
		AetherIP:     net.ParseIP(aetherIP),
		io:           io,
		mac:          mac,
		aetherIPChan: aetherchan,
	}
	ip.reqPingChan = make(chan RequestPing, 2)
	return ip
}
func (ip *IPstruct) Ping(IP string, count int) {
	fmt.Println("Ping", IP, "with", count, "packets")
	RequestPing := RequestPing{
		dstIP: net.ParseIP(IP),
		count: count,
	}
	ip.reqPingChan <- RequestPing
}

func (ip *IPstruct) Start() {
	for {
		// send a ping request
		select {
		case reqPing := <-ip.reqPingChan:
			dstIP := reqPing.dstIP
			count := reqPing.count
			for i := 0; i < count; i++ {
				buf, err := CreateICMPv4Packet(ip.AetherIP.String(), dstIP.String(), 23451+i, i)
				if err != nil {
					fmt.Println("Error creating ICMPv4 packet:", err)
					continue
				}
				ip.io.IPWriteBuffer(buf)
				startTime := time.Now()
				for {
					select {
					case echoPacketData := <-ip.aetherIPChan:
						elapse := time.Since(startTime)
						fmt.Println("Received a packet in", elapse)
						echoPacket := gopacket.NewPacket(echoPacketData, layers.LayerTypeIPv4, gopacket.Default)
						icmpLayer := echoPacket.Layer(layers.LayerTypeICMPv4)
						ipv4Layer := echoPacket.Layer(layers.LayerTypeIPv4)
						if icmpLayer != nil {
							icmp, _ := icmpLayer.(*layers.ICMPv4)
							ipv4, _ := ipv4Layer.(*layers.IPv4)
							if icmp.TypeCode == layers.ICMPv4TypeEchoReply {
								timeElapsed := time.Since(startTime)
								fmt.Println("Received ICMP Echo Reply from ", ipv4.DstIP, " in ", timeElapsed)
								// Access IPv4 fields
								fmt.Println("Length:", ipv4.Length)
								fmt.Println("Id:", ipv4.Id)
								fmt.Println("Flags:", ipv4.Flags)
								fmt.Println("TTL:", ipv4.TTL)
								fmt.Println("Protocol:", ipv4.Protocol)
								fmt.Println("Source IP:", ipv4.SrcIP)
								fmt.Println("Destination IP:", ipv4.DstIP)
							}
						}
					default:
						time.Sleep(1 * time.Millisecond)
					}

				}
			}
		default:

		}

		// idling, listening to the channel
		select {
		case requestPacketData := <-ip.aetherIPChan:
			fmt.Println("Received a packet")
			packet := gopacket.NewPacket(requestPacketData, layers.LayerTypeIPv4, gopacket.Default)
			icmpLayer := packet.Layer(layers.LayerTypeICMPv4)
			ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
			if icmpLayer != nil {
				icmp, _ := icmpLayer.(*layers.ICMPv4)
				ipv4, _ := ipv4Layer.(*layers.IPv4)
				if icmp.TypeCode == layers.ICMPv4TypeEchoRequest {
					fmt.Println("Received ICMP Echo Reply from ", ipv4.DstIP)
					// Access IPv4 fields
					fmt.Println("Length:", ipv4.Length)
					fmt.Println("Id:", ipv4.Id)
					fmt.Println("Flags:", ipv4.Flags)
					fmt.Println("TTL:", ipv4.TTL)
					fmt.Println("Protocol:", ipv4.Protocol)
					fmt.Println("Source IP:", ipv4.SrcIP)
					fmt.Println("Destination IP:", ipv4.DstIP)
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
					ip.io.IPWriteBuffer(buffer.Bytes())
				}
			}
		default:
		}

		time.Sleep(1 * time.Millisecond)
	}
}

func CreateICMPv4Packet(srcIP, dstIP string, id, seqNum int) ([]byte, error) {
	// 构建 IPv4 层
	ip := &layers.IPv4{
		Version:    4,
		IHL:        5,
		TOS:        0,
		Length:     0, // 将由 gopacket 自动计算
		Id:         23451,
		Flags:      layers.IPv4DontFragment,
		FragOffset: 0,
		TTL:        64,
		Protocol:   layers.IPProtocolICMPv4,
		Checksum:   0, // 将由 gopacket 自动计算
		SrcIP:      net.ParseIP(srcIP),
		DstIP:      net.ParseIP(dstIP),
		Options:    nil,
		Padding:    nil,
	}

	// 构建 ICMP 层（Echo Request）
	icmp := &layers.ICMPv4{
		TypeCode: layers.ICMPv4TypeEchoRequest,
		Id:       uint16(id),     // 标识符
		Seq:      uint16(seqNum), // 序列号
	}

	// 构建 ICMP 数据部分
	icmpPayload := []byte("Hello!")

	// 创建一个 gopacket 的 encoder 来组合数据包
	buffer := gopacket.NewSerializeBuffer()
	options := gopacket.SerializeOptions{}

	// 序列化 IP 包和 ICMP 包
	err := gopacket.SerializeLayers(buffer, options, ip, icmp, gopacket.Payload(icmpPayload))
	if err != nil {
		return nil, err
	}

	// 返回构建好的数据包
	return buffer.Bytes(), nil
}

func ModifyPacket(echoPacketData []byte) ([]byte, error) {
	// 解析原始数据包
	echoPacket := gopacket.NewPacket(echoPacketData, layers.LayerTypeIPv4, gopacket.Default)

	// 提取 IPv4 层
	ipv4Layer := echoPacket.Layer(layers.LayerTypeIPv4)
	if ipv4Layer == nil {
		return nil, fmt.Errorf("no IPv4 layer found")
	}
	ipv4, _ := ipv4Layer.(*layers.IPv4)

	// 提取 ICMP 层
	icmpLayer := echoPacket.Layer(layers.LayerTypeICMPv4)
	if icmpLayer == nil {
		return nil, fmt.Errorf("no ICMP layer found")
	}
	icmp, _ := icmpLayer.(*layers.ICMPv4)

	// 交换源和目标 IP 地址
	originalSrcIP := ipv4.SrcIP
	originalDstIP := ipv4.DstIP
	ipv4.SrcIP = originalDstIP
	ipv4.DstIP = originalSrcIP

	// 修改 ICMP 类型为 Echo Reply (type 0)
	icmp.TypeCode = layers.ICMPv4TypeEchoReply

	// 创建新的 buffer 来构造修改后的包
	buffer := gopacket.NewSerializeBuffer()
	options := gopacket.SerializeOptions{}

	// 序列化数据包
	err := gopacket.SerializePacket(buffer, options, echoPacket)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize packet: %v", err)
	}

	return buffer.Bytes(), nil
}
