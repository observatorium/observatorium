package cmdopt

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	tagName              = "opt"
	noValModifier        = "noval"
	singleHyphenModifier = "single-hyphen"
)

type CmdOptions struct{}

func GetOpts(obj interface{}) []string {
	ret := []string{}

	// if obj is nil, return empty slice
	if obj == nil {
		return ret
	}

	// if obj is a pointer, dereference it
	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		obj = reflect.ValueOf(obj).Elem().Interface()
	}

	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	for i := 0; i < t.NumField(); i++ {
		optTagVals := strings.Split(t.Field(i).Tag.Get(tagName), ",")
		if len(optTagVals) == 0 {
			continue
		}

		optName := getOptName(optTagVals)
		if optName == "" {
			continue
		}

		// Check if noval modifier is set.
		kind := t.Field(i).Type.Kind()
		value := v.Field(i)
		if isNoVal(optTagVals[0:], kind, value) {
			ret = append(ret, optName)
			continue
		}

		optValue := getOptValue(kind, value)
		if len(optValue) == 0 || optValue[0] == "" {
			continue
		}

		for _, v := range optValue {
			ret = append(ret, fmt.Sprintf("%s=%s", optName, v))
		}
	}

	return ret
}

func getOptName(opt []string) string {

	switch len(opt[0]) {
	case 0:
		return ""
	case 1:
		return "-" + opt[0]
	default:
		if isSingleHyphen(opt[1:]) {
			return "-" + opt[0]
		}

		return "--" + opt[0]
	}
}

func getOptValue(kind reflect.Kind, field reflect.Value) []string {
	ret := []string{}

	switch kind {
	case reflect.String:
		value := field.String()
		if value == "" {
			return ret
		}

		ret = append(ret, value)
	case reflect.Int:
		value := field.Int()
		if value == 0 {
			return ret
		}

		ret = append(ret, fmt.Sprintf("%d", value))
	case reflect.Bool:
		value := field.Bool()
		if !value {
			return ret
		}

		ret = append(ret, fmt.Sprintf("%t", value))
	case reflect.Float64:
		value := field.Float()
		if value == 0 {
			return ret
		}

		ret = append(ret, fmt.Sprintf("%.2f", value))
	case reflect.Struct, reflect.Int64: // Int64 for time.Duration
		if kind == reflect.Int64 {
			if field.Int() == 0 {
				return ret
			}
		}
		str, ok := field.Interface().(fmt.Stringer)
		if !ok {
			return ret
		}

		value := str.String()
		if value == "" {
			return ret
		}

		ret = append(ret, value)
	case reflect.Slice:
		if field.Len() == 0 {
			return ret
		}

		// get slice values recursively
		for i := 0; i < field.Len(); i++ {
			ret = append(ret, getOptValue(field.Index(i).Kind(), field.Index(i))...)
		}
	default:
		panic(fmt.Errorf("unsupported type: %s", kind))
	}

	return ret
}

func isNoVal(optVals []string, kind reflect.Kind, v reflect.Value) bool {
	for _, optVal := range optVals {
		if optVal == noValModifier && kind == reflect.Bool && v.Bool() {
			return true
		}
	}

	return false
}

func isSingleHyphen(optVals []string) bool {
	for _, optVal := range optVals {
		if optVal == singleHyphenModifier {
			return true
		}
	}

	return false
}

// ExtraOpts is a struct that can be embedded in a struct to add extra options.
// These options can be used without exposing them in the struct.
type ExtraOpts struct {
	opts []string
}

// AddOpts adds extra options to the struct.
func (e *ExtraOpts) AddOpts(s ...string) {
	e.opts = append(e.opts, s...)
}

// GetExtraOpts returns the extra options.
func (e *ExtraOpts) GetExtraOpts() []string {
	return e.opts
}
