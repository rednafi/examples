package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/rednafi/examples/wrapping-grpc-client/client"
)

func main() {
	c, err := client.New("localhost:9090")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	ctx := context.Background()

	err = c.Put(ctx, "greeting", []byte("hello"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("put greeting=hello")

	val, err := c.Get(ctx, "greeting")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("get greeting=%s\n", val)

	_, err = c.Get(ctx, "missing")
	if errors.Is(err, client.ErrNotFound) {
		fmt.Println("get missing: not found (expected)")
	}

	err = c.Delete(ctx, "greeting")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("deleted greeting")

	_, err = c.Get(ctx, "greeting")
	if errors.Is(err, client.ErrNotFound) {
		fmt.Println("get greeting after delete: not found (expected)")
	}
}
