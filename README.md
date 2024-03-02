# snapcast-go

This is a toy project with aims to reimplement client capabilities of [snapcast client](https://github.com/badaix/snapcast).

## Current state

Implemented so far is:

- Codecs:

  - flac
  - oggvorbis

- Messages:
  - [Hello](https://github.com/badaix/snapcast/blob/develop/doc/binary_protocol.md#hello)
  - [Server settings](https://github.com/badaix/snapcast/blob/develop/doc/binary_protocol.md#server-settings)
  - [Codec header](https://github.com/badaix/snapcast/blob/develop/doc/binary_protocol.md#codec-header)
  - [Wire chunk](https://github.com/badaix/snapcast/blob/develop/doc/binary_protocol.md#wire-chunk)

The other - missing - messages can be seen in the [binary protocol documentation](https://github.com/badaix/snapcast/blob/develop/doc/binary_protocol.md).

## Using

For now, this repository must be checked out, all go dependencies must be installed and then it can be compiled and run:

```
$ git clone <repo url>
$ go get
$ go run .
```
