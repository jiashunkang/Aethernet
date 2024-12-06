package shared

import (
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"strings"

	"github.com/google/gopacket/pcap"
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
				if strings.HasSuffix(address.IP.String(), ".1") && !strings.HasSuffix(address.IP.String(), "0.1") {
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
