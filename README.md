# UDP Ping

UDP Ping 是一个基于 UDP 协议的网络探测工具，用于测试 UDP 连接的可达性和丢包率。

## 项目结构

```
udpping/
├── cmd/
│   ├── udpping-client/     # UDP Ping 客户端
│   └── udpping-server/     # UDP Ping 服务端
├── pkg/
│   ├── protocol/           # 共享协议定义
│   └── stats/              # 统计模块
├── docs/                   # 文档目录
└── README.md
```

## 功能特性

### UDP Ping Server
- 监听 UDP 端口接收探测报文
- 统计接收到的报文数量
- 计算丢包率
- 定期向客户端发送反馈报文
- 实时显示统计信息

### UDP Ping Client
- 按照指定频率向服务端发送探测报文
- 接收服务端反馈报文
- 统计发送和接收情况
- 显示实时连接状态

## 构建

### 构建服务端
```bash
go build -o udpping-server ./cmd/udpping-server
```

### 构建客户端
```bash
go build -o udpping-client ./cmd/udpping-client
```

### 构建所有程序
```bash
go build ./cmd/udpping-server
go build ./cmd/udpping-client
```

## 使用方法

### 启动服务端

```bash
./udpping-server -port 9999 -report 10s
```

参数说明：
- `-port`: UDP 监听端口（默认：9999）
- `-report`: 统计信息报告间隔（默认：10s）

### 启动客户端

```bash
./udpping-client -server localhost:9999 -rate 1s
```

参数说明：
- `-server`: 服务端地址，格式为 host:port（默认：localhost:9999）
- `-rate`: 发送探测报文的频率（默认：1s）
- `-timeout`: 接收反馈报文的超时时间（默认：5s）

## 使用示例

### 本地测试

1. 启动服务端：
```bash
./udpping-server -port 9999
```

2. 在另一个终端启动客户端：
```bash
./udpping-client -server localhost:9999 -rate 1s
```

### 远程测试

1. 在服务器上启动服务端：
```bash
./udpping-server -port 9999
```

2. 在本地启动客户端：
```bash
./udpping-client -server <服务器IP>:9999 -rate 1s
```

## 协议格式

UDP Ping 使用自定义的二进制协议，每个数据包包含：

- Type (1 byte): 包类型（1=探测包，2=反馈包）
- Sequence (8 bytes): 序列号
- Timestamp (8 bytes): 时间戳（纳秒）

总计 17 字节的包头。

## 统计信息

程序会显示以下统计信息：
- Duration: 运行时长
- Sent: 已发送的报文数量
- Received: 已接收的报文数量
- Lost: 丢失的报文数量
- Loss Rate: 丢包率（百分比）

## 模块说明

### pkg/protocol
定义了 UDP Ping 协议的数据包格式和序列化/反序列化方法。

### pkg/stats
提供统计功能，包括发送/接收计数、丢包率计算等。

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
