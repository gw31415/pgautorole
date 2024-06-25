package internal

import (
	"errors"

	"github.com/gw31415/pgautorole/internal/utils"
)

// コース関連ロールのID
type CourseRelatedRoleID struct {
	id string
}

func (c CourseRelatedRoleID) String() string {
	return c.id
}

// コースID
type CourseRoleID struct{ CourseRelatedRoleID }

// コースレベルロールID
type CourseLevelRoleID struct{ CourseRelatedRoleID }

// コース関連ロールIDの管理構造体
type RoleIDRepository struct {
	// コースレベルロールIDからコースロールIDのマップ
	// キーとして存在すればコースレベルロールと判断する
	levelCourseMap map[CourseRelatedRoleID]*CourseRoleID
	// コースIDからコースレベルロールIDのマップ
	courseLevelMap map[CourseRelatedRoleID][]CourseLevelRoleID
}

// コース関連ロールIDのマップからRoleIDRepositoryを生成
func NewRoleIDRepository(c2lMap map[string][]string) (*RoleIDRepository, error) {
	levelCourseMap := make(map[CourseRelatedRoleID]*CourseRoleID)
	courseLevelMap := make(map[CourseRelatedRoleID][]CourseLevelRoleID)
	for cs, lls := range c2lMap {
		c := CourseRelatedRoleID{id: cs}
		if courseLevelMap[c] != nil || levelCourseMap[c] != nil {
			return nil, errors.New("Duplicated ID")
		}
		ll := []CourseLevelRoleID{}
		for _, ls := range lls {
			l := CourseRelatedRoleID{id: ls}
			if levelCourseMap[l] != nil || levelCourseMap[c] != nil {
				return nil, errors.New("Duplicated ID")
			}
			ll = append(ll, CourseLevelRoleID{l})
			levelCourseMap[l] = &CourseRoleID{c}
		}
		courseLevelMap[c] = ll
	}
	return &RoleIDRepository{
		levelCourseMap,
		courseLevelMap,
	}, nil
}

// コース関連ロールならCourseRelatedRoleIDにパースし、そうでなければnilを返す
func (m *RoleIDRepository) ParseCourseRelatedID(id string) *CourseRelatedRoleID {
	rid := CourseRelatedRoleID{id}
	if m.levelCourseMap[rid] != nil || m.courseLevelMap[rid] != nil {
		return &rid
	}
	return nil
}

// コースIDまたはコースレベルロールIDを分類する
func (m *RoleIDRepository) ClassifyCourseRelatedID(id CourseRelatedRoleID) interface{} {
	if m.levelCourseMap[id] != nil {
		return CourseLevelRoleID{id}
	} else if m.courseLevelMap[id] != nil {
		return CourseRoleID{id}
	}
	return nil
}

// IDからコース関連IDを抜き出す
func (m *RoleIDRepository) FilterCourseRelatedRoleIDs(ids []string) []CourseRelatedRoleID {
	return utils.SlicesFilterMap(ids, m.ParseCourseRelatedID)
}

// コースIDからコースレベルロールIDを取得
func (m *RoleIDRepository) GetSameCourseLevels(id CourseRoleID) (ids []CourseLevelRoleID) {
	return m.courseLevelMap[id.CourseRelatedRoleID]
}

// コースレベルロールIDからコースIDを取得
func (m *RoleIDRepository) GetCourseRoleID(id CourseLevelRoleID) CourseRoleID {
	return *m.levelCourseMap[id.CourseRelatedRoleID]
}
