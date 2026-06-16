package common

import "time"

// RegistryHeartbeatInterval 心跳间隔
const RegistryHeartbeatInterval = 30

// Timer 定时器（单次触发）
type Timer struct {
	C <-chan time.Time
	t *time.Timer
}

// NewTimer 创建单次定时器
func NewTimer(d time.Duration) *Timer {
	t := time.NewTimer(d)
	return &Timer{
		C: t.C,
		t: t,
	}
}

// Stop 停止定时器
func (t *Timer) Stop() {
	if t.t != nil {
		t.t.Stop()
	}
}

// Reset 重置定时器（用于循环触发）
func (t *Timer) Reset(d time.Duration) {
	if t.t != nil {
		t.t.Reset(d)
	}
}

// Ticker 周期性定时器
type Ticker struct {
	C <-chan time.Time
	t *time.Ticker
}

// NewTicker 创建周期性定时器
func NewTicker(d time.Duration) *Ticker {
	t := time.NewTicker(d)
	return &Ticker{
		C: t.C,
		t: t,
	}
}

// Stop 停止定时器
func (t *Ticker) Stop() {
	if t.t != nil {
		t.t.Stop()
	}
}
