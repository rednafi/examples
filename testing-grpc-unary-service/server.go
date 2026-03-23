package bookstore

import (
	"context"
	"fmt"
	"time"

	"github.com/rednafi/examples/testing-grpc-unary-service/api"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Book struct {
	ID     int64
	Title  string
	Author string
}

type Store interface {
	Create(ctx context.Context, title, author string) (int64, error)
	Get(ctx context.Context, id int64) (Book, error)
}

type Server struct {
	api.UnimplementedBookstoreServer
	store Store
}

func RegisterServer(srv *grpc.Server, store Store) {
	api.RegisterBookstoreServer(srv, &Server{store: store})
}

func (s *Server) CreateBook(
	ctx context.Context, req *api.CreateBookRequest,
) (*api.CreateBookResponse, error) {
	var violations []*errdetails.BadRequest_FieldViolation
	if req.Title == "" {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "title",
			Description: "title is required",
		})
	}
	if req.Author == "" {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "author",
			Description: "author is required",
		})
	}
	if len(violations) > 0 {
		st := status.New(codes.InvalidArgument, "invalid book request")
		st, err := st.WithDetails(&errdetails.BadRequest{
			FieldViolations: violations,
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "attaching details: %v", err)
		}
		return nil, st.Err()
	}

	id, err := s.store.Create(ctx, req.Title, req.Author)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "creating book: %v", err)
	}
	return &api.CreateBookResponse{Id: id}, nil
}

func (s *Server) GetBook(
	ctx context.Context, req *api.GetBookRequest,
) (*api.GetBookResponse, error) {
	book, err := s.store.Get(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "book %d not found", req.Id)
	}
	return &api.GetBookResponse{
		Id:     book.ID,
		Title:  book.Title,
		Author: book.Author,
	}, nil
}

func RequestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		id := fmt.Sprintf("%d", time.Now().UnixNano())
		grpc.SetHeader(ctx, metadata.Pairs("x-request-id", id))
		return handler(ctx, req)
	}
}
