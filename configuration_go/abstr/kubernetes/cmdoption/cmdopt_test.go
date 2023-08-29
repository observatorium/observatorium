package cmdopt_test

import (
	"testing"
	"time"

	cmdopt "github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/cmdoption"
)

type SubStruct struct {
	SubString string
}

func (s SubStruct) String() string {
	return s.SubString
}

type TestOptions struct {
	String       string        `opt:"string"`
	Int          int           `opt:"int"`
	Float        float64       `opt:"float"`
	Bool         bool          `opt:"bool"`
	Duration     time.Duration `opt:"duration"`
	Sub          SubStruct     `opt:"sub"`
	NoValue      bool          `opt:"no-value,noval"`
	Repeat       []string      `opt:"repeat"`
	SingleHyphen int           `opt:"single,single-hyphen"`
}

func TestCmdOptions(t *testing.T) {
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
		"float": {
			options: TestOptions{
				Float: 1.1,
			},
			expect: []string{"--float=1.10"},
		},
		"bool": {
			options: TestOptions{
				Bool: true,
			},
			expect: []string{"--bool=true"},
		},
		"duration": {
			options: TestOptions{
				Duration: time.Second,
			},
			expect: []string{"--duration=1s"},
		},
		"sub": {
			options: TestOptions{
				Sub: SubStruct{
					SubString: "sub",
				},
			},
			expect: []string{"--sub=sub"},
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
}
