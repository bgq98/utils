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

package grpcx

import (
	"context"
	"net"
	"strconv"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"

	"github.com/bgq98/utils/logger"
	"github.com/bgq98/utils/next"
)

type Server struct {
	*grpc.Server
	Port        int
	EtcdAddrs   []string
	EtcdTTL     int64
	EtcdClient  *clientv3.Client
	etcdManager endpoints.Manager
	etcdKey     string
	cancel      func()
	Name        string
	L           logger.Logger
}

func (s *Server) Serve() error {
	l, err := net.Listen("tcp", ":"+strconv.Itoa(s.Port))
	if err != nil {
		return err
	}
	err = s.register()
	if err != nil {
		return err
	}
	return s.Server.Serve(l)
}

// 使用 etcd 作为注册中心
func (s *Server) register() error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: s.EtcdAddrs,
	})
	if err != nil {
		return err
	}
	s.EtcdClient = cli
	serviceName := "service/" + s.Name
	em, err := endpoints.NewManager(cli, serviceName)
	if err != nil {
		return err
	}
	addr := next.GetOutboundIp() + ":" + strconv.Itoa(s.Port)
	key := "service/" + s.Name + "/" + addr
	s.etcdKey = key
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	leaseResp, err := cli.Grant(ctx, s.EtcdTTL)

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr: addr,
	}, clientv3.WithLease(leaseResp.ID))

	kaCtx, kaCancel := context.WithCancel(context.Background())
	s.cancel = kaCancel
	ch, err := cli.KeepAlive(kaCtx, leaseResp.ID)
	if err != nil {
		return err
	}
	go func() {
		for kaResp := range ch {
			s.L.Debug(kaResp.String())
		}
	}()
	return nil
}

func (s *Server) Close() error {
	if s.cancel != nil {
		s.cancel()
	}
	if s.etcdManager != nil {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err := s.etcdManager.DeleteEndpoint(ctx, s.etcdKey)
		if err != nil {
			return err
		}
	}
	err := s.EtcdClient.Close()
	if err != nil {
		return err
	}
	s.Server.GracefulStop()
	return nil
}
