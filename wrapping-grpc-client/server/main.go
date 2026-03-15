package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/rednafi/examples/wrapping-grpc-client/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type kvServer struct {
	api.UnimplementedKVServer
	mu   sync.RWMutex
	data map[string][]byte
}

func (s *kvServer) Put(
	ctx context.Context, req *api.PutRequest,
) (*api.PutResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[req.Key] = req.Value
	return &api.PutResponse{}, nil
}

func (s *kvServer) Get(
	ctx context.Context, req *api.GetRequest,
) (*api.GetResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[req.Key]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "key %q not found", req.Key)
	}
	return &api.GetResponse{Value: val, Found: true}, nil
}

func (s *kvServer) Delete(
	ctx context.Context, req *api.DeleteRequest,
) (*api.DeleteResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, req.Key)
	return &api.DeleteResponse{}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	api.RegisterKVServer(srv, &kvServer{data: make(map[string][]byte)})
	fmt.Println("listening on :9090")
	if err := srv.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
