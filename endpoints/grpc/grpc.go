// Package grpc is the gRPC wiring for the greet domain. It contains:
//
//  1. Wrap, the generic adapter that turns a (ctx, In) -> (Out, error)
//     function into a function with the unary handler shape grpc-go expects.
//  2. Per-RPC decode and encode functions.
//  3. Server, which satisfies the generated greeterpb.GreeterServer interface.
//  4. Error mapping from greet error codes to gRPC status codes.
//
// Package name grpc shadows google.golang.org/grpc for outside importers;
// alias this package as egrpc at the call site and leave google.golang.org/grpc
// plain.
package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rednafi/examples/endpoints/greet"
	pb "github.com/rednafi/examples/endpoints/grpc/api"
)

type validator interface{ Validate() error }

// Wrap turns a domain function into a unary gRPC handler. The decode and
// encode callbacks are the only place protobuf types appear.
func Wrap[WireIn, In, Out, WireOut any](
	decode func(WireIn) (In, error),
	fn func(context.Context, In) (Out, error),
	encode func(Out) (WireOut, error),
) func(context.Context, WireIn) (WireOut, error) {
	return func(ctx context.Context, wireIn WireIn) (WireOut, error) {
		var zero WireOut

		in, err := decode(wireIn)
		if err != nil {
			return zero, statusErr(err)
		}

		if v, ok := any(in).(validator); ok {
			if err := v.Validate(); err != nil {
				return zero, statusErr(err)
			}
		}

		out, err := fn(ctx, in)
		if err != nil {
			return zero, statusErr(err)
		}

		return encode(out)
	}
}

func decodeGreet(req *pb.GreetRequest) (greet.GreetIn, error) {
	return greet.GreetIn{
		Name:      req.GetName(),
		Formality: int(req.GetFormality()),
	}, nil
}

func encodeGreet(out greet.GreetOut) (*pb.GreetResponse, error) {
	return &pb.GreetResponse{Message: out.Message}, nil
}

func decodeSubscribe(req *pb.SubscribeRequest) (greet.SubscribeIn, error) {
	return greet.SubscribeIn{
		Email:     req.GetEmail(),
		Formality: int(req.GetFormality()),
	}, nil
}

func encodeSubscribe(out greet.SubscribeOut) (*pb.SubscribeResponse, error) {
	return &pb.SubscribeResponse{Id: out.ID}, nil
}

// Server satisfies pb.GreeterServer. The constructor builds wrapped functions
// once; each method body forwards to the matching wrapped function.
type Server struct {
	pb.UnimplementedGreeterServer

	greet     func(context.Context, *pb.GreetRequest) (*pb.GreetResponse, error)
	subscribe func(context.Context, *pb.SubscribeRequest) (*pb.SubscribeResponse, error)
}

func NewServer(svc *greet.Service) *Server {
	return &Server{
		greet:     Wrap(decodeGreet, svc.Greet, encodeGreet),
		subscribe: Wrap(decodeSubscribe, svc.Subscribe, encodeSubscribe),
	}
}

func (s *Server) Greet(ctx context.Context, req *pb.GreetRequest) (*pb.GreetResponse, error) {
	return s.greet(ctx, req)
}

func (s *Server) Subscribe(ctx context.Context, req *pb.SubscribeRequest) (*pb.SubscribeResponse, error) {
	return s.subscribe(ctx, req)
}

// Register attaches the Greeter server to a gRPC server.
func Register(srv *grpc.Server, svc *greet.Service) {
	pb.RegisterGreeterServer(srv, NewServer(svc))
}

func statusErr(err error) error {
	var de *greet.Error
	if !errors.As(err, &de) {
		return status.Error(codes.Internal, "internal error")
	}
	return status.Error(codeFor(de.Code), de.Message)
}

func codeFor(c greet.Code) codes.Code {
	switch c {
	case greet.CodeInvalidArgument:
		return codes.InvalidArgument
	case greet.CodeUnauthenticated:
		return codes.Unauthenticated
	case greet.CodePermissionDenied:
		return codes.PermissionDenied
	case greet.CodeNotFound:
		return codes.NotFound
	case greet.CodeAlreadyExists:
		return codes.AlreadyExists
	default:
		return codes.Internal
	}
}
