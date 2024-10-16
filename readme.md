# Project0
### How to run
The code for task 1 is in `CODE_FOR_TASK_1` file, copy it and paste in `main.go`, Then in command line, type
```
go run main.go
```
Speak to microphone for 10 seconds, and then your device will replay the recorded sound.

The code for task 2 is in `CODE_FOR_TASK_2` file, copy it and paste in `main.go`, Then in command line, type
```
go run main.go
```
Your device shall play `Sample.wav` for 10 seconds, and then your device will replay the recorded sound.

### Reference
`https://github.com/wenxuanjun/jack-starter`

### Tips
Remember to disable "enhancement" for microphone (记得关声音增强功能，否则麦克风自动过滤扬声器的声音)

用AMD Audio device而不是realtek立体声混音

# Project 1
To run the project, run the following command in the `computernetworks-shanghaitech` directory
```
go build
```
```
.\acoustic_link.exe
```
## 代码结构
### main.go
在main.go里的main函数注册了和jack audio的输入输出port，以及开了outputchannel用来从transmitter里传数据给output port

然后在初始化完成后 用```go transmitter.Start()``` 开一个线程运行transmitter的Start函数
### transmitter.go
- 在transmitter里首先NewTransmitter函数创建了一个struct用来存需要发送的data，outputchannel等一些信息，初始化一个480sample长的chirp从2000hz扫到10000hz再扫回2000hz，提前计算好存好。

- Start函数----最主要的一个函数

    1. 把10000个存在data里的bit数据切成100个bit为单位的frame
    2. 利用`util.g0` 里的CRC8算法算出一个8位bit但是以int数组方式存，append在原来的data后面，组成108个bit
    3. 用modulate进行调制，用PSK，如果是0就反转相位，并且每个bit重复48遍
    4. 进行播放（先播放ramdom个0，再播放preamble，再播放数据，再播放ramdom个0）
    5. 异常情况：发现传输过程中每个frame前8个bit一直会出错，于是给每个frame开头又加了8个trivial bit对冲错误，暂时没找到原因
    
    - frame structure：
    random 0 --- preamble(2khz-10khz-2khz)(480) --- modulate(trivial_bit(8bit)+frame(100bit))*48(repeat times) --- random 0

### receiver.go
- 创建 receiver struct，存储 chirp preamble， inputchannel，decode_data（解码后的01bit用来输入到txt最后比对）， carrier（一个10khz正弦波数组，用来和信号相乘解调）
- Start函数，用gpt把matlab代码直接翻译过来，感觉go没有很方便的矩阵计算的包，纯手搓点乘和滑动平均但目前性能足够。
    1. state 0： 还在通过preamble和信号相乘找同步
    2. state 1：用载波carrier和信号相乘解码 
### utils.go 
1. 处理int数组存的{0，1，0，1}，转换成byte形式的0101，方便调用外部的crc算法。

### Go语言
大写字母的函数变量是可以被其他文件引用的，小写字母的函数和变量只能在同一个文件里被引用





