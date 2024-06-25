package internal_test

import (
	"slices"
	"testing"

	"github.com/gw31415/pgautorole/course/internal"
)

func TestCourseLevelNames(t *testing.T) {
	c := internal.CourseName("test")
	cls := c.CourseLevelNames()
	if len(cls) != 4 {
		t.Errorf("unexpected length: %d", len(cls))
	}
	expected := []string{
		"test-アプレンティス",
		"test-アシスタント",
		"test-ノーマル",
		"test-リード",
	}
	for _, n := range expected {
		if !slices.ContainsFunc(cls, func(cl internal.CourseLevelName) bool {
			return cl.String() == n
		}) {
			t.Errorf("unexpected course: %s", n)
		}
	}
}

func TestWith(t *testing.T) {
	c := internal.CourseName("test")
	cl := c.With(internal.Apprentice)
	if cl.String() != "test-アプレンティス" {
		t.Errorf("unexpected course: %s", cl.String())
	}
}

func TestParseCourseLevel(t *testing.T) {
	cl := internal.ParseCourseLevel("test-アプレンティス")
	if cl == nil {
		t.Error("unexpected nil")
	}
	if cl.Course != "test" {
		t.Errorf("unexpected course: %s", cl.Course)
	}
	if cl.Level != internal.Apprentice {
		t.Errorf("unexpected level: %s", cl.Level)
	}
}
