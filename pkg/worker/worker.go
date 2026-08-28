// Copyright 2022 Douyu
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

package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/weijinwen/jupiter/pkg/xlog"
	"go.uber.org/zap"
)

var (
	// 全局注册表
	registry = &commandRegistry{
		commands: make(map[string]Command),
	}
)

// Command 命令接口
type Command interface {
	// Name 命令名称
	Name() string
	// Run 执行命令，传入上下文支持超时控制
	Run(ctx context.Context) error
}

// CommandFunc 函数类型命令适配器
type CommandFunc func(ctx context.Context) error

func (f CommandFunc) Name() string {
	return "anonymous"
}

func (f CommandFunc) Run(ctx context.Context) error {
	return f(ctx)
}

// commandRegistry 命令注册表
type commandRegistry struct {
	mu       sync.RWMutex
	commands map[string]Command
}

// Register 注册命令（支持重复注册，后面的覆盖前面的）
func (r *commandRegistry) Register(cmd Command) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands[cmd.Name()] = cmd
}

// RegisterFunc 注册函数命令
func (r *commandRegistry) RegisterFunc(name string, fn func(ctx context.Context) error) {
	r.Register(&namedCommand{
		name: name,
		fn:   fn,
	})
}

// Get 获取命令
func (r *commandRegistry) Get(name string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cmd, ok := r.commands[name]
	return cmd, ok
}

// List 列出所有已注册命令
func (r *commandRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.commands))
	for name := range r.commands {
		names = append(names, name)
	}
	return names
}

// Execute 执行指定命令
func (r *commandRegistry) Execute(ctx context.Context, name string) error {
	cmd, ok := r.Get(name)
	if !ok {
		return fmt.Errorf("command not found: %s", name)
	}

	start := time.Now()
	xlog.Jupiter().Info("command start", zap.String("name", name))

	err := cmd.Run(ctx)

	duration := time.Since(start)
	if err != nil {
		xlog.Jupiter().Error("command failed",
			zap.String("name", name),
			zap.Duration("duration", duration),
			zap.Error(err))
		return err
	}

	xlog.Jupiter().Info("command success",
		zap.String("name", name),
		zap.Duration("duration", duration))

	return nil
}

// namedCommand 带名称的命令
type namedCommand struct {
	name string
	fn   func(ctx context.Context) error
}

func (c *namedCommand) Name() string {
	return c.name
}

func (c *namedCommand) Run(ctx context.Context) error {
	return c.fn(ctx)
}

// ============ 公开函数 ===========

// Register 注册命令（包级快捷函数）
func Register(cmd Command) {
	registry.Register(cmd)
}

// RegisterFunc 注册函数命令
func RegisterFunc(name string, fn func(ctx context.Context) error) {
	registry.RegisterFunc(name, fn)
}

// Execute 执行命令
func Execute(ctx context.Context, name string) error {
	return registry.Execute(ctx, name)
}

// List 列出所有命令
func List() []string {
	return registry.List()
}

// ============ 增强功能 ============

// WithTimeout 包装命令，添加超时控制
func WithTimeout(cmd Command, timeout time.Duration) Command {
	return &timeoutCommand{
		Command: cmd,
		timeout: timeout,
	}
}

type timeoutCommand struct {
	Command
	timeout time.Duration
}

func (c *timeoutCommand) Run(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.Command.Run(ctx)
}

// Chain 链式执行多个命令
func Chain(cmds ...Command) Command {
	return &chainCommand{commands: cmds}
}

type chainCommand struct {
	commands []Command
}

func (c *chainCommand) Name() string {
	return "chain"
}

func (c *chainCommand) Run(ctx context.Context) error {
	for _, cmd := range c.commands {
		if err := cmd.Run(ctx); err != nil {
			return fmt.Errorf("chain failed at %s: %w", cmd.Name(), err)
		}
	}
	return nil
}

// ErrStopChain 停止链式执行的错误
var ErrStopChain = errors.New("stop chain execution")

// ChainWithStop 链式执行，支持中途停止
func ChainWithStop(cmds ...Command) Command {
	return &chainCommandWithStop{commands: cmds}
}

type chainCommandWithStop struct {
	commands []Command
}

func (c *chainCommandWithStop) Name() string {
	return "chain-with-stop"
}

func (c *chainCommandWithStop) Run(ctx context.Context) error {
	for _, cmd := range c.commands {
		if err := cmd.Run(ctx); err != nil {
			if errors.Is(err, ErrStopChain) {
				return nil // 正常停止
			}
			return fmt.Errorf("chain failed at %s: %w", cmd.Name(), err)
		}
	}
	return nil
}
