package game

import (
	"server/internal/pb"
)

// dispatchPkg 将应答包序列化并通过 Gate 发送给客户端
func (ss *Game) dispatchPkg(sess *actorSession, p *pb.Package) {
	ss.printPkg(sess, p)

	data, err := p.Bytes()
	if err != nil {
		ss.Errorf("marshal pkg failed: %v", err)
		return
	}
	ss.sendToClientByConnId(sess.connId, data)
}

func (ss *Game) printPkg(sess *actorSession, p *pb.Package) {
	key := p.Key()
	keyName := pb.EKey_T_name[int32(key)]
	ss.Debugf("Send(%v): %s (%+v)", sess.descriptor(), keyName, p)
}
