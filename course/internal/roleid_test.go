package internal_test

import (
	"slices"
	"testing"

	"github.com/gw31415/pgautorole/course/internal"
)

var repo *internal.RoleIDRepository

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

func TestParseCourseRelatedID(t *testing.T) {
	t.Run("CourseRelatedRoleID", func(t *testing.T) {
		id := "course1"
		cr := repo.ParseCourseRelatedID(id)
		if cr == nil {
			t.Fatalf("unexpected nil")
		}
		if cr.String() != id {
			t.Fatalf("unexpected string: %v", cr.String())
		}
	})

	t.Run("CourseLevelRoleID", func(t *testing.T) {
		id := "c1-level1"
		cr := repo.ParseCourseRelatedID(id)
		if cr == nil {
			t.Fatalf("unexpected nil")
		}
		if cr.String() != id {
			t.Fatalf("unexpected string: %v", cr.String())
		}
	})
	t.Run("InvalidRoleID", func(t *testing.T) {
		id := "invalid"
		cr := repo.ParseCourseRelatedID(id)
		if cr != nil {
			t.Fatalf("unexpected non-nil")
		}
	})
}

func TestClassifyCourseRelatedID(t *testing.T) {
	t.Run("CourseRoleID", func(t *testing.T) {
		id := "course1"
		cr := repo.ClassifyCourseRelatedID(*repo.ParseCourseRelatedID(id))
		switch cr.(type) {
		case internal.CourseRoleID:
		default:
			t.Fatalf("unexpected type: %T", cr)
		}
	})

	t.Run("CourseLevelRoleID", func(t *testing.T) {
		id := "c1-level1"
		cr := repo.ClassifyCourseRelatedID(*repo.ParseCourseRelatedID(id))
		switch cr.(type) {
		case internal.CourseLevelRoleID:
		default:
			t.Fatalf("unexpected type: %T", cr)
		}
	})
}

func TestFilterCourseLevelRoleID(t *testing.T) {
	t.Run("FilterCourseLevelRoleID", func(t *testing.T) {
		ids := []string{"c1-level1", "c1-level2", "course1", "hoge"}
		cr := repo.FilterCourseRelatedRoleIDs(ids)
		if len(cr) != 3 {
			t.Fatalf("unexpected length: %v", len(cr))
		}
	})
}

func TestGetSameCourseRelatedRoleID(t *testing.T) {
	t.Run("GetSameCourseRelatedRoleID", func(t *testing.T) {
		id := "course1"
		ids := repo.GetSameCourseLevels(internal.CourseRoleID{*repo.ParseCourseRelatedID(id)})
		if len(ids) != 2 {
			t.Fatalf("unexpected length: %v", len(ids))
		}
		expected := []string{"c1-level1", "c1-level2"}
		for _, n := range expected {
			valid := slices.ContainsFunc(ids, func(v internal.CourseLevelRoleID) bool {
				return v.String() == n
			})
			if !valid {
				t.Fatalf("element %v is unavailable", n)
			}
		}
	})
}

func TestGetCourseRoleID(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		id := "c1-level1"
		cr := internal.CourseLevelRoleID{*repo.ParseCourseRelatedID(id)}
		c := repo.GetCourseRoleID(cr)
		if c.String() != "course1" {
			t.Fatalf("unexpected string: %v", c.String())
		}
	})
}
