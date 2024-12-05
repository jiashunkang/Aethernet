package shared

// 常量定义
const (
	MaxBufferSize  = 1024
	ChirpStartFreq = 2000
	ChirpEndFreq   = 10000
	FS             = 48000 // Sample Frequency
	PreambleLength = 48    // Length of preamble signal
	FC             = 10000 // Carrier frequency
	SYNC_PARA      = 0.2
	DATA_SIZE      = 400 + 2                             //  data bit per mac frame, extra 2 bit to distinguish different IP pack(see readme proejct3 for detail ugly implementation ToT)
	S_WINDOW_SIZE  = 4                                   // sliding window size: choose from 0-8
	MAX_RESEND     = 16                                  // Max resend times before link error
	RTT            = 200                                 // Round Trip Time in milisecond
	POWER_SIGNAL   = 0.04                                // >powersignal means signal, it is examnied by matlab
	MTU            = (DATA_SIZE - 2) * S_WINDOW_SIZE / 8 // IP MTU
)
