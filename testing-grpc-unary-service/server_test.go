package bookstore

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/rednafi/examples/testing-grpc-unary-service/api"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

var _ Store = (*memStore)(nil)

type memStore struct {
	mu    sync.Mutex
	books map[int64]Book
	next  int64
}

func (m *memStore) Create(_ context.Context, title, author string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.next++
	m.books[m.next] = Book{ID: m.next, Title: title, Author: author}
	return m.next, nil
}

func (m *memStore) Get(_ context.Context, id int64) (Book, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.books[id]
	if !ok {
		return Book{}, fmt.Errorf("book %d not found", id)
	}
	return b, nil
}

// Direct handler tests: call the server methods without gRPC transport.

func TestDirect_CreateAndGetBook(t *testing.T) {
	store := &memStore{books: make(map[int64]Book)}
	srv := &Server{store: store}

	created, err := srv.CreateBook(t.Context(), &api.CreateBookRequest{
		Title:  "DDIA",
		Author: "Martin Kleppmann",
	})
	if err != nil {
		t.Fatalf("CreateBook: %v", err)
	}
	if created.Id == 0 {
		t.Fatal("expected non-zero ID")
	}

	got, err := srv.GetBook(t.Context(), &api.GetBookRequest{Id: created.Id})
	if err != nil {
		t.Fatalf("GetBook: %v", err)
	}
	if got.Title != "DDIA" {
		t.Errorf("title = %q, want DDIA", got.Title)
	}
	if got.Author != "Martin Kleppmann" {
		t.Errorf("author = %q, want Martin Kleppmann", got.Author)
	}
}

func TestDirect_GetBook_NotFound(t *testing.T) {
	store := &memStore{books: make(map[int64]Book)}
	srv := &Server{store: store}

	_, err := srv.GetBook(t.Context(), &api.GetBookRequest{Id: 999})
	if err == nil {
		t.Fatal("expected error")
	}
	s, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if s.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", s.Code())
	}
}

