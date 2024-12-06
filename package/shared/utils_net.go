package shared

import (
	"log"
	"net"
	"strings"

	"github.com/google/gopacket/pcap"
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
