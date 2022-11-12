package main

func ValueOrDefault[T comparable](value T, defaultValue T) T {
	if value == *new(T) {
		return defaultValue
	}
	return value
}
