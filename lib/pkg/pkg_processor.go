package pkg

import (
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

	return nil
}
