package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/rednafi/examples/wrapping-grpc-client/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var ErrNotFound = errors.New("key not found")

type KV interface {
	Put(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, key string) (value []byte, err error)
	Delete(ctx context.Context, key string) error
}

type Client struct {
	conn *grpc.ClientConn
	kv   api.KVClient
}

func New(addr string, opts ...grpc.DialOption) (*Client, error) {
	if len(opts) == 0 {
		opts = []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
	}
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("connecting to %s: %v", addr, err)
	}
	return &Client{
		conn: conn,
		kv:   api.NewKVClient(conn),
	}, nil
}

func (c *Client) Put(ctx context.Context, key string, value []byte) error {
	_, err := c.kv.Put(ctx, &api.PutRequest{
		Key:   key,
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("putting key %s: %v", key, err)
	}
	return nil
}

func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	resp, err := c.kv.Get(ctx, &api.GetRequest{Key: key})
	if err != nil {
		if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting key %s: %v", key, err)
	}
	return resp.Value, nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.kv.Delete(ctx, &api.DeleteRequest{Key: key})
	if err != nil {
		return fmt.Errorf("deleting key %s: %v", key, err)
	}
	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
