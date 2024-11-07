package testutil

import (
	"reflect"
	"testing"
)

func AssertDeepEquals(t *testing.T, left, right any) {
	t.Helper()
	if !reflect.DeepEqual(left, right) {
		t.Errorf("%s is not equal to %s", left, right)
	}
}

func AssertLength[T any](t *testing.T, slice []T, expectedLength int) {
	t.Helper()
	if len(slice) != expectedLength {
		t.Errorf("Expected %d items (actual count: %d).", expectedLength, len(slice))
	}
}

func AssertContains[T any](t *testing.T, items []T, fn func(item T) bool) {
	t.Helper()
	for _, item := range items {
		if fn(item) {
			return
		}
	}
	t.Errorf("The collection does not contain the specified item. Collection: %v", items)
}

func AssertNotContains[T any](t *testing.T, items []T, fn func(item T) bool) {
	t.Helper()
	for _, item := range items {
		if fn(item) {
			t.Errorf("The collection contains the specified item. Collection: %v", items)
		}
	}

}
