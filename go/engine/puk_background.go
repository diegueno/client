// Copyright 2017 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// PerUserKeyBackground runs PerUserKeyUpgrade in the background once in a while.
// It brings users without per-user-keys up to having them.
// Note that this engine is long-lived and potentially has to deal with being
// logged out and logged in as a different user, etc.

package engine

import (
	"sync"
	"time"

	"github.com/keybase/client/go/libkb"
)

var PerUserKeyBackgroundSettings = BackgroundTaskSettings{
	// Wait after starting the app
	Start: 30 * time.Second,
	// When waking up on mobile lots of timers will go off at once. We wait an additional
	// delay so as not to add to that herd and slow down the mobile experience when opening the app.
	WakeUp: 10 * time.Second,
	// Wait between checks
	Interval: 1 * time.Hour,
	// Time limit on each round
	Limit: 5 * time.Minute,
}

// PerUserKeyBackground is an engine.
type PerUserKeyBackground struct {
	libkb.Contextified
	sync.Mutex

	args *PerUserKeyBackgroundArgs
	task *BackgroundTask
}

type PerUserKeyBackgroundArgs struct {
	// Channels used for testing. Normally nil.
	testingMetaCh     chan<- string
	testingRoundResCh chan<- error
}

// NewPerUserKeyBackground creates a PerUserKeyBackground engine.
func NewPerUserKeyBackground(g *libkb.GlobalContext, args *PerUserKeyBackgroundArgs) *PerUserKeyBackground {
	task := NewBackgroundTask(g, &BackgroundTaskArgs{
		Name:     "PerUserKeyBackground",
		F:        PerUserKeyBackgroundRound,
		Settings: PerUserKeyBackgroundSettings,

		testingMetaCh:     args.testingMetaCh,
		testingRoundResCh: args.testingRoundResCh,
	})
	return &PerUserKeyBackground{
		Contextified: libkb.NewContextified(g),
		args:         args,
		// Install the task early so that Shutdown can be called before RunEngine.
		task: task,
	}
}

// Name is the unique engine name.
func (e *PerUserKeyBackground) Name() string {
	return "PerUserKeyBackground"
}

// GetPrereqs returns the engine prereqs.
func (e *PerUserKeyBackground) Prereqs() Prereqs {
	return Prereqs{}
}

// RequiredUIs returns the required UIs.
func (e *PerUserKeyBackground) RequiredUIs() []libkb.UIKind {
	return []libkb.UIKind{}
}

// SubConsumers returns the other UI consumers for this engine.
func (e *PerUserKeyBackground) SubConsumers() []libkb.UIConsumer {
	return []libkb.UIConsumer{&PerUserKeyUpgrade{}}
}

// Run starts the engine.
// Returns immediately, kicks off a background goroutine.
func (e *PerUserKeyBackground) Run(ctx *Context) (err error) {
	return RunEngine(e.task, ctx)
}

func (e *PerUserKeyBackground) Shutdown() {
	e.task.Shutdown()
}

func PerUserKeyBackgroundRound(g *libkb.GlobalContext, ectx *Context) error {
	if !g.Env.GetUpgradePerUserKey() {
		g.Log.CDebugf(ectx.GetNetContext(), "CheckUpgradePerUserKey disabled")
		return nil
	}

	if g.ConnectivityMonitor.IsConnected(ectx.GetNetContext()) == libkb.ConnectivityMonitorNo {
		g.Log.CDebugf(ectx.GetNetContext(), "CheckUpgradePerUserKey giving up offline")
		return nil
	}

	// Do a fast local check to see if our work is done.
	pukring, err := g.GetPerUserKeyring()
	if err == nil {
		if pukring.HasAnyKeys() {
			g.Log.CDebugf(ectx.GetNetContext(), "CheckUpgradePerUserKey already has keys")
			return nil
		}
	}

	arg := &PerUserKeyUpgradeArgs{}
	eng := NewPerUserKeyUpgrade(g, arg)
	err = RunEngine(eng, ectx)
	return err
}
