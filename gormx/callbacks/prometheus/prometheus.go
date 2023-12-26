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

package prometheus

import (
	"time"

	"gorm.io/gorm"

	promsdk "github.com/prometheus/client_golang/prometheus"
)

type Callbacks struct {
	Namespace  string
	Subsystem  string
	Name       string
	InstanceID string
	Help       string
	vector     *promsdk.SummaryVec
}

func (c *Callbacks) Register(db *gorm.DB) error {
	vector := promsdk.NewSummaryVec(
		promsdk.SummaryOpts{
			Namespace: c.Namespace,
			Subsystem: c.Subsystem,
			Name:      c.Name,
			Help:      c.Help,
			ConstLabels: map[string]string{
				"db_name":     db.Name(),
				"instance_id": c.InstanceID,
			},
			Objectives: map[float64]float64{
				0.5:   0.01,
				0.9:   0.01,
				0.99:  0.005,
				0.999: 0.0001,
			},
		}, []string{"typ", "table"})
	promsdk.MustRegister(vector)
	c.vector = vector

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
	return nil
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
