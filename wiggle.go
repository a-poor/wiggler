package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go.uber.org/atomic"
)

const (
	WindowWidth       = 600 // Fixed window width
	WindowSmallHeight = 300 // Height of the window when details closed
	WindowLargeHeight = 600 // Height of the window when details open
)

type WiggleEvent int

const (
	WiggleEventStop WiggleEvent = iota
	WiggleEventStart
)

type WiggleConfig struct {
	WiggleSpeed int `json:"wiggleSpeed"`
	WaitTime    int `json:"waitTime"`
}

func NewDefaultWiggleConfig() WiggleConfig {
	return WiggleConfig{
		WiggleSpeed: 1,
		WaitTime:    5,
	}
}

type Wiggler struct {
	ctx    context.Context // The wails context
	config struct {
		lock      sync.RWMutex // Locks the wiggler's config data
		moveSpeed int          // How fast should the wiggler move? (in seconds)
		waitTime  int          // How long should the wiggler wait before moving again? (in seconds)
	}

	isReady    *atomic.Bool // Is the wiggler ready?
	isWiggling *atomic.Bool // Is the wiggler currently wiggling?

	cancel  context.CancelFunc // Cancel function for the wiggler
	doneWG  sync.WaitGroup     // Wait group for the wiggler. Done signals that it's fully complete.
	wevents chan WiggleEvent   // Channel for wiggle events
}

func NewWiggler(cancel context.CancelFunc, wevents chan WiggleEvent, cfg WiggleConfig) *Wiggler {
	return &Wiggler{
		config: struct {
			lock      sync.RWMutex
			moveSpeed int
			waitTime  int
		}{
			moveSpeed: cfg.WiggleSpeed,
			waitTime:  cfg.WaitTime,
		},
		isReady:    atomic.NewBool(false),
		isWiggling: atomic.NewBool(false),
		cancel:     cancel,
		wevents:    wevents,
	}
}

func (w *Wiggler) SetWindowSmall() {
	runtime.WindowSetSize(w.ctx, WindowWidth, WindowSmallHeight)
}

func (w *Wiggler) SetWindowLarge() {
	runtime.WindowSetSize(w.ctx, WindowWidth, WindowLargeHeight)
}

func (w *Wiggler) IsReady() bool {
	return w.isReady.Load()
}

func (w *Wiggler) IsWiggling() bool {
	return w.isWiggling.Load()
}

func (w *Wiggler) GetMoveSpeed() int {
	w.config.lock.RLock()
	defer w.config.lock.RUnlock()
	return w.config.moveSpeed
}

func (w *Wiggler) SetMoveSpeed(s int) {
	w.config.lock.Lock()
	defer w.config.lock.Unlock()
	w.config.moveSpeed = s
}

func (w *Wiggler) GetWaitTime() int {
	w.config.lock.RLock()
	defer w.config.lock.RUnlock()
	return w.config.waitTime
}

func (w *Wiggler) SetWaitTime(t int) {
	w.config.lock.Lock()
	defer w.config.lock.Unlock()
	w.config.waitTime = t
}

func (w *Wiggler) GetConfig() WiggleConfig {
	w.config.lock.RLock()
	defer w.config.lock.RUnlock()
	return WiggleConfig{
		WiggleSpeed: w.config.moveSpeed,
		WaitTime:    w.config.waitTime,
	}
}

func (w *Wiggler) SetConfig(speed, wait int) {
	runtime.LogInfo(w.ctx, fmt.Sprintf("Setting wiggler config to speed: %d, wait: %d", speed, wait))
	w.config.lock.Lock()
	defer w.config.lock.Unlock()
	w.config.moveSpeed = speed
	w.config.waitTime = wait
}

func (w *Wiggler) OnStartup(ctx context.Context) {
	// Save runtime
	w.ctx = ctx
	runtime.LogInfo(ctx, "Wiggler initialised")

	// Set ready-ness
	w.isReady.Store(true)
}

func (w *Wiggler) OnDomReady(ctx context.Context) {
	// Emit a startup event
	runtime.EventsEmit(w.ctx, "ready")
}

func (w *Wiggler) OnShutdown(ctx context.Context) {
	w.CancelWiggler()
}

func (w *Wiggler) StartWiggle() {
	runtime.LogInfo(w.ctx, "Starting wiggle")
	w.wevents <- WiggleEventStart
}

func (w *Wiggler) StopWiggle() {
	runtime.LogInfo(w.ctx, "Starting wiggle")
	w.wevents <- WiggleEventStop
}

// CancelWiggler cancels the whole wiggler loop
func (w *Wiggler) CancelWiggler() {
	runtime.LogInfo(w.ctx, "Cancelling wiggler")

	// Cancel the wiggler, if it's running
	if w.cancel != nil {
		w.cancel()
	}

	// Wait for cleanup to complete
	w.doneWG.Wait()
}
