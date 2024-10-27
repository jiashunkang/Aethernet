package main

// 常量定义
const (
	MaxBufferSize  = 1024
	ChirpStartFreq = 2000
	ChirpEndFreq   = 10000
	FS             = 48000 // Sample Frequency
	PreambleLength = 48    // Length of preamble signal
	FC             = 10000 // Carrier frequency
	SYNC_PARA      = 0.04
)

// PHY frame
type PHY_frame struct {
	Len         []int
	Crc         []int
	Phy_payload MAC_frame
}

//MAC frame
type MAC_frame struct {
	data []int // dest,src,type,mac_payload
}
