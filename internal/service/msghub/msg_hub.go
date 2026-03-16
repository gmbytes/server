package msg_hub

import "github.com/gmbytes/snow/routines/node"

//nolint:gochecknoinits // service registration pattern
func init() {
	node.Register[MsgHub, *MsgHub]("MsgHub")
}

type MsgHub struct {
	node.Service
}

func (*MsgHub) Init(_ any) {

}
