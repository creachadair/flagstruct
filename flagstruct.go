// Package flagstruct supports automatic registration of struct fields as
// flags.  Flags are associated with exported struct fields that have a tag
// declaring them as a flag with the specified name and help text:
//
//   flag:"flagname,help description"
//
// A flag may optionally be given a default value, using the tag:
//
//   flag-default:"default flag value"
//
// If a default value is not provided as a tag, the existing value of the
// target is used as the default.
package flagstruct

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// flagInfo captures the information needed to register a struct field in a
// flag.FlagSet.
type flagInfo struct {
	field interface{} // must be of pointer type
	name  string
	help  string
	dval  *string // default value if not nil, encoded as input to Set
}

func (fi *flagInfo) setDefault() error {
	if fi.dval == nil {
		return nil
	}
	switch t := fi.field.(type) {
	case flag.Value:
		return t.Set(*fi.dval)
	case *bool:
		b, err := strconv.ParseBool(*fi.dval)
		if err != nil {
			return err
		}
		*t = b
	case *time.Duration:
		d, err := time.ParseDuration(*fi.dval)
		if err != nil {
			return err
		}
		*t = d
	case *float64:
		f, err := strconv.ParseFloat(*fi.dval, 64)
		if err != nil {
			return err
		}
		*t = f
	case *int, *int64:
		z, err := strconv.ParseInt(*fi.dval, 0, 64)
		if err != nil {
			return err
		}
		switch u := t.(type) {
		case *int:
			*u = int(z)
		case *int64:
			*u = z
		}
	case *string:
		*t = *fi.dval
	case *uint, *uint64:
		z, err := strconv.ParseUint(*fi.dval, 0, 64)
		if err != nil {
			return err
		}
		switch u := t.(type) {
		case *uint:
			*u = uint(z)
		case *uint64:
			*u = z
		}
	default:
		panic("invalid target for default")
	}
	return nil
}

// register registers fi with fs if fi.field implements flag.Value or is one of
// the supported built-in types.
func (fi *flagInfo) register(fs *flag.FlagSet, prefix string) error {
	p := func(s string) string { return prefix + s }
	if err := fi.setDefault(); err != nil {
		return err
	}
	switch t := fi.field.(type) {
	case flag.Value:
		fs.Var(t, p(fi.name), fi.help)
	case *bool:
		fs.BoolVar(t, p(fi.name), *t, fi.help)
	case *time.Duration:
		fs.DurationVar(t, p(fi.name), *t, fi.help)
	case *float64:
		fs.Float64Var(t, p(fi.name), *t, fi.help)
	case *int64:
		fs.Int64Var(t, p(fi.name), *t, fi.help)
	case *int:
		fs.IntVar(t, p(fi.name), *t, fi.help)
	case *string:
		fs.StringVar(t, p(fi.name), *t, fi.help)
	case *uint64:
		fs.Uint64Var(t, p(fi.name), *t, fi.help)
	case *uint:
		fs.UintVar(t, p(fi.name), *t, fi.help)
	default:
		return fmt.Errorf("type %T does not implement flag.Value", fi.field)
	}
	return nil
}

func (fi *flagInfo) String() string { return fmt.Sprintf("#<flag %q help=%q>", fi.name, fi.help) }

// newFlagInfo extracts the flag name and help string from the tag of sf and
// constructs a *flagInfo if possible.  If not, newFlagInfo returns nil, false.
func newFlagInfo(sf reflect.StructField, v reflect.Value) (*flagInfo, bool) {
	tag := sf.Tag.Get("flag")
	if tag == "" || sf.PkgPath != "" {
		return nil, false // no tag, or field is unexported
	}
	fi := &flagInfo{
		field: v.Addr().Interface(),
		name:  tag,
		help:  tag,
	}
	if ps := strings.SplitN(tag, ",", 2); len(ps) == 2 {
		fi.name = ps[0]
		fi.help = ps[1]
	}
	if dval := sf.Tag.Get("flag-default"); dval != "" {
		fi.dval = &dval
		log.Printf("MJF :: flag-default for %q is %q", tag, dval)
	}
	return fi, true
}

// parseFlags returns a flagInfo record for each field of v that supports
// registration with the flag package.
func parseFlags(v interface{}) ([]*flagInfo, error) {
	s := reflect.ValueOf(v)
	if s.Kind() != reflect.Ptr {
		return nil, errors.New("value must be a pointer")
	}
	s = reflect.Indirect(s)
	if s.Kind() != reflect.Struct {
		return nil, errors.New("value must be a struct")
	}

	t := s.Type()
	var flags []*flagInfo
	for i := 0; i < s.NumField(); i++ {
		fi, ok := newFlagInfo(t.Field(i), s.Field(i))
		if ok {
			flags = append(flags, fi)
		}
	}
	return flags, nil
}

// Register adds a flag to fs for each field of v that is flaggable.  It is an
// error if v is not a pointer to a struct value.
//
// An exported field is flaggable if it has a field tag of the form
// `flag:"name,usage"` and a pointer to its type implements the flag.Value
// interface.  As a special case, the built-in types supported by the flag
// package are also allowed (bool, int, time.Duration, float64, etc.).
//
// Unexported fields and fields without flag tags are skipped without error;
// however it is an error if there are no flaggable fields in the type.
func Register(v interface{}, fs *flag.FlagSet) error { return RegisterTag("", v, fs) }

// RegisterTag behaves as Register, with the name of each flag prefixed by the
// given tag.
func RegisterTag(tag string, v interface{}, fs *flag.FlagSet) error {
	flags, err := parseFlags(v)
	if err != nil {
		return err
	} else if len(flags) == 0 {
		return errors.New("struct contains no flaggable fields")
	}
	for _, fi := range flags {
		if err := fi.register(fs, tag); err != nil {
			return err
		}
	}
	return nil
}
