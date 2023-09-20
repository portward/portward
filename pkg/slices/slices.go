package slices

// Map applies f to each element in s and returns a new slice with the results of those function calls.
func Map[T, U any](s []T, f func(T) U) []U {
	if s == nil {
		return nil
	}

	r := make([]U, 0, len(s))

	for _, v := range s {
		r = append(r, f(v))
	}

	return r
}

// TryMap applies f to each element in s and returns a new slice with the results of those function calls.
// If any of the calls to f result in an error, TryMap immediately returns nil and the error.
func TryMap[T, U any](s []T, f func(T) (U, error)) ([]U, error) {
	if s == nil {
		return nil, nil
	}

	r := make([]U, 0, len(s))

	for _, v := range s {
		vv, err := f(v)
		if err != nil {
			return nil, err
		}

		r = append(r, vv)
	}

	return r, nil
}
