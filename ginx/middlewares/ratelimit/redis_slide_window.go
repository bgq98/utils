package ratelimit

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

/**
   @author：biguanqun
   @since： 2023/9/21
   @desc：  Redis 的滑动窗口算法限流器实现
**/

//go:embed slide_window.lua
var luaSlideWidow string

type RedisSlideWindowLimiter struct {
	cmd redis.Cmdable

	// interval 和 rate 组合的意思为
	// interval 内允许 rate 个请求  eg: 1s 内允许 3000 个请求
	interval time.Duration // 窗口大小
	rate     int           // 阈值
}

func NewRedisSlideWindowLimiter(cmd redis.Cmdable,
	interval time.Duration, rate int) Limiter {
	return &RedisSlideWindowLimiter{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}
}

func (r *RedisSlideWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return r.cmd.Eval(ctx, luaSlideWidow, []string{key},
		r.interval.Milliseconds(), r.rate, time.Now().UnixMilli()).Bool()
}
