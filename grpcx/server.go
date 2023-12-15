package grpcx

import (
	"net"

	"google.golang.org/grpc"
)

/**
   @author：biguanqun
   @since： 2023/12/15
   @desc：
**/

type Server struct {
	*grpc.Server
	Port    string
	Network string
}

func (s *Server) Serve() error {
	l, err := net.Listen(s.Network, s.Port)
	if err != nil {
		panic(err)
	}
	return s.Server.Serve(l)
}
