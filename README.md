# Rua!

一个简单的、高度可定制化的游戏服务器框架

![architecture](./img/architecture.png)

## 特性

- 帧同步(lockstep)服务器
  - 支持根据延迟动态调节步长
- FIFO事件驱动服务器
- 支持使用自定义网络通信协议
  - 目前已经提供了websocket、kcp两种协议的插件
- 生命周期钩子
- 级联架构
- 标记Peer
- 自定义日志系统
- Stdio交互
- 文件输出

## 安装

安装服务器本体

```bash
go get github.com/DiscreteTom/rua
```

安装插件（以websocket为例）

```bash
go get github.com/DiscreteTom/rua/plugin/network/websocket
```

## 入门

> 运行此示例代码需要安装websocket插件

```go
package main

import (
	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/websocket"
)

func main() {
	s := rua.NewEventDrivenServer().
		SetHandleKeyboardInterrupt(true).
		OnPeerMsg(func(peers map[int]rua.Peer, msg *rua.PeerMsg, _ *rua.EventDrivenServer) {
			// broadcast to everyone
			for _, p := range peers {
				go p.Write(msg.Data)
			}
		})

	// start websocket listener, which will generate peer to the game server
	go websocket.NewWebsocketListener(":8080", s).Start()

	// start event driven game server
	s.Start()
}
```

## 更多示例

- [Websocket](https://github.com/DiscreteTom/rua/tree/main/plugin/network/websocket/_example)
- [KCP](https://github.com/DiscreteTom/rua/tree/main/plugin/network/kcp/_example)

## 日志

- rua自身提供了默认的logger，可以使用`rua.GetDefaultLogger()`来获取它，或者使用`rua.SetDefaultLogger`来覆盖它
  - 比如使用`rua.SetDefaultLogger(rua.NewDefaultLogger().WithLevel(rua.DEBUG))`来设置日志等级
- 您也可以使用您自己的logger，比如使用logrus：`rua.SetDefaultLogger(logrus.New())`
- 您可以对服务器使用`.GetLogger()`获取服务器的logger，进行日志输出
- 您也可以使用`rua.NewBasicLogger`、`rua.NewBasicSimpleLogger`这两个helper函数，快速构建自定义的logger

## TODO

- [ ] 流数据输入/输出（用来进行异步外挂检测）
  - [ ] Kafka
- [ ] 持久化（用来实现大数据分析、战斗回放等）
  - [ ] 数据库
- [ ] WebSocket + TLS
- [ ] KCP + smux
- [ ] Docker

