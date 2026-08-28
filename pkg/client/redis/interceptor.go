package redis

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/redis/go-redis/v9"

	prome "github.com/weijinwen/jupiter/pkg/core/metric"
	"github.com/weijinwen/jupiter/pkg/util/xstring"
	"github.com/weijinwen/jupiter/pkg/xlog"
)

type redigoContextKeyType struct{}

var ctxBegKey = redigoContextKeyType{}

type interceptor struct {
	beforeProcess         func(ctx context.Context, cmd redis.Cmder) (context.Context, error)
	afterProcess          func(ctx context.Context, cmd redis.Cmder) error
	beforeProcessPipeline func(ctx context.Context, cmds []redis.Cmder) (context.Context, error)
	afterProcessPipeline  func(ctx context.Context, cmds []redis.Cmder) error
}

func (i *interceptor) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (i *interceptor) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		ctx, err := i.beforeProcess(ctx, cmd)
		if err != nil {
			return err
		}
		err = next(ctx, cmd)
		if err != nil {
			return err
		}
		return i.afterProcess(ctx, cmd)
	}
}

func (i *interceptor) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		ctx, err := i.beforeProcessPipeline(ctx, cmds)
		if err != nil {
			return err
		}
		err = next(ctx, cmds)
		if err != nil {
			return err
		}
		return i.afterProcessPipeline(ctx, cmds)
	}
}

func newInterceptor(compName string, config *Config, logger *xlog.Logger) *interceptor {
	return &interceptor{
		beforeProcess: func(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
			return ctx, nil
		},
		afterProcess: func(ctx context.Context, cmd redis.Cmder) error {
			return nil
		},
		beforeProcessPipeline: func(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
			return ctx, nil
		},
		afterProcessPipeline: func(ctx context.Context, cmds []redis.Cmder) error {
			return nil
		},
	}
}

func (i *interceptor) setBeforeProcess(p func(ctx context.Context, cmd redis.Cmder) (context.Context, error)) *interceptor {
	i.beforeProcess = p
	return i
}

func (i *interceptor) setAfterProcess(p func(ctx context.Context, cmd redis.Cmder) error) *interceptor {
	i.afterProcess = p
	return i
}

func (i *interceptor) setBeforeProcessPipeline(p func(ctx context.Context, cmds []redis.Cmder) (context.Context, error)) *interceptor {
	i.beforeProcessPipeline = p
	return i
}

