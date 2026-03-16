package net_pkg

import (
	"bytes"
	"crypto/md5" //nolint:gosec // md5 used for non-security handshake protocol
	"encoding/binary"
	"fmt"
	"io"
	"math/rand/v2"
	"net"

	"github.com/gmbytes/snow/pkg/xnet"
)

const (
	netMagicNumber     uint16 = 0xffff
	clientHandshakeLen uint16 = 8
)

var ServerPkgPreprocessor xnet.IPreprocessor = &serverPkgPreprocessor{}

type serverPkgPreprocessor struct {
}

func (*serverPkgPreprocessor) Process(conn net.Conn) error {
	bs := make([]byte, 4)
	if _, err := io.ReadFull(conn, bs); err != nil {
		return err
	}

	n := binary.LittleEndian.Uint16(bs)
	if n != netMagicNumber {
		return fmt.Errorf("magic number not match")
	}

	sum := md5.Sum(bs)     //nolint:gosec // md5 used for non-security handshake protocol
	bs[0] = sum[bs[2]>>4]  //nolint:gosec // G602: index bounded by nibble (0-15), sum is [16]byte
	bs[1] = sum[bs[2]&0xf] //nolint:gosec // G602: index bounded by nibble (0-15), sum is [16]byte
	bs[2] = sum[bs[3]&0xf] //nolint:gosec // G602: index bounded by nibble (0-15), sum is [16]byte
	bs[3] = sum[bs[3]>>4]  //nolint:gosec // G602: index bounded by nibble (0-15), sum is [16]byte

	if _, err := conn.Write(bs); err != nil {
		return err
	}

	return nil
}

var ClientPkgPreprocessor xnet.IPreprocessor = &clientPkgPreprocessor{}

type clientPkgPreprocessor struct {
}

func (*clientPkgPreprocessor) Process(conn net.Conn) error {
	bs := make([]byte, clientHandshakeLen)
	binary.LittleEndian.PutUint16(bs, netMagicNumber)
	binary.LittleEndian.PutUint16(bs[2:], uint16(rand.Uint32()))

	if _, err := conn.Write(bs[:4]); err != nil {
		return err
	}

	sum := md5.Sum(bs[:4]) //nolint:gosec // md5 used for non-security handshake protocol
	bs[4] = sum[bs[2]>>4]
	bs[5] = sum[bs[2]&0xf]
	bs[6] = sum[bs[3]&0xf]
	bs[7] = sum[bs[3]>>4]

	if _, err := io.ReadFull(conn, bs[:4]); err != nil {
		return err
	}

	if !bytes.Equal(bs[:4], bs[4:]) {
		return fmt.Errorf("gs response check failed")
	}

	return nil
}
