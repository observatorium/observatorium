package cmdopt_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/observatorium/observatorium/configuration_go/kubegen/cmdopt"
)

type SubStruct struct {
	SubString string
}

func (s SubStruct) String() string {
	return s.SubString
}

type SubStructPtr struct {
	SubString string
}

func (s *SubStructPtr) String() string {
	return s.SubString
}

type Dummy struct{}

func (d Dummy) GoString() string {
	return "dummy"
}

type TestOptions struct {
	String       string         `opt:"string"`
	Int          int            `opt:"int"`    // Zero value is ignored
	IntPtr       *int           `opt:"intptr"` // Zero value is not ignored
	Float        float64        `opt:"float"`
	FloatPtr     *float64       `opt:"floatptr"`
	Bool         bool           `opt:"bool"`
	BoolPtr      *bool          `opt:"bool"`
	Duration     time.Duration  `opt:"duration"`
	DurationPtr  *time.Duration `opt:"duration"`
	Sub          SubStruct      `opt:"sub"`
	SubPtr       *SubStructPtr  `opt:"subptr"`
	NoValue      bool           `opt:"no-value,noval"`
	Repeat       []string       `opt:"repeat"`
	SingleHyphen int            `opt:"single,single-hyphen"`
	Interface    fmt.Stringer   `opt:"stringer"`

	// Limits tests
	DoubleType             string         `opt:"string,int"`
	NovalWithoutBoolType   float64        `opt:"nobool,noval"`
	PointerToSupportedType *string        `opt:"string"`
	DoublePointer          **string       `opt:"string"`
	RepeatToPointer        []*string      `opt:"repeat"`
	EmptyTag               string         `opt:""`
	OtherTagName           string         `opts:"other"`
	InvalidInterface       fmt.GoStringer `opt:"stringer"`

	privateField string `opt:"string"`
}

