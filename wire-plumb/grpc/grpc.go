// Package grpc is the gRPC wiring for the greet domain.
//
// Package name grpc shadows google.golang.org/grpc for outside importers;
// alias this package as egrpc at the call site and leave
// google.golang.org/grpc plain.
package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rednafi/examples/wire-plumb/greet"
	pb "github.com/rednafi/examples/wire-plumb/grpc/api"
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
		UserID:    req.GetUserId(),
		Formality: int(req.GetFormality()),
	}, nil
}

func encodeGreet(out greet.GreetOut) (*pb.GreetResponse, error) {
	return &pb.GreetResponse{Message: out.Message}, nil
}

func decodeFarewell(req *pb.FarewellRequest) (greet.FarewellIn, error) {
	return greet.FarewellIn{UserID: req.GetUserId()}, nil
}

func encodeFarewell(out greet.FarewellOut) (*pb.FarewellResponse, error) {
	return &pb.FarewellResponse{Message: out.Message}, nil
}

// Server holds wrapped function values built once in NewServer; each method
// body forwards to the matching wrapped function.
type Server struct {
	pb.UnimplementedGreeterServer

	greet    func(context.Context, *pb.GreetRequest) (*pb.GreetResponse, error)
	farewell func(context.Context, *pb.FarewellRequest) (*pb.FarewellResponse, error)
}

func NewServer(svc *greet.Service) *Server {
	return &Server{
		greet:    Wrap(decodeGreet, svc.Greet, encodeGreet),
		farewell: Wrap(decodeFarewell, svc.Farewell, encodeFarewell),
	}
}

func (s *Server) Greet(ctx context.Context, req *pb.GreetRequest) (*pb.GreetResponse, error) {
	return s.greet(ctx, req)
}

func (s *Server) Farewell(ctx context.Context, req *pb.FarewellRequest) (*pb.FarewellResponse, error) {
	return s.farewell(ctx, req)
}

func Register(srv *grpc.Server, svc *greet.Service) {
	pb.RegisterGreeterServer(srv, NewServer(svc))
}

func statusErr(err error) error {
	var domainErr *greet.Error
	if !errors.As(err, &domainErr) {
		return status.Error(codes.Internal, "internal error")
	}
	return status.Error(codeFor(domainErr.Code), domainErr.Message)
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
