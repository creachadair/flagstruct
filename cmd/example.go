// Program example demonstrates the use of the flagstruct package.
//
// Usage example:
//   go run example.go -in foo -out bar -count 33
package main

import (
	"flag"
	"fmt"
	"log"

	"bitbucket.org/creachadair/flagstruct"
)

// You can define a named type for configuration, but it also works fine with
// anonymous structs.
var config = &struct {
	Input  string `flag:"in,The path of the input file"`
	Output string `flag:"out,The path of the output file"`
	Count  int    `flag:"count,The number of lines to process"`
	Other  string // field will not be flagged; there is no struct tag
	inner  byte   // field will not be flagged; it is unexported
}{
	Input: "default",
	Count: 17,
	Other: "blub",
	inner: 'x',
}

func init() {
	// Register the configuration with the flag package before calling flag.Parse.
	if err := flagstruct.Register(config, flag.CommandLine); err != nil {
		log.Fatalf("Error registering flags: %v", err)
	}
}

func main() {
	flag.Parse()

	fmt.Printf("-- Configuration after flag parsing:\n%+v\n", config)
}
