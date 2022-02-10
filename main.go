package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

const (
	version = "0.1.0"
	github  = "https://github.com/a-poor/wiggler"
)

// noOpCancel is a no-op function for that satisfies the
// context.CancelFunc type.
func noOpCancel() {}

//go:embed frontend/build
var assets embed.FS

//go:embed apple-icon.png
var macIcon []byte

func main() {
	// Create a context to cancel the wiggler
	ctx, cancel := context.WithCancel(context.Background())

	// Create a channel to receive wiggle events
	wevents := make(chan WiggleEvent)

	// Create the Wiggler
	wiggler := NewWiggler(cancel, wevents, NewDefaultWiggleConfig())

	// Increment the top-level wait group
	wiggler.doneWG.Add(1)

	// Run the wiggle-watcher
	go func() {
		// Signal that the top-level wiggler process is done running...
		defer wiggler.doneWG.Done()

		// Sub-context & corresponding cancel function
		// for a running wiggler
		var wctx context.Context
		wcancel := noOpCancel
		defer wcancel()

		// Wait group to track the wiggler status
		var wg sync.WaitGroup

		// This outer loop will watch for wiggle events or app cancelation
	outerLoop:
		for {
			select {
			case <-ctx.Done():
				break outerLoop

			case e := <-wevents:
				// A wiggle event has been received...

				// ...what type is it?
				switch e {
				case WiggleEventStop: // Is it a STOP event?
					// If the wiggle is already stopped, ignore this event
					if !wiggler.isWiggling.Load() {
						continue outerLoop
					}

					// ...otherwise, there is a running wiggle event

					// Yes, cancel the wiggler
					wcancel()

					// Reset the cancel function to a no-op
					wcancel = noOpCancel

					// Reset the wiggler status
					wiggler.isWiggling.Store(false)

				case WiggleEventStart: // Is it a START event?

					// Handle wiggler already running...
					if wiggler.isWiggling.Load() {
						// Cancel the running wiggler
						wcancel()

						// Wait for it to finish
						wg.Wait()

						// ...and continue on to start a new one
					}

					// Set to running...
					wiggler.isWiggling.Store(true)

					// Start a new wiggle...
					wg.Add(1) // Increment the wait group counter
					go func() {
						// Decrement the wait group counter
						defer wg.Done()

						// Create a new context for the wiggle event
						wctx, wcancel = context.WithCancel(ctx)
						defer wcancel()

						// Create a random number generator
						rs := rand.New(rand.NewSource(time.Now().UnixNano()))
						rng := rand.New(rs)

						// Default robotgo settings
						rbgoHigh := 1.0

						// Create a ticker to run for each wiggele
						moveSpeed := float64(wiggler.GetMoveSpeed())
						waitTime := float64(wiggler.GetWaitTime())

						d := time.Duration(moveSpeed+waitTime) * time.Second
						wiggleTicker := time.NewTicker(d)

						defer wiggleTicker.Stop() // Stop the ticker when the wiggler is done

						for {
							select {
							case <-wctx.Done(): // If the wiggle event is cancelled...
								return // ...then exit the wiggle event

							case <-wiggleTicker.C: // If the wiggle event is not cancelled...
								// ...then run the wiggler

								// Get the screen size
								width, height := robotgo.GetScreenSize()

								// Pick a random position
								x := rng.Intn(width)
								y := rng.Intn(height)

								// Move the mouse to the random position
								robotgo.MoveSmooth(x, y, rbgoHigh, moveSpeed)
							}
						}
					}()
				}
			}
		}

		wg.Wait() // Wait for any wiggle events to finish...
	}()

	// Run the app
	err := wails.Run(&options.App{
		Title:         "The Wiggler",
		DisableResize: true,
		Width:         WindowWidth,
		Height:        WindowSmallHeight,
		Assets:        assets,
		Bind: []interface{}{
			wiggler,
		},
		OnStartup:  wiggler.OnStartup,
		OnShutdown: wiggler.OnShutdown,
		Mac: &mac.Options{
			About: &mac.AboutInfo{
				Title:   "The Wiggler",
				Message: fmt.Sprintf("Version: %s\n(c) 2022 Austin Poor\n%s", version, github),
				Icon:    macIcon,
			},
		},
	})
	if err != nil {
		log.Panic(err)
	}
}
