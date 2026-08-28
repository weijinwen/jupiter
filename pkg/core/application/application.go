// Copyright 2021 rex lv
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package application

import (
	"context"
	"fmt"
	"sync"
	"time"

	//go-lint
	_ "github.com/weijinwen/jupiter/pkg/conf/datasource/file"
	_ "github.com/weijinwen/jupiter/pkg/conf/datasource/http"
	_ "github.com/weijinwen/jupiter/pkg/core/autoproc"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
	"github.com/weijinwen/jupiter/pkg/conf"
	"github.com/weijinwen/jupiter/pkg/core/ecode"
	"github.com/weijinwen/jupiter/pkg/core/hooks"
	"github.com/weijinwen/jupiter/pkg/core/signals"
	"github.com/weijinwen/jupiter/pkg/flag"
	"github.com/weijinwen/jupiter/pkg/server"
	"github.com/weijinwen/jupiter/pkg/util/xcycle"
	"github.com/weijinwen/jupiter/pkg/util/xdebug"
	"github.com/weijinwen/jupiter/pkg/util/xgo"
	"github.com/weijinwen/jupiter/pkg/worker"
	"github.com/weijinwen/jupiter/pkg/xlog"
	"golang.org/x/sync/errgroup"
)

// Application is the framework's instance, it contains the servers, workers, client and configuration settings.
// Create an instance of Application, by using &Application{}
type Application struct {
	cycle    *xcycle.Cycle
	smu      *sync.RWMutex
	initOnce sync.Once
	// startupOnce  sync.Once
	stopOnce sync.Once
	servers  []server.Server
	logger   *xlog.Logger
	// hooks        map[uint32]*xdefer.DeferStack
	configParser conf.Unmarshaller
	disableMap   map[Disable]bool
	HideBanner   bool
	stopped      chan struct{}
}

// New create a new Application instance
func New(fns ...func() error) (*Application, error) {
	app := &Application{}
	if err := app.Startup(fns...); err != nil {
		return nil, err
	}
	return app, nil
}

func DefaultApp() *Application {
	app := &Application{}
	app.initialize()
	return app
}

// run hooks
func (app *Application) runHooks(stage hooks.Stage) {
	hooks.Do(stage)
}

// RegisterHooks register a stage Hook
func (app *Application) RegisterHooks(stage hooks.Stage, fns ...func()) {
	hooks.Register(stage, fns...)
}

func (app *Application) RegisterWorker(name string, fn func(ctx context.Context) error) {
	worker.RegisterFunc(name, fn)
}

// initialize application
func (app *Application) initialize() {
	app.initOnce.Do(func() {
		//assign
		app.cycle = xcycle.NewCycle()
		app.smu = &sync.RWMutex{}
		app.servers = make([]server.Server, 0)
		app.logger = xlog.Jupiter()
		app.configParser = toml.Unmarshal
		app.disableMap = make(map[Disable]bool)
		app.stopped = make(chan struct{})
		//private method

		_ = app.parseFlags()
		_ = app.printBanner()
		// app.initLogger()
	})
}

// Startup ..
func (app *Application) Startup(fns ...func() error) error {
	app.initialize()
	// if err := app.startup(); err != nil {
	// 	return err
	// }

	return xgo.SerialUntilError(fns...)()
}

// Defer ..
// Deprecated: use AfterStop instead
// func (app *Application) Defer(fns ...func() error) {
// 	app.AfterStop(fns...)
// }

// BeforeStop hook
// Deprecated: use RegisterHooks instead
// func (app *Application) BeforeStop(fns ...func() error) {
// 	app.RegisterHooks(StageBeforeStop, fns...)
// }

// AfterStop hook
// Deprecated: use RegisterHooks instead
// func (app *Application) AfterStop(fns ...func() error) {
// 	app.RegisterHooks(StageAfterStop, fns...)
// }

// Serve start server
func (app *Application) Serve(s ...server.Server) error {
	app.smu.Lock()
	defer app.smu.Unlock()
	app.servers = append(app.servers, s...)
	return nil
}

// SetGovernor set governor addr (default 127.0.0.1:0)
// Deprecated
//func (app *Application) SetGovernor(addr string) {
//	app.governorAddr = addr
//}

