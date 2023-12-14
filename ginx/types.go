package ginx

import "github.com/golang-jwt/jwt/v5"

/**
   @author：biguanqun
   @since： 2023/11/4
   @desc：
**/

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

type UserClaims struct {
	Id        int64
	UserAgent string
	Ssid      string
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	Uid  int64
	Ssid string
	jwt.RegisteredClaims
}
