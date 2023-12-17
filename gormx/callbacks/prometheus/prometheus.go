package prometheus

import (
	"time"

	"gorm.io/gorm"

	promsdk "github.com/prometheus/client_golang/prometheus"
)

/**
   @author：biguanqun
   @since： 2023/12/17
   @desc：
**/

type Callbacks struct {
	vector *promsdk.SummaryVec
}

func NewCallbacks() *Callbacks {
	vector := promsdk.NewSummaryVec(promsdk.SummaryOpts{
		Namespace: "bgq",
		Subsystem: "webook",
		Name:      "gorm_query_time",
		Help:      "统计 gorm 语句的执行时间",
		ConstLabels: map[string]string{
			"db": "webook",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.9:   0.01,
			0.99:  0.005,
			0.999: 0.0001,
		},
	}, []string{"typ", "table"})

	pcb := &Callbacks{
		vector: vector,
	}
	promsdk.MustRegister(vector)
	return pcb
}

func (c *Callbacks) registerAll(db *gorm.DB) {
	err := db.Callback().Create().Before("*").
		Register("prometheus_create_before", c.before())
	if err != nil {
		panic(err)
	}

	err = db.Callback().Create().After("*").
		Register("prometheus_create_after", c.after("create"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Update().Before("*").
		Register("prometheus_update_before", c.before())
	if err != nil {
		panic(err)
	}

	err = db.Callback().Update().After("*").
		Register("prometheus_update_after", c.after("update"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Delete().Before("*").
		Register("prometheus_delete_before", c.before())
	if err != nil {
		panic(err)
	}

	err = db.Callback().Delete().After("*").
		Register("prometheus_delete_after", c.after("delete"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Raw().Before("*").
		Register("prometheus_raw_before", c.before())
	if err != nil {
		panic(err)
	}

	err = db.Callback().Raw().After("*").
		Register("prometheus_raw_after", c.after("raw"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Row().Before("*").
		Register("prometheus_row_before", c.before())
	if err != nil {
		panic(err)
	}

	err = db.Callback().Row().After("*").
		Register("prometheus_row_after", c.after("row"))
	if err != nil {
		panic(err)
	}

}

func (c *Callbacks) before() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		// 真正逻辑的位置
		startTime := time.Now()
		db.Set("start_time", startTime)
	}
}

func (c *Callbacks) after(typ string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("start_time")
		startTime, ok := val.(time.Time)
		if !ok {
			return
		}
		table := db.Statement.Table
		if table == "" {
			table = "unknown"
		}
		c.vector.WithLabelValues(typ, table).Observe(float64(time.Since(startTime).Milliseconds()))
	}
}
