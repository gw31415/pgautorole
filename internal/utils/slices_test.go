package utils_test

import (
	"testing"

	"github.com/gw31415/pgautorole/internal/utils"
)

func TestSlicesMap(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	f := func(v int) int { return v * 2 }
	result := utils.SlicesMap(s, f)
	expected := []int{2, 4, 6, 8, 10}
	if len(result) != len(expected) {
		t.Fatalf("unexpected length: %v", result)
	}
	for i, v := range result {
		if v != expected[i] {
			t.Fatalf("unexpected element: %v", v)
		}
	}
}

func TestSlicesCount(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	test := func(v int) bool { return v%2 == 0 }
	result := utils.SlicesCount(s, test)
	expected := 2
	if result != expected {
		t.Fatalf("unexpected count: %v", result)
	}
}

func TestSlicesFilter(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	test := func(v int) bool { return v%2 == 0 }
	result := utils.SlicesFilter(s, test)
	expected := []int{2, 4}
	if len(result) != len(expected) {
		t.Fatalf("unexpected length: %v", result)
	}
	for i, v := range result {
		if v != expected[i] {
			t.Fatalf("unexpected element: %v", v)
		}
	}
}

func TestSlicesFilterMap(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	f := func(v int) *int {
		if v%2 == 0 {
			v /= 2
			return &v
		}
		return nil
	}
	result := utils.SlicesFilterMap(s, f)
	expected := []int{1, 2}
	if len(result) != len(expected) {
		t.Fatalf("unexpected length: %v", result)
	}
	for i, v := range result {
		if v != expected[i] {
			t.Fatalf("unexpected element: %v", v)
		}
	}
}

func TestSlicesIntersect(t *testing.T) {
	s1 := []int{1, 2, 3, 4, 5}
	s2 := []int{4, 5, 6, 7, 8}
	result := utils.SlicesIntersect(s1, s2)
	expected := []int{4, 5}
	if len(result) != len(expected) {
		t.Fatalf("unexpected length: %v", result)
	}
	for i, v := range result {
		if v != expected[i] {
			t.Fatalf("unexpected element: %v", v)
		}
	}
}

func TestSlicesDifference(t *testing.T) {
	s1 := []int{1, 2, 3, 4, 5}
	s2 := []int{4, 5, 6, 7, 8}
	result := utils.SlicesDifference(s1, s2)
	expected := []int{1, 2, 3}
	if len(result) != len(expected) {
		t.Fatalf("unexpected length: %v", result)
	}
	for i, v := range result {
		if v != expected[i] {
			t.Fatalf("unexpected element: %v", v)
		}
	}
}
