# CHANGELOG

## v0.4.3

- Add `BroadcastPeer`.
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