func (i *interceptor) setAfterProcessPipeline(p func(ctx context.Context, cmds []redis.Cmder) error) *interceptor {
	i.afterProcessPipeline = p
	return i
}
func fixedInterceptor(compName string, addr string, config *Config, logger *xlog.Logger) *interceptor {
	return newInterceptor(compName, config, logger).
		setBeforeProcess(func(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
			return context.WithValue(ctx, ctxBegKey, time.Now()), nil
		}).
		setBeforeProcessPipeline(func(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
			return context.WithValue(ctx, ctxBegKey, time.Now()), nil
		}).setAfterProcess(func(ctx context.Context, cmd redis.Cmder) error {
		cost := time.Since(ctx.Value(ctxBegKey).(time.Time))

		if config.SlowLogThreshold > time.Duration(0) && cost > config.SlowLogThreshold {
			logger.Error("slow",
				xlog.FieldErr(errors.New("redis slow command")),
				xlog.FieldName(cmd.Name()),
				xlog.FieldAddr(addr),
				xlog.FieldCost(cost))
		}
		return nil
	}).setAfterProcessPipeline(func(ctx context.Context, cmds []redis.Cmder) error {
		cost := time.Since(ctx.Value(ctxBegKey).(time.Time))

		if config.SlowLogThreshold > time.Duration(0) && cost > config.SlowLogThreshold {
			logger.Error("slow",
				xlog.FieldErr(errors.New("redis slow command")),
				xlog.FieldType("pipeline"),
				xlog.FieldName(getCmdsName(cmds)),
				xlog.FieldAddr(addr),
				xlog.FieldCost(cost))
		}
		return nil
	})

}
func debugInterceptor(compName string, addr string, config *Config, logger *xlog.Logger) *interceptor {

	return newInterceptor(compName, config, logger).
		setAfterProcess(func(ctx context.Context, cmd redis.Cmder) error {
			cost := time.Since(ctx.Value(ctxBegKey).(time.Time))
			err := cmd.Err()
			fmt.Println(xstring.CallerName(6))
			fmt.Printf("[redis ] %s (%s) :\n", addr, cost) // nolint
			if err != nil {
				fmt.Print(color.RedString("# %s %+v, ERR=(%s)\n\n", cmd.Name(), cmd.Args(), err.Error())) // nolint
			} else {
				fmt.Print(color.YellowString("# %s %+v: %s\n\n", cmd.Name(), cmd.Args(), response(cmd))) // nolint
			}
			return nil
		}).
		setAfterProcessPipeline(func(ctx context.Context, cmds []redis.Cmder) error {
			cost := time.Since(ctx.Value(ctxBegKey).(time.Time))
			fmt.Println(xstring.CallerName(8))
			fmt.Printf("[redis pipeline] %s (%s) :\n", addr, cost) // nolint
			for _, cmd := range cmds {
				err := cmd.Err()
				if err != nil {
					fmt.Print(color.RedString("* %s %+v, ERR=<%s>\n", cmd.Name(), cmd.Args(), err.Error())) // nolint
				} else {
					fmt.Print(color.YellowString("* %s %+v: %s\n", cmd.Name(), cmd.Args(), response(cmd))) // nolint
				}
			}
			fmt.Print("  \n") // nolint
			return nil
		})
}
func metricInterceptor(compName string, addr string, config *Config, logger *xlog.Logger) *interceptor {

	return newInterceptor(compName, config, logger).
		setAfterProcess(func(ctx context.Context, cmd redis.Cmder) error {
			cost := time.Since(ctx.Value(ctxBegKey).(time.Time))
			err := cmd.Err()
			name := strings.ToUpper(cmd.Name())
			prome.LibHandleHistogram.WithLabelValues(prome.TypeRedis, name, addr).Observe(cost.Seconds())
			if err != nil {
				if errors.Is(err, redis.Nil) {
					prome.LibHandleCounter.Inc(prome.TypeRedis, name, addr, "Empty")
				}
				prome.LibHandleCounter.Inc(prome.TypeRedis, name, addr, "Error")
			}
			prome.LibHandleCounter.Inc(prome.TypeRedis, name, addr, "OK")
			return nil
		}).setAfterProcessPipeline(func(ctx context.Context, cmds []redis.Cmder) error {
		cost := time.Since(ctx.Value(ctxBegKey).(time.Time))
		names := strings.ToUpper(getCmdsName(cmds))
		prome.LibHandleHistogram.WithLabelValues(prome.TypeRedis, names, addr).Observe(cost.Seconds())
		for _, cmd := range cmds {
			name := strings.ToUpper(cmd.Name())
			if cmd.Err() != nil {
				if errors.Is(cmd.Err(), redis.Nil) {
					prome.LibHandleCounter.Inc(prome.TypeRedis, name, addr, "Empty")
				}
				prome.LibHandleCounter.Inc(prome.TypeRedis, name, addr, "Error")
			}
			prome.LibHandleCounter.Inc(prome.TypeRedis, name, addr, "OK")
		}
		return nil
	})
}
func accessInterceptor(compName string, addr string, config *Config, logger *xlog.Logger) *interceptor {
	return newInterceptor(compName, config, logger).
		setAfterProcess(func(ctx context.Context, cmd redis.Cmder) error {
			var fields = make([]xlog.Field, 0, 15)
			var err = cmd.Err()
			cost := time.Since(ctx.Value(ctxBegKey).(time.Time))
			fields = append(fields, xlog.FieldKey(compName),
				xlog.FieldMethod(cmd.Name()),
				xlog.FieldAddr(addr),
				xlog.Any("req", cmd.Args()),
				xlog.FieldCost(cost))

			// error
			if err != nil {
				fields = append(fields, xlog.FieldErr(err))
				if errors.Is(err, redis.Nil) {
					logger.Warn("access", fields...)
					return nil
				}
				logger.Error("access", fields...)
				return nil
			}
			fields = append(fields, xlog.Any("res", response(cmd)))
			logger.Info("access", fields...)

			return nil
		},
		).setAfterProcessPipeline(func(ctx context.Context, cmds []redis.Cmder) error {
		cost := time.Since(ctx.Value(ctxBegKey).(time.Time))

		for _, cmd := range cmds {
			var fields = make([]xlog.Field, 0, 15)
			var err = cmd.Err()
			fields = append(fields, xlog.FieldKey(compName),
				xlog.FieldType("pipeline"),
				xlog.FieldMethod(cmd.Name()),
				xlog.Any("req", cmd.Args()),
				xlog.FieldCost(cost))

			// error
			if err != nil {
				fields = append(fields, xlog.FieldErr(err))
				if errors.Is(err, redis.Nil) {
					logger.Warn("access", fields...)
					continue
				}
				logger.Error("access", fields...)
				continue
			}
			fields = append(fields, xlog.Any("res", response(cmd)))
			logger.Info("access", fields...)

			continue
		}
		return nil
	})
}

func response(cmd redis.Cmder) string {
	switch recv := cmd.(type) {
	case *redis.Cmd:
		return fmt.Sprintf("%v", recv.Val())
	case *redis.StringCmd:
		return fmt.Sprintf("%v", recv.Val())
	case *redis.StatusCmd:
		return fmt.Sprintf("%v", recv.Val())
	case *redis.IntCmd:
		return fmt.Sprintf("%v", recv.Val())
	case *redis.DurationCmd:
		return fmt.Sprintf("%v", recv.Val())
	case *redis.BoolCmd:
		return fmt.Sprintf("%v", recv.Val())
	case *redis.CommandsInfoCmd:
		return fmt.Sprintf("%v", recv.Val())
	case *redis.StringSliceCmd:
		return fmt.Sprintf("%v", recv.Val())
	default:
		return ""
	}
}
func getCmdsName(cmds []redis.Cmder) string {
	cmdNameMap := map[string]bool{}
	cmdName := []string{}
	for _, cmd := range cmds {
		if !cmdNameMap[cmd.Name()] {
			cmdName = append(cmdName, cmd.Name())
			cmdNameMap[cmd.Name()] = true
		}
	}
	return strings.Join(cmdName, "_")
}
