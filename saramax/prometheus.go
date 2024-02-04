package saramax

import "github.com/prometheus/client_golang/prometheus"

/*
   @Author: biguanqun
   @Time:   2024/2/4 14:15
   @File:   prometheus.go
   @Desc:
*/

var vector *prometheus.CounterVec

func InitCounter(opt prometheus.CounterOpts, topic []string) *prometheus.CounterVec {
	vector = prometheus.NewCounterVec(opt, topic)
	prometheus.MustRegister(vector)
	return vector
}
