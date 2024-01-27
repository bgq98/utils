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

package wrr

import (
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

const name = "custom_wrr"

func init() {
	balancer.Register(base.NewBalancerBuilder(name, &PickerBuilder{}, base.Config{HealthCheck: false}))
}

type PickerBuilder struct {
}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*conn, 0, len(info.ReadySCs))
	// sc  => SubConn
	// sci => SubConnInfo
	for sc, sci := range info.ReadySCs {
		cc := &conn{
			cc: sc,
		}
		md, ok := sci.Address.Metadata.(map[string]interface{})
		if ok {
			weightVal := md["weight"]
			weight, _ := weightVal.(float64)
			cc.weight = int(weight)
			cc.labels = md["labels"].([]string)
		}

		if cc.weight == 0 {
			cc.weight = 10
		}
		cc.currentWeight = cc.weight
		conns = append(conns, cc)
	}
	return &Picker{
		conns: conns,
	}
}

type Picker struct {
	conns []*conn
	lock  sync.Mutex
}

// Pick 在这里实现基于权重的负载均衡算法
func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if len(p.conns) == 0 {
		// 没有候选节点
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	var total int
	var maxCC *conn

	// 计算当前权重
	// label := info.Ctx.Value("label")
	for _, cc := range p.conns {
		// 如果要是 cc 里面的所有标签都不包含这个 label ，就跳过
		total += cc.weight
		cc.currentWeight = cc.currentWeight + cc.weight
		if maxCC == nil || cc.currentWeight > maxCC.currentWeight {
			maxCC = cc
		}
	}
	maxCC.currentWeight = maxCC.currentWeight - total
	return balancer.PickResult{
		SubConn: maxCC.cc,
		Done: func(info balancer.DoneInfo) {
			// 很多动态算法,根据调用结果来调整权重,就在这里
		},
	}, nil
}

type conn struct {
	weight        int
	currentWeight int
	cc            balancer.SubConn
	labels        []string
}