// Run run application
func (app *Application) Run(servers ...server.Server) error {
	app.smu.Lock()
	app.servers = append(app.servers, servers...)
	app.smu.Unlock()

	hooks.Do(hooks.Stage_BeforeRun)

	app.waitSignals() //start signal listen task in goroutine
	defer app.clean()

	// start servers and govern server
	app.cycle.Run(app.startServers)

	//blocking and wait quit
	if err := <-app.cycle.Wait(); err != nil {
		app.logger.Error("jupiter shutdown with error", xlog.FieldMod(ecode.ModApp), xlog.FieldErr(err))
		return err
	}
	app.logger.Info("shutdown jupiter, bye!", xlog.FieldMod(ecode.ModApp))
	return nil
}

// clean after app quit
func (app *Application) clean() {
	_ = xlog.Default().Sync()
	_ = xlog.Jupiter().Sync()
}

// Stop application immediately after necessary cleanup
func (app *Application) Stop() (err error) {
	app.stopOnce.Do(func() {
		app.stopped <- struct{}{}
		app.runHooks(hooks.Stage_BeforeStop)
		//stop servers
		for _, s := range app.servers {
			func(s server.Server) {
				app.smu.RLock()
				app.logger.Info("exit server", xlog.FieldMod(ecode.ModApp), xlog.FieldEvent("stop"), xlog.FieldName(s.Info().Name), xlog.FieldAddr(s.Info().Address))

				app.cycle.Run(s.Stop)
				app.smu.RUnlock()
			}(s)
		}
		<-app.cycle.Done()
		// run hook
		app.runHooks(hooks.Stage_AfterStop)
		app.cycle.Close()
	})
	return
}

// GracefulStop application after necessary cleanup
func (app *Application) GracefulStop(ctx context.Context) (err error) {
	app.stopOnce.Do(func() {
		app.stopped <- struct{}{}
		app.runHooks(hooks.Stage_BeforeStop)
		//stop servers
		for _, s := range app.servers {
			func(s server.Server) {
				app.cycle.Run(func() error {
					app.smu.RLock()
					defer app.smu.RUnlock()
					app.logger.Info("exit server", xlog.FieldMod(ecode.ModApp), xlog.FieldEvent("graceful stop"), xlog.FieldName(s.Info().Name), xlog.FieldAddr(s.Info().Address))

					return s.GracefulStop(ctx)
				})
			}(s)
		}
		<-app.cycle.Done()
		// run hooks
		app.runHooks(hooks.Stage_AfterStop)
		app.cycle.Close()
	})
	return err
}

// waitSignals wait signal
func (app *Application) waitSignals() {
	app.logger.Info("init listen signal", xlog.FieldMod(ecode.ModApp), xlog.FieldEvent("init"))
	signals.Shutdown(func(grace bool) { //when get shutdown signal
		if grace {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			_ = app.GracefulStop(ctx)
		} else {
			_ = app.Stop()
		}
	})
}

func (app *Application) startServers() error {
	var eg errgroup.Group
	// start multi servers
	app.smu.Lock()
	for _, s := range app.servers {
		eg.Go(func() (err error) {
			time.AfterFunc(time.Second, func() {
				app.logger.Info("start server", xlog.FieldMod(ecode.ModApp), xlog.FieldEvent("init"), xlog.FieldName(s.Info().Name), xlog.FieldAddr(s.Info().Address), xlog.Any("scheme", s.Info().Scheme))
			})
			err = s.Serve()
			return
		})
	}

	app.smu.Unlock()
	return eg.Wait()
}

// parseFlags init
func (app *Application) parseFlags() error {
	if app.isDisable(DisableParserFlag) {
		app.logger.Info("parseFlags disable", xlog.FieldMod(ecode.ModApp))
		return nil
	}

	return flag.Parse()
}

func (app *Application) isDisable(d Disable) bool {
	b, ok := app.disableMap[d]
	if !ok {
		return false
	}
	return b
}

// printBanner init
func (app *Application) printBanner() error {
	if app.HideBanner {
		return nil
	}

	if xdebug.IsTestingMode() {
		return nil
	}

	const banner = `
   (_)_   _ _ __ (_) |_ ___ _ __
   | | | | | '_ \| | __/ _ \ '__|
   | | |_| | |_) | | ||  __/ |
  _/ |\__,_| .__/|_|\__\___|_|
 |__/      |_|

 Welcome to jupiter, starting application ...
`
	fmt.Println(color.GreenString(banner))
	return nil
}
