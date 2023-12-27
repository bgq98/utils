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

package ginx

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/bgq98/utils/logger"
)

/**
   @author：biguanqun
   @since： 2023/10/16
   @desc：  统一处理web层日志
**/

var log logger.Logger = logger.NewNoOpLogger()

func SetLogger(l logger.Logger) {
	log = l
}

var vector *prometheus.CounterVec

func InitCounter(opt prometheus.CounterOpts) {
	vector = prometheus.NewCounterVec(opt, []string{"code"})
	prometheus.MustRegister(vector)
}

// WrapClaims 复制粘贴
func WrapClaims(fn func(*gin.Context, UserClaims) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 可以用包变量来配置,这里因为泛型的限制,只能用包变量
		rawval, ok := ctx.Get("claims")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			log.Error("获取claims失败",
				logger.String("path", ctx.Request.URL.Path))
			return
		}

		// 注意:这里要求放进去 ctx 的不能是指针 *UserClaims
		claims, ok := rawval.(UserClaims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			log.Error("获取claims失败",
				logger.String("path", ctx.Request.URL.Path))
			return
		}
		res, err := fn(ctx, claims)
		if err != nil {
			log.Error("执行业务逻辑失败",
				logger.Error(err))
		}
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		ctx.JSON(http.StatusOK, res)
	}
}

// WrapClaimsAndReq 如果做成中间件来源出去，那么直接耦合 UserClaims 也是不好的。
func WrapClaimsAndReq[Req interface{}](fn func(*gin.Context, Req, UserClaims) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			log.Error("解析请求失败", logger.Error(err))
			return
		}
		// 可以用包变量来配置，还是那句话，因为泛型的限制，这里只能用包变量
		rawVal, ok := ctx.Get("claims")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			log.Error("无法获得 claims",
				logger.String("path", ctx.Request.URL.Path))
			return
		}
		// 注意，这里要求放进去 ctx 的不能是*UserClaims，这是常见的一个错误
		claims, ok := rawVal.(UserClaims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			log.Error("无法获得 claims",
				logger.String("path", ctx.Request.URL.Path))
			return
		}
		res, err := fn(ctx, req, claims)
		if err != nil {
			log.Error("执行业务逻辑失败",
				logger.Error(err))
		}
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		ctx.JSON(http.StatusOK, res)
	}
}

// WrapReq
func WrapReq[Req interface{}](fn func(*gin.Context, Req) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			log.Error("解析请求失败", logger.Error(err))
			return
		}

		res, err := fn(ctx, req)
		if err != nil {
			log.Error("执行业务逻辑失败",
				logger.Error(err))
		}
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		ctx.JSON(http.StatusOK, res)
	}
}
