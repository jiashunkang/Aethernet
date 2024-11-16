package main

// 常量定义
const (
	MaxBufferSize  = 1024
	ChirpStartFreq = 2000
	ChirpEndFreq   = 10000
	FS             = 48000 // Sample Frequency
	PreambleLength = 48    // Length of preamble signal
	FC             = 10000 // Carrier frequency
	SYNC_PARA      = 0.2
	DATA_SIZE      = 500 //  500 data bit per frame
	S_WINDOW_SIZE  = 2   // sliding window size: choose from 0-8
	MAX_RESEND     = 8   // Max resend times before link error
	RTT            = 200 // Round Trip Time in milisecond
	POWER_SIGNAL   = 0.1 // >powersignal means signal, it is examnied by matlab
)
