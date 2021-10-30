# CHANGELOG

## v0.6.0

Inspired by the [ruast](https://github.com/DiscreteTom/ruast) project, refactor this project.

Nodes:

- StdioNode
- Ctrlc
- FileNode
- Lockstep

Model:

- StoppableHandle
- WritableStoppableHandle

## v0.5.0

- Add `BroadcastPeer`.
- Add `BufferPeer`.
- Change model.
- Add helper functions. See `utils.go`.

## v0.4.2

- Enhance peers' security.

## v0.4.1

- `EventDrivenServer.BeforeAddPeer` can not access the new peer's id.

## v0.4.0

- Optimize code.
- Add `SafePeer`.
- Almost all APIs are changed.

## v0.3.0

- Supported server types:
  - `LockstepServer`.
  - `EventDrivenServer`
- Supported peers:
  - `StdioPeer`
  - `FilePeer`
  - `NetPeer`
  - `KcpPeer`
  - `WebsocketPeer`
  - `KinesisPeer`
  - `BasicPeer`
