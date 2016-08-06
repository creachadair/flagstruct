package flagstruct

import (
	"flag"
	"fmt"
	"log"
	"testing"
	"time"
)

func TestTypeErrors(t *testing.T) {
	// Each of these types should generate an error from Register.
	tests := []interface{}{
		"foo",             // not a pointer, not a struct
		37,                // ditto
		new(int),          // not a struct
		struct{ S int }{}, // not a pointer
		&struct{}{},       // no flaggable fields
		&struct { // no flaggable fields
			z int     // unexported, no struct tag
			r rune    `flag:"r,rune"` // unexported
			F float64 // no struct tag
			U uint    `json:"q"` // missing flag: clause in struct tag
			S string  `flag:""`  // bogus flag: clause in struct tag
		}{},
	}
	fs := flag.NewFlagSet("dummy", flag.PanicOnError)
	for _, bad := range tests {
		if err := Register(bad, fs); err != nil {
			t.Logf("Register(%T) gave expected error: %v", bad, err)
		} else {
			t.Errorf("Register(%T) expected error, but got %+v", bad, fs)
		}
	}
}

// dummy implements a no-op flag.Value for testing purposes.
type dummy struct{}

func (d *dummy) String() string     { return "" }
func (d *dummy) Set(s string) error { return nil }

func TestRegistration(t *testing.T) {
	tests := []struct {
		name, help string
		input      interface{}
	}{
		{"dummy", "fake", &struct {
			D dummy `flag:"dummy,fake"` // *D implements flag.Value directly
		}{}},
		{"b", "bool", &struct {
			B bool `flag:"b,bool"`
		}{}},
		{"d", "duration", &struct {
			D time.Duration `flag:"d,duration"`
		}{}},
		{"f64", "float64", &struct {
			F float64 `flag:"f64,float64"`
		}{}},
		{"z64", "int64", &struct {
			Z int64 `flag:"z64,int64"`
		}{}},
		{"z", "int", &struct {
			Z int `flag:"z,int"`
		}{}},
		{"s", "string", &struct {
			S string `flag:"s,string"`
		}{}},
		{"u64", "uint64", &struct {
			UZ uint64 `flag:"u64,uint64"`
		}{}},
		{"u", "uint", &struct {
			UZ uint `flag:"u,uint"`
		}{}},
		{"wat", "wat", &struct {
			S string `flag:"wat"` // missing help is OK
		}{}},
	}
	for _, tag := range []string{"", "foo_"} {
		for _, test := range tests {
			fs := flag.NewFlagSet("dummy", flag.PanicOnError)
			if err := RegisterTag(tag, test.input, fs); err != nil {
				t.Errorf("Registering %T (%v) failed: %v", test.input, test.input, err)
				continue
			}
			name := tag + test.name
			got := fs.Lookup(name)
			if got == nil {
				t.Errorf("Lookup %q failed: flag not found", name)
				continue
			}
			t.Logf("Found flag for %q: %+v", name, got)
			if got.Usage != test.help {
				t.Errorf("Flag %+v help: got %q, want %q", test.input, got.Usage, test.help)
			}
		}
	}
}

func ExampleUsage() {
	type Config struct {
		Input  string `flag:"in,The path of the input file"`
		Output string `flag:"out,The path of the output file"`
		Count  int    `flag:"count,The number of lines to process"`
		Other  string // field will not be flagged; there is no struct tag
		inner  byte   // field will not be flagged; it is unexported
	}

	// Register a pointer to the config type with the flag set.  Default values
	// are taken from the struct at registration time and non-flag fields are
	// not touched.
	c := &Config{Count: 17, Other: "p", inner: 'x'}
	if err := Register(c, flag.CommandLine); err != nil {
		log.Fatalf("Registration failed for %+v: %v", c, err)
	}

	// In a real program, you'd just call flag.Parse(); here we simulate that
	// by invoking the parser directly on some arguments.
	args := []string{"-in", "in.bin", "-out", "out.bin", "foo"}
	if err := flag.CommandLine.Parse(args); err != nil {
		log.Fatalf("flag.Parse() failed: %v", err)
	}

	fmt.Printf("in=%s out=%s count=%d other=%s inner=%c\n", c.Input, c.Output, c.Count, c.Other, c.inner)
	// Output: in=in.bin out=out.bin count=17 other=p inner=x
}
