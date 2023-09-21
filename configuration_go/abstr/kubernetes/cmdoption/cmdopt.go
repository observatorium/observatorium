package cmdopt

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

const (
	tagName              = "opt"
	noValModifier        = "noval"
	singleHyphenModifier = "single-hyphen"
)

type CmdOptions struct{}

// GetOpts is used to generate command line options from a struct
// by mapping the fields values to the option name defined in the opt tag (first position).
// GetOpts returns a slice of strings, each string representing an option.
// Following types are supported: string, int, bool, float64, time.Duration, slice of supported types, sub struct implementing the Stringer interface.
// Pointer types are used if not nil. Private fields are ignored.
// Additional tags can be added to the opt tag, separated by a comma, to modify its default behavior:
// - noval: the option is added without a value if the field is true.
// - single-hyphen: the option is prefixed with a single hyphen instead of a double hyphen.
func GetOpts(obj interface{}) []string {
	ret := []string{}

	// if obj is nil, return empty slice
	if obj == nil {
		return ret
	}

	// Extract extra options using ExtraOpts interface if it is implemented.
	// Append them to the result slice at the end.
	var extraOpts []string
	getExtraOptsMethod := reflect.ValueOf(obj).MethodByName("GetExtraOpts")
	if getExtraOptsMethod.IsValid() {
		result := getExtraOptsMethod.Call(nil)
		if len(result) > 0 {
			extraOpts, _ = result[0].Interface().([]string)
		}
	}

	// if obj is a pointer, dereference it
	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		obj = reflect.ValueOf(obj).Elem().Interface()
	}

	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	for i := 0; i < t.NumField(); i++ {
		fieldKind := t.Field(i).Type.Kind()
		fieldValue := v.Field(i)

		// If the field is not exported, skip it.
		if t.Field(i).PkgPath != "" {
			continue
		}

		optTagVals := strings.Split(t.Field(i).Tag.Get(tagName), ",")
		if len(optTagVals) == 0 {
			continue
		}

		optName := getOptName(optTagVals)
		if optName == "" {
			continue
		}

		// Check if noval modifier is set.
		if isNoVal(optTagVals[0:], fieldKind, fieldValue) {
			ret = append(ret, optName)
			continue
		}

		optValue := getOptValue(fieldKind, fieldValue)
		if len(optValue) == 0 || optValue[0] == "" {
			continue
		}

		for _, v := range optValue {
			ret = append(ret, fmt.Sprintf("%s=%s", optName, v))
		}
	}

	ret = append(ret, extraOpts...)

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

func getOptValue(kind reflect.Kind, rValue reflect.Value) []string {
	ret := []string{}

	// If pointer type and nil, skip it, otherwise dereference it.
	if kind == reflect.Ptr {
		if rValue.IsNil() {
			return ret
		}

		// Check if Stringer interface is implemented on pointer receiver.
		if str := getStringerValue(kind, rValue); str != "" {
			ret = append(ret, str)
			return ret
		}

		rValue = rValue.Elem()
		kind = rValue.Kind()
	}

	switch kind {
	case reflect.String:
		value := rValue.String()
		if value == "" {
			return ret
		}

		ret = append(ret, value)
	case reflect.Int:
		value := rValue.Int()
		if value == 0 {
			return ret
		}

		ret = append(ret, fmt.Sprintf("%d", value))
	case reflect.Bool:
		value := rValue.Bool()
		if !value {
			return ret
		}

		ret = append(ret, fmt.Sprintf("%t", value))
	case reflect.Float64:
		value := rValue.Float()
		if value == 0 {
			return ret
		}

		ret = append(ret, fmt.Sprintf("%.2f", value))
	case reflect.Struct, reflect.Int64: // Int64 for time.Duration
		if kind == reflect.Int64 {
			if rValue.Int() == 0 {
				return ret
			}
		}

		// Check if Stringer interface is implemented on struct receiver.
		str := getStringerValue(kind, rValue)
		if str == "" {
			return ret
		}

		ret = append(ret, str)

	case reflect.Slice:
		if rValue.Len() == 0 {
			return ret
		}

		// get slice values recursively
		for i := 0; i < rValue.Len(); i++ {
			ret = append(ret, getOptValue(rValue.Index(i).Kind(), rValue.Index(i))...)
		}
	default:
		log.Printf("unsupported type %q by cmdopt is ignored", kind)
	}

	return ret
}

func getStringerValue(kind reflect.Kind, rValue reflect.Value) string {
	str, ok := rValue.Interface().(fmt.Stringer)
	if !ok {
		return ""
	}

	return str.String()
}

func isNoVal(optVals []string, kind reflect.Kind, v reflect.Value) bool {
	// If pointer type and nil, skip it, otherwise dereference it.
	if kind == reflect.Ptr {
		if v.IsNil() {
			return false
		}

		v = v.Elem()
		kind = v.Kind()
	}

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
func (e *ExtraOpts) AddExtraOpts(s ...string) {
	e.opts = append(e.opts, s...)
}

// GetExtraOpts returns the extra options.
func (e *ExtraOpts) GetExtraOpts() []string {
	return e.opts
}

// DeleteExtraOpts deletes the extra options.
func (e *ExtraOpts) DeleteExtraOpts() {
	e.opts = []string{}
}
