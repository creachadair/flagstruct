// Program flagstruct_test demonstrates the use of the flagstruct package.
package flagstruct_test

import (
	"flag"
	"fmt"
	"log"

	"github.com/creachadair/flagstruct"
)

func Example() {
	// You can define a named type for configuration, but it also works fine with
	// anonymous structs.
	var config = struct {
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

	// Register the configuration with the flag package before parsing.
	// For this example, we've created a separate flag set.
	fs := flag.NewFlagSet("example", flag.PanicOnError)
	if err := flagstruct.Register(&config, fs); err != nil {
		log.Fatalf("Error registering flags: %v", err)
	}
	fs.Parse([]string{"-in", "apple", "-count", "37", "-out", "orange", "a", "b", "c"})

	fmt.Printf("%+v\n", config)
	fmt.Println(fs.Args())
	// Output:
	// {Input:apple Output:orange Count:37 Other:blub inner:120}
	// [a b c]
}
