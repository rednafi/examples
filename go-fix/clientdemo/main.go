package main

import (
	"fmt"

	"github.com/rednafi/examples/go-fix/client"
)

func main() {
	b, err := client.Fetch("https://example.com")
	fmt.Println(string(b), err)
	var c client.Config
	fmt.Println(c)
}
