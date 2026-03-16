package game

import (
	"fmt"
	"server/pkg/utils"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gmbytes/snow/core/task"
)

type metric struct {
	maxTime    int64
	totalTime  int64
	totalCount int64
}

var metricMap = make(map[string]*metric)
var metricMapLock sync.Mutex

func addMetric(name string, timeMs int64) {
	metricMapLock.Lock()
	defer metricMapLock.Unlock()

	m := metricMap[name]
	if m == nil {
		m = &metric{}
		metricMap[name] = m
	}

	m.totalCount++
	m.totalTime += timeMs

	if m.maxTime < timeMs {
		m.maxTime = timeMs
	}
}

func startMetric(interval int) {
	task.Execute(func() {
		for {
			select {
			case <-time.Tick(time.Duration(interval) * time.Second):
				printMetric()
			}
		}
	})
}

func printMetric() {
	metricMapLock.Lock()

	keys := utils.Keys(metricMap)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("\t%-30v %12v %12v %12v", "Name", "Count", "Avg Time", "Max Time"))
	for _, name := range keys {
		m := metricMap[name]
		avgTime := int64(0)
		if m.totalCount > 0 {
			avgTime = m.totalTime / m.totalCount
		}
		sb.WriteString(fmt.Sprintf("\n\t%-30v %12v %12v %12v", name, m.totalCount, avgTime, m.maxTime))
	}

	metricMapLock.Unlock()

	println(sb.String())
}

func (ss *Game) startMetric() {
	if ss.opt.MetricInterval > 0 {
		startMetric(ss.opt.MetricInterval)
	}
}
