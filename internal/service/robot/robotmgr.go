package robot

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type RobotMgr struct {
	wsURL   string
	verbose bool

	mu      sync.Mutex
	robots  []*Robot
	running atomic.Bool

	totalCreated atomic.Int64
	totalStopped atomic.Int64
}

func NewRobotMgr(wsURL string, verbose bool) *RobotMgr {
	return &RobotMgr{
		wsURL:   wsURL,
		verbose: verbose,
	}
}

func (m *RobotMgr) Start(count int, interval time.Duration) {
	if m.running.Swap(true) {
		fmt.Println("[RobotMgr] already running")
		return
	}

	fmt.Printf("[RobotMgr] starting %d robots, interval=%v, target=%s\n", count, interval, m.wsURL)

	go func() {
		for i := 0; i < count; i++ {
			if !m.running.Load() {
				break
			}
			idx := int(m.totalCreated.Add(1))
			bot := NewRobot(idx, m.wsURL, m)
			m.mu.Lock()
			m.robots = append(m.robots, bot)
			m.mu.Unlock()

			go func(b *Robot) {
				b.Run()
				m.totalStopped.Add(1)
			}(bot)

			if interval > 0 && i < count-1 {
				time.Sleep(interval)
			}
		}
		fmt.Printf("[RobotMgr] all %d robots launched\n", count)
	}()
}

func (m *RobotMgr) StopAll() {
	m.running.Store(false)
	m.mu.Lock()
	bots := make([]*Robot, len(m.robots))
	copy(bots, m.robots)
	m.mu.Unlock()

	fmt.Printf("[RobotMgr] stopping %d robots...\n", len(bots))
	for _, b := range bots {
		b.Stop()
	}
}

func (m *RobotMgr) StopN(n int) {
	m.mu.Lock()
	alive := m.getAlive()
	if n > len(alive) {
		n = len(alive)
	}
	toStop := alive[:n]
	m.mu.Unlock()

	fmt.Printf("[RobotMgr] stopping %d robots\n", len(toStop))
	for _, b := range toStop {
		b.Stop()
	}
}

func (m *RobotMgr) PrintStats() {
	m.mu.Lock()
	alive := m.getAlive()
	m.mu.Unlock()

	var (
		totalSend  int64
		totalRecv  int64
		totalErr   int64
		totalRTT   int64
		rttCount   int64
		stateCount = make(map[State]int)
	)

	for _, b := range alive {
		totalSend += b.stats.SendCount.Load()
		totalRecv += b.stats.RecvCount.Load()
		totalErr += b.stats.ErrorCount.Load()
		totalRTT += b.stats.RTTSum.Load()
		rttCount += b.stats.RTTCount.Load()
		stateCount[b.state]++
	}

	avgRTT := int64(0)
	if rttCount > 0 {
		avgRTT = totalRTT / rttCount
	}

	fmt.Println("═══════════════════════════════════════════════")
	fmt.Printf("  Robots alive:    %d / %d created / %d stopped\n",
		len(alive), m.totalCreated.Load(), m.totalStopped.Load())
	fmt.Printf("  States:          connected=%d loggedIn=%d roleReady=%d inScene=%d\n",
		stateCount[StateConnected], stateCount[StateLoggedIn],
		stateCount[StateRoleReady], stateCount[StateInScene])
	fmt.Printf("  Packets:         sent=%d  recv=%d  errors=%d\n",
		totalSend, totalRecv, totalErr)
	fmt.Printf("  Ping RTT avg:    %dms  (samples=%d)\n", avgRTT, rttCount)
	fmt.Println("═══════════════════════════════════════════════")
}

func (m *RobotMgr) AliveCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.getAlive())
}

func (m *RobotMgr) getAlive() []*Robot {
	alive := make([]*Robot, 0, len(m.robots))
	for _, b := range m.robots {
		if !b.stopped.Load() {
			alive = append(alive, b)
		}
	}
	return alive
}
