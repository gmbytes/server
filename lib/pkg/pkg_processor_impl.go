package pkg

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand/v2"
	"net"

	"github.com/gmbytes/snow/core/xnet"
)

const (
	netMagicNumber uint16 = 0xffff
)

var ServerPkgPreprocessor xnet.IPreprocessor = &serverPkgPreprocessor{}

type serverPkgPreprocessor struct {
}

func (ss *serverPkgPreprocessor) Process(conn net.Conn) error {
	bs := make([]byte, 4)
	if _, err := io.ReadFull(conn, bs); err != nil {
		return err
	}

	n := binary.LittleEndian.Uint16(bs)
	if n != netMagicNumber {
		return fmt.Errorf("magic number not match")
	}

	sum := md5.Sum(bs[:])
	bs[0] = sum[bs[2]>>4]
	bs[1] = sum[bs[2]&0xf]
	bs[2] = sum[bs[3]&0xf]
	bs[3] = sum[bs[3]>>4]

	if _, err := conn.Write(bs); err != nil {
		return err
	}

	return nil
}

var ClientPkgPreprocessor xnet.IPreprocessor = &clientPkgPreprocessor{}

type clientPkgPreprocessor struct {
}

func (ss *clientPkgPreprocessor) Process(conn net.Conn) error {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint16(bs, netMagicNumber)
	binary.LittleEndian.PutUint16(bs[2:], uint16(rand.Uint32()))

	if _, err := conn.Write(bs[:4]); err != nil {
		return err
	}

	sum := md5.Sum(bs[:4])

	bs[4] = sum[bs[2]>>4]
	bs[5] = sum[bs[2]&0xf]
	bs[6] = sum[bs[3]&0xf]
	bs[7] = sum[bs[3]>>4]

	if _, err := io.ReadFull(conn, bs[:4]); err != nil {
		return err
	}

	if !bytes.Equal(bs[:4], bs[4:]) {
		return fmt.Errorf("server response check failed")
	}

	return nil
}
