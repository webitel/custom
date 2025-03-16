package postgres

import "reflect"

func coalesce[T any](vs ...T) T {
	var rv reflect.Value
	for _, v := range vs {
		rv = reflect.ValueOf(v)
		if !rv.IsValid() || rv.IsZero() {
			continue // nil | empty
		}
		// switch rv.Kind() {
		// case reflect.Interface, reflect.Slice,
		// 	reflect.Chan, reflect.Func, reflect.Map,
		// 	reflect.Pointer, reflect.UnsafePointer:
		// 	{
		// 		if rv.IsNil() {
		// 			continue // nil
		// 		}
		// 	}
		// default:
		// 	{
		// 		if rv.IsZero() {
		// 			continue // zero
		// 		}
		// 	}
		// }
		return v
	}
	rt := reflect.TypeFor[T]()
	rv = reflect.New(rt)
	rv = reflect.Indirect(rv)
	return rv.Interface().(T)
}
