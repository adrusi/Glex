package glex

import (
	"errors"
	"fmt"
	"reflect"
)

type injector map[reflect.Type]reflect.Value

// Expects the underlying value of all reflect.Values in vals to be *pointers
// to* the values of the type they provide. This allows any dependency to be
// provided by reference, which is vital for dependencies such as flags and
// counters.
func newInjector(vals ...reflect.Value) (j injector) {
	j = make(injector)
	for _, v := range vals {
		j[v.Type()] = v
	}
	return
}

func (j injector) call(f interface{}) ([]interface{}, error) {
	v := reflect.ValueOf(f)
	t := v.Type()
	if t.Kind() != reflect.Func {
		return nil,
			errors.New("injector.call requires its argument to be a func.")
	}
	args := make([]reflect.Value, t.NumIn())
	for i := 0; i < t.NumIn(); i++ {
		in := t.In(i)
		ref := false
		// detect whether the function expects the dependancy to be injected by
		// reference or by value, which is indicated by the type of the
		// dependency parameter being wraped in an anonymous pointer. Ex:
		//     func(counter *counterType) // ref = true
		//     func(counter  counterType) // ref = false
		if in.Name() == "" && in.Kind() == reflect.Ptr {
			ref = true
			in = in.Elem()
		} else {
			in = reflect.PtrTo(in)
		}
		arg, ok := j[t.In(i)]
		// This is really weird, the above should be functionally identical to
		// the bottom, but only the bottom actually works. My guess is that
		// maps operate with identity equality and == works with value
		// equality. TODO revisit this, make hash-based lookups work.
		for t, v := range j {
			if t == in {
				ok = true
				arg = v
				break
			}
		}
		if !ok {
			return nil, errors.New(fmt.Sprintf(
				"injector.call could not inject requested type %s.",
				in.Elem()))
		}
		if ref {
			args[i] = arg
		} else {
			args[i] = arg.Elem()
		}
	}
	results := v.Call(args)
	ret := make([]interface{}, len(results))
	for i, result := range results {
		ret[i] = result.Interface()
	}
	return ret, nil
}
