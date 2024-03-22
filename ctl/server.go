package ctl

import (
	"context"
	"net"

	pb "github.com/f-taxes/csv_import/proto"
	"github.com/kataras/golog"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PluginCtl struct {
	pb.UnimplementedPluginCtlServer
}

func (s *PluginCtl) UpdateTransactions(ctx context.Context, job *pb.TxUpdate) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func Start(address string) {
	srv := &PluginCtl{}
	lis, err := net.Listen("tcp", address)
	if err != nil {
		golog.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterPluginCtlServer(s, srv)
	golog.Infof("Ctl server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		golog.Fatalf("failed to serve: %v", err)
	}
}
