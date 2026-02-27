package msg_hub

import "github.com/gmbytes/snow/routines/node"

func init() {
	node.Register[MsgHub, *MsgHub]("MsgHub")
}

type MsgHub struct {
	node.Service
}

func (ss *MsgHub) Init(args any) {

}
