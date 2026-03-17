package pkg

import (
	"encoding/binary"

	"github.com/gmbytes/snow/pkg/xnet"

	"server/internal/pb"
	"server/pkg/net_pkg"
)

const (
	GS2GatePkgDispatch uint16 = 0xFF00 + iota
	GS2GatePkgSingleMessage
	GS2GatePkgBroadcastMessage
	GS2GatePkgCloseSessionMessage
)

const (
	Gate2GamePkgForward uint16 = 0xFE00 + iota
	Gate2GamePkgNewClient
	Gate2GamePkgClientDisconnect
	Gate2GamePkgClientReconnect
)

type IPkgBytes interface {
	Bytes() ([]byte, error)
}

var ServerPkgPreprocessor xnet.IPreprocessor = net_pkg.ServerPkgPreprocessor

type GS2GatePkgSingle struct {
	ConnId uint64
	Pkg    *pb.Package
}

func (p *GS2GatePkgSingle) Bytes() ([]byte, error) {
	if p == nil || p.Pkg == nil {
		return nil, nil
	}
	bs, err := p.Pkg.Marshal()
	if err != nil {
		return nil, err
	}
	out := make([]byte, 14+len(bs))
	binary.LittleEndian.PutUint16(out[0:2], GS2GatePkgSingleMessage)
	binary.LittleEndian.PutUint32(out[2:6], uint32(8+len(bs)))
	binary.LittleEndian.PutUint64(out[6:14], p.ConnId)
	copy(out[14:], bs)
	return out, nil
}

type GS2GatePkgCloseSession struct {
	ConnId       uint64
	DelaySeconds uint32
}

func (p *GS2GatePkgCloseSession) Bytes() ([]byte, error) {
	if p == nil {
		return nil, nil
	}
	out := make([]byte, 18)
	binary.LittleEndian.PutUint16(out[0:2], GS2GatePkgCloseSessionMessage)
	binary.LittleEndian.PutUint32(out[2:6], 12)
	binary.LittleEndian.PutUint64(out[6:14], p.ConnId)
	binary.LittleEndian.PutUint32(out[14:18], p.DelaySeconds)
	return out, nil
}
