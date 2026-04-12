package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rednafi/examples/error-translation/grpc/api"
	"github.com/rednafi/examples/error-translation/grpc/sqlite"
	"github.com/rednafi/examples/error-translation/grpc/user"
)

type handler struct {
	api.UnimplementedUserServiceServer
	svc *user.Service
}

func (h *handler) GetUser(
	ctx context.Context, req *api.GetUserRequest,
) (*api.GetUserResponse, error) {
	u, err := h.svc.GetUser(ctx, req.GetId())
	if err != nil {
		slog.ErrorContext(ctx, "get user failed",
			"user_id", req.GetId(),
			"err", err,
		)
		return nil, toStatus(err)
	}
	return &api.GetUserResponse{
		Id: u.ID, Name: u.Name, Email: u.Email,
	}, nil
}

func (h *handler) CreateUser(
	ctx context.Context, req *api.CreateUserRequest,
) (*api.CreateUserResponse, error) {
	u, err := h.svc.CreateUser(ctx, req.GetName(), req.GetEmail())
	if err != nil {
		slog.ErrorContext(ctx, "create user failed",
			"name", req.GetName(),
			"email", req.GetEmail(),
			"err", err,
		)
		return nil, toStatus(err)
	}
	return &api.CreateUserResponse{Id: u.ID}, nil
}

func toStatus(err error) error {
	switch {
	case errors.Is(err, user.ErrNotFound):
		return status.Error(codes.NotFound, "not found")
	case errors.Is(err, user.ErrConflict):
		return status.Error(codes.AlreadyExists, "conflict")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func main() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE
	)`)
	if err != nil {
		log.Fatal(err)
	}

	store := sqlite.NewUserStore(db)
	svc := user.NewService(store)

	srv := grpc.NewServer()
	api.RegisterUserServiceServer(srv, &handler{svc: svc})

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("listening on :9090")
	log.Fatal(srv.Serve(lis))
}
