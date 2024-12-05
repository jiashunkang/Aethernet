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
- 大写字母的函数变量是可以被其他文件引用的，小写字母的函数和变量只能在同一个文件里被引用
- 修好了每个frame前8个bit必错的问题：原因是frame是从data里切出来的一个slice，对于frameCRC来说append(frame,CRC(frame))似乎并不能达到预期的效果，需要提前给frameCRC分配好内存make108大小的数组
- 修好了最后一个frame错误率很高的问题，大概是因为outputbuffer延迟问题吧。。

# 多线程问题：Channel大小 和 Jack buffer大小 对正确率影响
1. Transmitter outputchannel和outbuffer问题
   1. transimitter需要做的是把数字的0、1bit调制成一串float32数组，并把这串数组传给outputchannel，然后通过go的channel将float32数值传给jack的callback函数process中的outbuffer[]。transimitter处理数组是很快的，播放声音又被限制在48000hz比较慢，所以每次处理到outputchannel size这么多的值，channel就会阻塞，transmitter线程就会卡住。至于卡住会导致什么后果我还没仔细研究，但是如果我把outputchannel调的非常大，也就是杜绝这种卡顿现象出现，正确率由每次错30bit左右上升到每次错1-2bit甚至全对。
2. Receiver inputChannel和inputBuffer
   1. inputChannel相对来说不会像outputChannel那样容易因为值传递的太快导致被装满而阻塞。inputChannel和Receiver面临的问题是，接收声音的频率比较慢，所以Receiver可能因为没有等到值而卡顿，为了避免卡顿，可以在jack里把buffer的大小调小一点比如32和64，减少Receiver卡顿的情况。但这个问题似乎对于正确率结果影响不大。
   2. 为什么会注意到这个问题？因为我写了一个receiver_test.go,我把麦克风接收到的输入存了一个备份（input_track.csv）用来调试，结果发现居然实时测量会比之后调试多错几个bit？？于是怀疑是线程和同步导致的错误，但在解决了transmitter的问题后，我就再也没有复现出这个问题了。

# Project 2

## frame结构
### MAC层
- 滑动窗口实现
  - transmitter.send函数在写入outputChannel前加一个mutex lock，防止几个线程同时写入outputchannel引起collision
  - mac 创建存储data frame的一个list，命名为window，窗口大小为8
  - mac 维护 timeout channel list 和 freetimeout channel list，大小为16（因为对应frame的id，需要窗口大小的两倍，防止混淆）
- Data
   - 读入数据,把byte切成bit
   - 封装mac frame **Dest(1bit) Src(1bit) Type(1bit) ID(4bit - 16需要时窗口大小的两倍) Data(500bit)** 总计504bit
     - id bit区分前后发送frame的不同
   - 将frame传递给物理层transmitter，封装**preamble（44bit） + mac length in bits(9bit 即512 留出冗余) + macframe + crc**
- ACK
    - 封装mac frame **Dest(1bit) Src(1bit) Type(1bit) ID(4bit - 16需要时窗口大小的两倍)** 总计7bit
    - 将frame传递给物理层transmitter，**封装preamble（44bit） + mac length in bits(9bit 即512 留出冗余) + macframe + crc**
- 解调步骤
  - 照搬receiver的sync
  - 如果接收到sync信号，切换为读取header模式
  - 解码mac length + macframe的前3个bit（src，dest，type）
    - 先判断src dest是不是自己发自己，是的话就切回寻找sync的状态
    - Data type == 0
      - 照旧处理，收集一个macframe，返回给MAC层，等mac去发ack
      - 如果crc error就返回error
    - ACK type == 1
      - 如果满足 CRC 正确，那么就返回MAC层成功

### Timeout机制


### 线程安排
- MAC线程，运行状态机
- Receiver线程 
- transmitter每次传输时由MAC创建go transimtter.send
  - 发送完后维护timeout channel，当计时(time.sleep)达到timeout时发起重传  

### 传输媒介物理性质问题
发现如果用连续的1和-1代表1和0在单线连自己的时候没有问题，但是当使用mixer连接第二台电脑时会出现很多识别错误，这个问题在把调制方式改为{-1,-1,1,1}代表1，{1,1,-1,-1}代表0后被修正，~~但还不知道具体原理~~ 因为两台电脑的基准电压不一样，在第一台电脑的0可能会被第二台当作-0.5，这样原来<0.5的值被默认<0，然后原来的1会被解码成0.

### CSMA机制
- 什么时候发起sense&backoff
  - 每次发送数据包之前需要做这样的操作
  - 发送ACK不需要sense和backoff，直接发送
  - 如果MAC线程正处于backoff阶段就不再重复backoff
- Sense
  - 由Receiver线程计算power，并持续发送通过powerChan发送给MAC线程
  - mac每次调用senseSignal函数时，会把powerChan中数据全读完，只取最后一个作为最近的sense结果，如果大于POWER_SIGNAL,就说明当前有数据在传输，需要bckoff
- Backoff
  - 每次sense到有信号，就backoff一段时间

# Project 3
### IP层和MAC层数据交流
IP层将IP的内容包括包头转为byte[], 交给IO helper
  - MTU = DATA_SIZE * S_WINDOW_SIZE / 8
  - IP层自己决定IP packet大小是否超过MTU，超过就要自己分好再交给MAC层
  - 暂定 400 bit----50 byte 一个mac packet
  - 暂定 400 bit*4 ---- 50*4=200 byte 一个MTU
  - mac层新增两个bit
    - 00 此IP包第一个mac frame
    - 11 IP包最后一个mac frame（即可以返回上层了）
    - 01 表示这个IP包就这一个mac frame （即可以返回上层了）
    - 10 IP包中间的frame

**io helper作为ip和mac中间层的工作**
  - 维护一个data buffer列表，链表一个slot代表一个IP包
  - 可以供mac层读入固定大小bit数组
  - 可以返回IP层完整的IP byte数组

### IP层构建ICMP包
- 第三方库 gopacket，输入各字段的内容，然后调用serialize转化为byte数组，交给mac层
```bash
go get github.com/google/gopacket
```
```bash
go get github.com/google/gopacket/layers
```

### 记得装依赖和把防火墙ICMPv4入站出站打开

ping 172.182.3.111 -n 3

### JACK和ASIO4ALL
JACK如果使用默认的driver会有很高的延迟，实测如果设置128的buffer大小，大概50ms才会执行一次process callback函数。但是在qjackctl的setup里面可以改，选interface里的“ASIO::ASIO4ALL v2”选项。

- 奇怪的bug
  - 在台式机上选择ASIO后运行，apply之后，点启动jack没有问题，但是笔记本上就会报错，也找不到具体原因，重装了jack和asio之后依旧报错。
    - 解决方案，JACK2路径下（有jackd.exe）在这个目录下打开终端，输入
      ```
      ./jackd.exe -S -X winmme -dportaudio -d"ASIO::ASIO4ALL v2" -r48000 -p128
      ```
      输入以上命令手动打开貌似就可以运行了，不知道为什么会有这么奇怪的错。
  - ASIO有时不能识别我们想用的USB声卡而是会默认选择电脑自带的声卡
    - 解决方案：在打开qjackctl之前，先把电脑音频选择为电脑自带的声卡，然后打开jack，启动，这样可能asio识别到电脑声卡被占用就会选择USB声卡了。
  
