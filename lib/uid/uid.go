package uid

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

type Uid int64

var Zero Uid = 0

var VAR_SCENE_EntityId = Uid(1)

func (x Uid) ToString() string {
	return strconv.FormatInt(int64(x), 10)
}

func (x Uid) ToInt64() int64 {
	return int64(x)
}

func (x Uid) IsValid() bool {
	return x != 0
}

func FromString(str string) Uid {
	if len(str) == 0 {
		return 0
	}

	id, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		log.Fatalf("string(%s) to uid error: %v", str, err)
	}

	return Uid(id)
}

const (
	startTime = 1710138991 // 2024-03-11 14:36:31
	maxIndex  = 1<<20 - 1
	// node id ：自动生成范围[1,6000),正式部署范围[1,10000),跨服范围[10000,15000), 其它用途[15000,16000)
	maxNodeID = 1<<14 - 1
)

var (
	lastTime  = time.Now().Unix() - startTime
	lastIndex int64
	nodeID    int64
	lock      int32
)

// nid ：自动生成范围[1,6000),正式部署范围[1,10000),跨服范围[10000,15000), 其它用途[15000,16000)
func Init(nid int64) {
	if nid < 0 || nid > maxNodeID {
		panic(fmt.Sprintf("node id overflow  ,node id:%d", nid))
	}
	nodeID = nid
}

func Gen() Uid {
	for !atomic.CompareAndSwapInt32(&lock, 0, 1) {
		runtime.Gosched()
	}

	for {
		t := time.Now().Unix() - startTime
		if t == lastTime {
			if lastIndex >= maxIndex {
				continue
			}

			lastIndex++
		} else {
			lastIndex = 0
			lastTime = t
		}
		id := (lastTime<<34 | (nodeID << 20) | lastIndex) & (1<<63 - 1)
		atomic.StoreInt32(&lock, 0)
		return Uid(id)
	}
}
