package account

import (
	"sync"

	"github.com/gmbytes/snow/routines/node"
)

type Account struct {
	node.Service

	sDb node.IProxy
}

func (ss *Account) Start(args any) {
	
}

func (ss *Account) Stop(wg *sync.WaitGroup) {

}
