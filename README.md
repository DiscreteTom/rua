# Rua!

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/DiscreteTom/rua?style=flat-square)
![GitHub](https://img.shields.io/github/license/DiscreteTom/rua?style=flat-square)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/DiscreteTom/rua?style=flat-square)

> [English README](#english)

事件驱动异步通信框架，基于[Go 语言](https://golang.org/)编写。

此项目还有一个[Rust 语言](https://www.rust-lang.org/)的版本：[Ruast](https://github.com/DiscreteTom/ruast)!

## 安装

安装服务器本体

```bash
go get github.com/DiscreteTom/rua
```

安装插件（以[websocket](https://github.com/DiscreteTom/rua/tree/main/plugin/network/websocket)为例）

```bash
go get github.com/DiscreteTom/rua/plugin/network/websocket
```

## 举个例子

下面的代码创建了一个 WebSocket 广播服务器：

```go
package main

import (
	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/websocket"
)

// Use `wscat -c ws://127.0.0.1:8080` to connect to the websocket server.
func main() {
	// create broadcaster
	bc := rua.NewBroadcaster()

	// websocket listener
	ws, _ := websocket.NewWsListener("127.0.0.1:8080").OnNewPeer(func(wn *websocket.WsNode) {
		// add new peer to the broadcaster
		bc.AddTarget(wn.OnMsg(func(b []byte) {
			// new message will be write to the broadcaster
			bc.Write(b)
		}).Go())
	}).Go() // start the listener

	// also broadcast to stdout
	bc.AddTarget(rua.DefaultStdioNode().Go())

	// wait for ctrl-c
	rua.NewCtrlc().OnSignal(func() {
		bc.StopAll()
		ws.Stop()
	}).Wait()
}
```

## 更多示例

- [基本示例](https://github.com/DiscreteTom/rua/tree/main/example)
- [Websocket](https://github.com/DiscreteTom/rua/tree/main/plugin/network/websocket/_example)
- [KCP](https://github.com/DiscreteTom/rua/tree/main/plugin/network/kcp/_example)

## [更新日志](https://github.com/DiscreteTom/rua/blob/main/CHANGELOG.md)

# English

Rua is an event-driven async messaging framework written with [golang](https://golang.org/).

This project also has a [Rust-lang](https://www.rust-lang.org/) implementation: [Ruast](https://github.com/DiscreteTom/ruast)!

## Installation

Install the rua server itself:

```bash
go get github.com/DiscreteTom/rua
```

Install plugins (e.g. [websocket](https://github.com/DiscreteTom/rua/tree/main/plugin/network/websocket)):

```bash
go get github.com/DiscreteTom/rua/plugin/network/websocket
```

## Getting Started

The following code shows how to create a websocket broadcast server.

```go
package main

import (
	"github.com/DiscreteTom/rua"
	"github.com/DiscreteTom/rua/plugin/network/websocket"
)

// Use `wscat -c ws://127.0.0.1:8080` to connect to the websocket server.
func main() {
	// create broadcaster
	bc := rua.NewBroadcaster()

	// websocket listener
	ws, _ := websocket.NewWsListener("127.0.0.1:8080").OnNewPeer(func(wn *websocket.WsNode) {
		// add new peer to the broadcaster
		bc.AddTarget(wn.OnMsg(func(b []byte) {
			// new message will be write to the broadcaster
			bc.Write(b)
		}).Go())
	}).Go() // start the listener

	// also broadcast to stdout
	bc.AddTarget(rua.DefaultStdioNode().Go())

	// wait for ctrl-c
	rua.NewCtrlc().OnSignal(func() {
		bc.StopAll()
		ws.Stop()
	}).Wait()
}
```

## More Examples

- [Basics](https://github.com/DiscreteTom/rua/tree/main/example)
- [Websocket](https://github.com/DiscreteTom/rua/tree/main/plugin/network/websocket/_example)
- [KCP](https://github.com/DiscreteTom/rua/tree/main/plugin/network/kcp/_example)

## [Change Log](https://github.com/DiscreteTom/rua/blob/main/CHANGELOG.md)
