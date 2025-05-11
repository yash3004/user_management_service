package conversion

func BoolPtr(b bool) *bool {
	return &b
}

func StrPtr(s string) *string {
	return &s
}

func IntPtr(i int) *int {
	return &i
}

func Int32Ptr(i int32) *int32 {
	return &i
}

func Int64Ptr(i int64) *int64 {
	return &i
}

func Deref[T any](val *T) T {
	if val == nil {
		var t T
		return t
	}

	return *val
}

func Ptr[T any](val T) *T {
	return &val
}
