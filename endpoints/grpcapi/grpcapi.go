// Package grpcapi has the gRPC wiring: a decode and encode function per
// RPC, plus a Server that satisfies the generated EndpointsServer
// interface by forwarding to functions built with transport.WrapGRPC.
package grpcapi

import (
	"context"

	"github.com/rednafi/examples/endpoints/api"
	"github.com/rednafi/examples/endpoints/greet"
	"github.com/rednafi/examples/endpoints/transport"
)

func decodeGreet(req *api.GreetRequest) (greet.GreetIn, error) {
	return greet.GreetIn{
		Name:      req.GetName(),
		Formality: int(req.GetFormality()),
	}, nil
}

func encodeGreet(out greet.GreetOut) (*api.GreetResponse, error) {
	return &api.GreetResponse{Message: out.Message}, nil
}

func decodeCreateUser(req *api.CreateUserRequest) (greet.CreateUserIn, error) {
	return greet.CreateUserIn{
		Email: req.GetEmail(),
		Age:   int(req.GetAge()),
	}, nil
}

func encodeCreateUser(out greet.CreateUserOut) (*api.CreateUserResponse, error) {
	return &api.CreateUserResponse{Id: out.ID}, nil
}

type Server struct {
	api.UnimplementedEndpointsServer

	greet      func(context.Context, *api.GreetRequest) (*api.GreetResponse, error)
	createUser func(context.Context, *api.CreateUserRequest) (*api.CreateUserResponse, error)
}

func NewServer(svc *greet.Service) *Server {
	return &Server{
		greet:      transport.WrapGRPC(decodeGreet, svc.Greet, encodeGreet),
		createUser: transport.WrapGRPC(decodeCreateUser, svc.CreateUser, encodeCreateUser),
	}
}

func (s *Server) Greet(ctx context.Context, req *api.GreetRequest) (*api.GreetResponse, error) {
	return s.greet(ctx, req)
}

func (s *Server) CreateUser(ctx context.Context, req *api.CreateUserRequest) (*api.CreateUserResponse, error) {
	return s.createUser(ctx, req)
}
