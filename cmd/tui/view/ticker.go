/*
Copyright 2024 Nokia.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package view

import (
	"time"

	"github.com/rivo/tview"
)

type ticker struct {
	app       *tview.Application
	frequency time.Duration
	f         func()

	stopch chan struct{}
	ticker *time.Ticker
}

func NewTicker(app *tview.Application, frequency time.Duration, f func()) *ticker {
	return &ticker{
		app:       app,
		frequency: frequency,
		f:         f,
		stopch:    make(chan struct{}),
	}
}

func (r *ticker) start() {
	// call the function
	r.f()
	r.ticker = time.NewTicker(r.frequency)
	go func() {
		for {
			select {
			case <-r.stopch:
				return
			case <-r.ticker.C:
				//fmt.Println("tick")
				r.app.QueueUpdateDraw(func() {
					r.f()
				})
			}
		}
	}()
}

func (r *ticker) stop() {
	if r.ticker != nil {
		r.ticker.Stop()
	}
	r.stopch <- struct{}{}
}
