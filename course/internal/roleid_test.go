package internal_test

import (
	"slices"
	"testing"

	"github.com/gw31415/pgautorole/course/internal"
)

var repo internal.RoleIDRepository

func init() {
	var err error
	m := make(map[string][]string)

	m["course1"] = []string{"c1-level1", "c1-level2"}
	m["course2"] = []string{"c2-level1", "c2-level2"}
	repo, err = internal.NewRoleIDRepository(m)
	if err != nil {
		panic(err)
	}
}

func TestFindID(t *testing.T) {
	t.Run("CourseRelatedRoleID", func(t *testing.T) {
		id := "course1"
		cr := repo.FindID(id)
		if cr == nil {
			t.Fatalf("unexpected nil")
		}
		if cr.String() != id {
			t.Fatalf("unexpected string: %v", cr.String())
		}
		switch cr.(type) {
		case *internal.CourseRoleID:
		default:
			t.Fatalf("unexpected type: %T", cr)
		}
	})

	t.Run("CourseLevelRoleID", func(t *testing.T) {
		id := "c1-level1"
		cr := repo.FindID(id)
		if cr == nil {
			t.Fatalf("unexpected nil")
		}
		if cr.String() != id {
			t.Fatalf("unexpected string: %v", cr.String())
		}
		switch cr.(type) {
		case *internal.CourseLevelRoleID:
		default:
			t.Fatalf("unexpected type: %T", cr)
		}
	})
	t.Run("InvalidRoleID", func(t *testing.T) {
		id := "invalid"
		cr := repo.FindID(id)
		if cr != nil {
			t.Fatalf("unexpected non-nil")
		}
	})
}

func TestFilterCourseLevelRoleID(t *testing.T) {
	t.Run("FilterCourseLevelRoleID", func(t *testing.T) {
		ids := []string{"c1-level1", "c1-level2", "course1", "hoge"}
		cr := repo.FilterIDs(ids)
		if len(cr) != 3 {
			t.Fatalf("unexpected length: %v", len(cr))
		}
	})
}

func TestGetCourseLevelIDs(t *testing.T) {
	t.Run("Course", func(t *testing.T) {
		id := "course1"
		cr := repo.FindID(id)
		if cr == nil {
			t.Fatalf("unexpected nil")
		}
		cls := cr.GetCourseLevelIDs()
		if len(cls) != 2 {
			t.Fatalf("unexpected length: %v", len(cls))
		}
		expected := []string{"c1-level1", "c1-level2"}
		if slices.ContainsFunc(cls, func(i *internal.CourseLevelRoleID) bool {
			return !slices.Contains(expected, i.String())
		}) {
			t.Fatalf("unexpected value: %v", cls)
		}
	})
	t.Run("CourseLevel", func(t *testing.T) {
		id := "c1-level1"
		cr := repo.FindID(id)
		if cr == nil {
			t.Fatalf("unexpected nil")
		}
		cls := cr.GetCourseLevelIDs()
		if len(cls) != 2 {
			t.Fatalf("unexpected length: %v", len(cls))
		}
		expected := []string{"c1-level1", "c1-level2"}
		if slices.ContainsFunc(cls, func(i *internal.CourseLevelRoleID) bool {
			return !slices.Contains(expected, i.String())
		}) {
			t.Fatalf("unexpected value: %v", cls)
		}
	})
}

func TestGetCourseRoleID(t *testing.T) {
	t.Run("Course", func(t *testing.T) {
		id := "course1"
		cr := repo.FindID(id)
		if cr == nil {
			t.Fatalf("unexpected nil")
		}
		if cr.GetCourseRoleID() != cr {
			t.Fatalf("unexpected nil")
		}
	})
	t.Run("CourseLevel", func(t *testing.T) {
		id := "c1-level1"
		cr := repo.FindID(id)
		if cr == nil {
			t.Fatalf("unexpected nil")
		}
		l := cr.GetCourseRoleID()
		if l.String() != "course1" {
			t.Fatalf("unexpected string: %v", l.String())
		}
	})
}
