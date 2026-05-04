package collections

import (
	inner "github.com/microsoft/typescript-go/internal/collections"
	_ "unsafe"
)

var _ inner.OrderedMap[string, any]

type OrderedMap[K comparable, V any] = inner.OrderedMap[K, V]
