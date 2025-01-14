package shared

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/jackpal/gateway"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// Get the IP if this computer function as a hotspot
func GetHotSpotIP() string {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal("Error finding devices: ", err)
	}
	for _, device := range devices {
		if device.Addresses != nil {
			for _, address := range device.Addresses {
				if strings.HasSuffix(address.IP.String(), ".1") && !strings.HasSuffix(address.IP.String(), "0.1") && strings.Contains(device.Description, "Wi-Fi") {
					return address.IP.String()
				}
			}
		}
	}
	return ""
}

// Get device name coresponding to the IP
func GetDeviceNameByIp(ip string) string {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal("Error finding devices: ", err)
	}
	for _, device := range devices {
		if device.Addresses != nil {
			for _, address := range device.Addresses {
				if address.IP.String() == ip {
					return device.Name
				}
			}
		}
	}
	return ""
}

// Get the subnet mask by IP
func GetSubnetMaskByIP(ip string) *net.IPNet {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			log.Fatal(err)
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			if ipNet.IP.To4() != nil {
				if ipNet.IP.String() == ip {
					return ipNet
				}
			}
		}
	}
	return nil
}

// Get Default Outbound IP ----- ref: stackoverflow
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func GetInterfaceMACByIP(ip string) string {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			log.Fatal(err)
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			if ipNet.IP.To4() != nil {
				if ipNet.IP.String() == ip {
					return iface.HardwareAddr.String()
				}
			}
		}
	}
	return ""
}

func decodeGBK(input []byte) (string, error) {
	decoder := simplifiedchinese.GBK.NewDecoder()
	output, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(input), decoder))
	return string(output), err
}

func GetMACAddressByArp(ip string) string {
	// 执行系统 arp 命令
	out, err := exec.Command("arp", "-a", ip).Output()
	if err != nil {
		return ""
	}
	outString, _ := decodeGBK(out)
	lines := strings.Fields(outString)
	for i, line := range lines {
		if strings.Contains(line, ip) {
			return lines[i+1]
		}
	}

	return ""
}

func GetOutBoundRouterIP() string {
	gateway, _ := gateway.DiscoverGateway()
	return gateway.String()
}

func GetDomainIPLookup(domain string) string {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return ""
	}
	return ips[0].String()
}

func GetDomainIP(domain string) string {
	// DNS 服务器地址和端口
	dnsServer := "1.1.1.1:53"

	// 创建 UDP 连接
	conn, err := net.Dial("udp", dnsServer)
	if err != nil {
		fmt.Println("Failed to create UDP connection: ", err)
	}
	defer conn.Close()

	dns := layers.DNS{
		ID:     12345,
		QR:     false,
		OpCode: 0,
		RD:     true,
		AA:     true,
		Questions: []layers.DNSQuestion{
			{
				Name:  []byte(domain),
				Type:  layers.DNSTypeA,
				Class: layers.DNSClassIN,
			},
		},
	}

	// 将 DNS 查询包序列化为字节
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
	if err := dns.SerializeTo(buf, opts); err != nil {
		fmt.Println("Failed to serialize DNS query: ", err)
	}

	// 发送 DNS 查询包
	if _, err := conn.Write(buf.Bytes()); err != nil {
		fmt.Println("Failed to send DNS query: ", err)
	}
	fmt.Printf("Sent DNS query for %s to %s\n", domain, dnsServer)

	// 设置接收超时时间
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// 接收响应
	resp := make([]byte, 512)
	n, err := conn.Read(resp)
	if err != nil {
		fmt.Println("Failed to read DNS response: ", err)
	}

	packet := gopacket.NewPacket(resp[:n], layers.LayerTypeDNS, gopacket.Default)
	dnsLayer := packet.Layer(layers.LayerTypeDNS)
	if dnsLayer == nil {
		fmt.Println("No DNS layer found in response")
	}

	dnsResp, _ := dnsLayer.(*layers.DNS)
	fmt.Printf("Received DNS response with %d answers:\n", len(dnsResp.Answers))

	for _, answer := range dnsResp.Answers {
		fmt.Printf("Name: %s, Type: %v, IP: %v\n", string(answer.Name), answer.Type, answer.IP)
	}

	return dnsResp.Answers[0].IP.String()
}