func TestCmdOptions(t *testing.T) {
	myStringPointer := &[]string{"string"}[0]

	testCases := map[string]struct {
		options TestOptions
		expect  []string
	}{
		"empty": {
			options: TestOptions{},
			expect:  []string{},
		},
		"string": {
			options: TestOptions{
				String: "string",
			},
			expect: []string{"--string=string"},
		},
		"int": {
			options: TestOptions{
				Int: 1,
			},
			expect: []string{"--int=1"},
		},
		"int-zero": {
			options: TestOptions{
				Int: 0,
			},
			expect: []string{},
		},
		"int-ptr-zero": {
			options: TestOptions{
				IntPtr: new(int),
			},
			expect: []string{"--intptr=0"},
		},
		"float": {
			options: TestOptions{
				Float: 1.1,
			},
			expect: []string{"--float=1.1"},
		},
		"float-zero": {
			options: TestOptions{
				Float: 0,
			},
			expect: []string{},
		},
		"float-ptr-zero": {
			options: TestOptions{
				FloatPtr: new(float64),
			},
			expect: []string{"--floatptr=0"},
		},
		"bool": {
			options: TestOptions{
				Bool: true,
			},
			expect: []string{"--bool=true"},
		},
		"bool-false": {
			options: TestOptions{
				Bool: false,
			},
			expect: []string{},
		},
		"bool-ptr-false": {
			options: TestOptions{
				BoolPtr: new(bool),
			},
			expect: []string{"--bool=false"},
		},
		"duration": {
			options: TestOptions{
				Duration: time.Second,
			},
			expect: []string{"--duration=1s"},
		},
		"duration-zero": {
			options: TestOptions{
				Duration: 0,
			},
			expect: []string{},
		},
		"duration-ptr-zero": {
			options: TestOptions{
				DurationPtr: new(time.Duration),
			},
			expect: []string{"--duration=0s"},
		},
		"sub": {
			options: TestOptions{
				Sub: SubStruct{
					SubString: "sub",
				},
			},
			expect: []string{"--sub=sub"},
		},
		"sub-zero": {
			options: TestOptions{
				Sub: SubStruct{},
			},
			expect: []string{},
		},
		"sub-ptr-zero": {
			options: TestOptions{
				SubPtr: &SubStructPtr{},
			},
			expect: []string{},
		},
		"sub with stringer interface on pointer receiver": {
			options: TestOptions{
				SubPtr: &SubStructPtr{
					SubString: "sub",
				},
			},
			expect: []string{"--subptr=sub"},
		},
		"no-value": {
			options: TestOptions{
				NoValue: true,
			},
			expect: []string{"--no-value"},
		},
		"repeat": {
			options: TestOptions{
				Repeat: []string{"repeat1", "repeat2"},
			},
			expect: []string{"--repeat=repeat1", "--repeat=repeat2"},
		},
		"single-hyphen": {
			options: TestOptions{
				SingleHyphen: 1,
			},
			expect: []string{"-single=1"},
		},
		"many": {
			options: TestOptions{
				String: "string",
				Int:    1,
			},
			expect: []string{"--string=string", "--int=1"},
		},
		"double type ignored": {
			options: TestOptions{
				DoubleType: "string",
			},
			expect: []string{"--string=string"},
		},
		"noval without bool type ignored": {
			options: TestOptions{
				NovalWithoutBoolType: 1.1,
			},
			expect: []string{"--nobool=1.1"},
		},
		"pointer to supported type": {
			options: TestOptions{
				PointerToSupportedType: &[]string{"string"}[0],
			},
			expect: []string{"--string=string"},
		},
		"nil pointer to supported type": {
			options: TestOptions{
				PointerToSupportedType: nil,
			},
			expect: []string{},
		},
		"private field ignored": {
			options: TestOptions{
				privateField: "string",
			},
			expect: []string{},
		},
		"double pointer ignored": {
			options: TestOptions{
				DoublePointer: &myStringPointer,
			},
			expect: []string{},
		},
		"repeat to pointer": {
			options: TestOptions{
				RepeatToPointer: []*string{myStringPointer},
			},
			expect: []string{"--repeat=string"},
		},
		"empty tag ignored": {
			options: TestOptions{
				EmptyTag: "string",
			},
			expect: []string{},
		},
		"other tag ignored": {
			options: TestOptions{
				OtherTagName: "string",
			},
			expect: []string{},
		},
		"interface with stringer": {
			options: TestOptions{
				Interface: SubStruct{
					SubString: "sub",
				},
			},
			expect: []string{"--stringer=sub"},
		},
		"interface without stringer": {
			options: TestOptions{
				InvalidInterface: Dummy{},
			},
			expect: []string{},
		},
	}

	t.Parallel()

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			args := cmdopt.GetOpts(tc.options)
			if len(args) != len(tc.expect) {
				t.Fatalf("expected %d args, got %d: %s", len(tc.expect), len(args), args)
			}

			for i := range args {
				if args[i] != tc.expect[i] {
					t.Fatalf("expected %s, got %s", tc.expect[i], args[i])
				}
			}
		})
	}
}

// TestUnsupportedType tests that usupported types are ignored.
// Done in a separate test because it generates a log message.
func TestUnsupportedType(t *testing.T) {
	unsupportedType := struct {
		Chan chan int `opt:"chan"`
	}{
		Chan: make(chan int),
	}

	args := cmdopt.GetOpts(&unsupportedType)
	if len(args) != 0 {
		t.Fatalf("expected 0 args, got %d: %s", len(args), args)
	}
}

func TestCmdOptionsExtraOpts(t *testing.T) {
	objWithExtraOpts := struct {
		String string `opt:"string"`
		cmdopt.ExtraOpts
	}{
		String: "string",
	}

	objWithExtraOpts.AddExtraOpts("--extra1", "--extra2")

	expected := []string{"--string=string", "--extra1", "--extra2"}

	args := cmdopt.GetOpts(&objWithExtraOpts)
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d: %s", len(args), args)
	}

	for i := range args {
		if args[i] != expected[i] {
			t.Fatalf("expected %s, got %s", expected[i], args[i])
		}
	}

	objWithExtraOpts.DeleteExtraOpts()

	args = cmdopt.GetOpts(&objWithExtraOpts)
	if len(args) != 1 {
		t.Fatalf("expected 1 args, got %d: %s", len(args), args)
	}
}
