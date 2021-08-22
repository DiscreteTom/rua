# Rua!

一个简单的、高度可定制化的游戏服务器框架

## 特性

- 帧同步(lockstep)服务器
  - 支持根据延迟动态调节步长
- FIFO事件驱动服务器
- 支持使用自定义网络通信协议
  - 目前已经提供了websocket、kcp两种协议的插件
- 生命周期钩子

## 安装

```bash
go get github.com/DiscreteTom/rua
```

## 入门

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

## [更多示例](https://github.com/DiscreteTom/rua/tree/main/example)



