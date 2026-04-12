package main

type Result[T any] struct {
	value T
	err   error
	ok    bool
}

func Ok[T any](value T) Result[T] {
	return Result[T]{value: value, ok: true}
}

func Fail[T any](err error) Result[T] {
	return Result[T]{err: err, ok: false}
}

func (r Result[T]) IsOk() bool    { return r.ok }
func (r Result[T]) Value() T      { return r.value }
func (r Result[T]) Error() error  { return r.err }

func (r Result[T]) Unwrap() (T, error) {
	if r.ok {
		return r.value, nil
	}
	return r.value, r.err
}

func Then[T any, U any](r Result[T], f func(T) Result[U]) Result[U] {
	if !r.ok {
		return Fail[U](r.err)
	}
	return f(r.value)
}
