package ratelimit

import "context"

/**
   @author：biguanqun
   @since： 2023/9/21
   @desc：
**/

//go:generate mockgen -source=./types.go -package=limitmocks -destination=mocks/ratelimit.mock.go Limiter
type Limiter interface {

	/*Limit 就是有没有触发限流, key 就是限流对象
	  如果是对ip限流,那 key 就是ip地址
	  bool 代表是否限流, true 就是要限流
	  error 代表限流器本身是否有错误 */
	Limit(ctx context.Context, key string) (bool, error)
}
