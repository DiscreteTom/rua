# Rua!

一个简单的游戏服务器框架。**仍在开发中，非常不稳定**

## 特性

- 帧同步(lockstep)服务器
  - 支持根据延迟动态调节步长
- FIFO事件驱动服务器
- 支持使用自定义网络通信协议
  - 目前已经提供了websocket、kcp两种协议的插件

## 安装

```bash
go get github.com/DiscreteTom/rua
```

## [示例](https://github.com/DiscreteTom/rua/tree/main/example)



