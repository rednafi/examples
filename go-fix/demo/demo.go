// Command demo calls the deprecated greet.Hello. Because greet.Hello carries a
// //go:fix inline directive, plain go fix migrates the call to greet.Greet. No
// custom tool required:
//
//	go fix -diff ./demo
package main

import (
	"fmt"

	"github.com/rednafi/examples/go-fix/greet"
)

func main() {
	fmt.Println(greet.Hello("Go"))
}
