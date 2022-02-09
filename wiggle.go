package main

import (
	"context"
	"sync"

	"github.com/wailsapp/wails"
	"go.uber.org/atomic"
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
	runtime *wails.Runtime      // Pointer to Wails runtime
	logger  *wails.CustomLogger // Pointer to Wails logger

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
	w.logger.Debugf("Setting wiggler config to speed: %d, wait: %d", speed, wait)
	w.config.lock.Lock()
	defer w.config.lock.Unlock()
	w.config.moveSpeed = speed
	w.config.waitTime = wait
}

func (w *Wiggler) WailsInit(runtime *wails.Runtime) error {
	// Save runtime
	w.runtime = runtime
	w.logger = runtime.Log.New("Wiggler")
	w.logger.Info("Wiggler initialised")

	// Emit a startup event
	w.runtime.Events.Emit("ready")

	// Set ready-ness
	w.isReady.Store(true)

	return nil
}

func (w *Wiggler) WailsShutdown() {
	w.CancelWiggler()

	// Emit a shutdown event
	// w.runtime.Events.Emit("stopped")
}

func (w *Wiggler) StartWiggle() {
	w.logger.Debug("Starting wiggle")
	w.wevents <- WiggleEventStart
	// w.runtime.Events.Emit("wiggle-started")
}

func (w *Wiggler) StopWiggle() {
	w.logger.Debug("Starting wiggle")
	w.wevents <- WiggleEventStop
	// w.runtime.Events.Emit("wiggle-stopped")
}

// CancelWiggler cancels the whole wiggler loop
func (w *Wiggler) CancelWiggler() {
	w.logger.Info("Cancelling wiggler")

	// Cancel the wiggler, if it's running
	if w.cancel != nil {
		w.cancel()
	}

	// Wait for cleanup to complete
	w.doneWG.Wait()
}