func TestDirect_CreateBook_EmptyTitle(t *testing.T) {
	store := &memStore{books: make(map[int64]Book)}
	srv := &Server{store: store}

	_, err := srv.CreateBook(t.Context(), &api.CreateBookRequest{
		Title:  "",
		Author: "Someone",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	s, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", s.Code())
	}
}

// bufconn tests: exercise the full gRPC stack in-memory.

func startServer(t *testing.T, store Store, opts ...grpc.ServerOption) api.BookstoreClient {
	t.Helper()

	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer(opts...)
	RegisterServer(srv, store)

	go srv.Serve(lis)
	t.Cleanup(srv.GracefulStop)

	conn, err := grpc.NewClient("passthrough:///bufconn",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("connecting to bufconn: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	return api.NewBookstoreClient(conn)
}

func TestCreateAndGetBook(t *testing.T) {
	store := &memStore{books: make(map[int64]Book)}
	client := startServer(t, store)

	created, err := client.CreateBook(t.Context(), &api.CreateBookRequest{
		Title:  "DDIA",
		Author: "Martin Kleppmann",
	})
	if err != nil {
		t.Fatalf("CreateBook: %v", err)
	}
	if created.Id == 0 {
		t.Fatal("expected non-zero ID")
	}

	got, err := client.GetBook(t.Context(), &api.GetBookRequest{Id: created.Id})
	if err != nil {
		t.Fatalf("GetBook: %v", err)
	}
	if got.Title != "DDIA" {
		t.Errorf("title = %q, want DDIA", got.Title)
	}
	if got.Author != "Martin Kleppmann" {
		t.Errorf("author = %q, want Martin Kleppmann", got.Author)
	}
}

func TestGetBook_NotFound(t *testing.T) {
	store := &memStore{books: make(map[int64]Book)}
	client := startServer(t, store)

	_, err := client.GetBook(t.Context(), &api.GetBookRequest{Id: 999})
	if err == nil {
		t.Fatal("expected error")
	}
	s, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if s.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", s.Code())
	}
}

func TestCreateBook_EmptyTitle(t *testing.T) {
	store := &memStore{books: make(map[int64]Book)}
	client := startServer(t, store)

	_, err := client.CreateBook(t.Context(), &api.CreateBookRequest{
		Title:  "",
		Author: "Someone",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	s, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", s.Code())
	}
}

// Interceptor tests.

func TestRequestIDInterceptor(t *testing.T) {
	store := &memStore{books: make(map[int64]Book)}
	client := startServer(t, store,
		grpc.UnaryInterceptor(RequestIDInterceptor()),
	)

	var header metadata.MD
	_, err := client.CreateBook(t.Context(), &api.CreateBookRequest{
		Title:  "DDIA",
		Author: "Martin Kleppmann",
	}, grpc.Header(&header))
	if err != nil {
		t.Fatalf("CreateBook: %v", err)
	}

	ids := header.Get("x-request-id")
	if len(ids) == 0 {
		t.Fatal("expected x-request-id in response headers")
	}
	if ids[0] == "" {
		t.Fatal("x-request-id is empty")
	}
}

// Deadline tests.

type slowStore struct {
	*memStore
	delay time.Duration
}

func (s *slowStore) Get(ctx context.Context, id int64) (Book, error) {
	select {
	case <-time.After(s.delay):
		return s.memStore.Get(ctx, id)
	case <-ctx.Done():
		return Book{}, ctx.Err()
	}
}

func TestGetBook_DeadlineExceeded(t *testing.T) {
	base := &memStore{books: make(map[int64]Book)}
	store := &slowStore{memStore: base, delay: 2 * time.Second}
	client := startServer(t, store)

	created, err := client.CreateBook(t.Context(), &api.CreateBookRequest{
		Title:  "DDIA",
		Author: "Martin Kleppmann",
	})
	if err != nil {
		t.Fatalf("CreateBook: %v", err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()

	_, err = client.GetBook(ctx, &api.GetBookRequest{Id: created.Id})
	if err == nil {
		t.Fatal("expected error")
	}
	s, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if s.Code() != codes.DeadlineExceeded {
		t.Errorf("code = %v, want DeadlineExceeded", s.Code())
	}
}

// Metadata tests.

func echoRequestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if ids := md.Get("x-request-id"); len(ids) > 0 {
				grpc.SetHeader(ctx, metadata.Pairs("x-request-id", ids[0]))
			}
		}
		return handler(ctx, req)
	}
}

func TestMetadataPropagation(t *testing.T) {
	store := &memStore{books: make(map[int64]Book)}
	client := startServer(t, store,
		grpc.UnaryInterceptor(echoRequestIDInterceptor()),
	)

	ctx := metadata.AppendToOutgoingContext(t.Context(), "x-request-id", "abc-123")

	var header metadata.MD
	_, err := client.CreateBook(ctx, &api.CreateBookRequest{
		Title:  "DDIA",
		Author: "Martin Kleppmann",
	}, grpc.Header(&header))
	if err != nil {
		t.Fatalf("CreateBook: %v", err)
	}

	ids := header.Get("x-request-id")
	if len(ids) == 0 {
		t.Fatal("expected x-request-id in response headers")
	}
	if ids[0] != "abc-123" {
		t.Errorf("x-request-id = %q, want abc-123", ids[0])
	}
}

// Rich error detail tests.

func TestCreateBook_ValidationDetails(t *testing.T) {
	store := &memStore{books: make(map[int64]Book)}
	client := startServer(t, store)

	_, err := client.CreateBook(t.Context(), &api.CreateBookRequest{
		Title:  "",
		Author: "",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	s, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if s.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", s.Code())
	}

	details := s.Details()
	if len(details) == 0 {
		t.Fatal("expected error details")
	}
	br, ok := details[0].(*errdetails.BadRequest)
	if !ok {
		t.Fatalf("expected BadRequest, got %T", details[0])
	}
	if len(br.FieldViolations) != 2 {
		t.Fatalf("expected 2 field violations, got %d", len(br.FieldViolations))
	}

	fields := make(map[string]string)
	for _, v := range br.FieldViolations {
		fields[v.Field] = v.Description
	}
	if fields["title"] != "title is required" {
		t.Errorf("title violation = %q, want %q", fields["title"], "title is required")
	}
	if fields["author"] != "author is required" {
		t.Errorf("author violation = %q, want %q", fields["author"], "author is required")
	}
}
