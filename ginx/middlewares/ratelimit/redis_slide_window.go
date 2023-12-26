/*
   Copyright 2023 bgq98

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
