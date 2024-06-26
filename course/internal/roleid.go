package internal

import (
	"errors"
)

// ある瞬間のコース関連ロールIDの相互関係
type RoleIDRepository interface {
	// IDからコース関連ロールIDを取得
	FindID(id string) CourseRelatedRoleID
	// IDからコース関連ロールIDを抽出
	FilterIDs(ids []string) []CourseRelatedRoleID
}

// コース関連ロールIDの管理構造体
type roleIDRepository struct {
	// コースレベルロールIDからコースロールIDのマップ
	// キーとして存在すればコースレベルロールと判断する
	levelCourseMap map[string]*CourseRoleID
	// コースIDからIDのマップ
	courseLevelMap map[string][]*CourseLevelRoleID
}

// コース関連ロールIDのマップからRoleIDRepositoryを生成
func NewRoleIDRepository(c2lMap map[string][]string) (RoleIDRepository, error) {
	levelCourseMap := make(map[string]*CourseRoleID)
	courseLevelMap := make(map[string][]*CourseLevelRoleID)
	for cs, lls := range c2lMap {
		c := courseRelatedRoleID{id: cs}
		if courseLevelMap[c.id] != nil || levelCourseMap[c.id] != nil {
			return nil, errors.New("Duplicated ID")
		}
		ll := []*CourseLevelRoleID{}
		for _, ls := range lls {
			l := courseRelatedRoleID{id: ls}
			if levelCourseMap[l.id] != nil || levelCourseMap[c.id] != nil {
				return nil, errors.New("Duplicated ID")
			}
			ll = append(ll, &CourseLevelRoleID{l})
			levelCourseMap[l.id] = &CourseRoleID{c}
		}
		courseLevelMap[c.id] = ll
	}
	repo := roleIDRepository{
		levelCourseMap,
		courseLevelMap,
	}

	ids := []CourseRelatedRoleID{}
	// 各コース関連ロールIDにRoleIDRepositoryを設定
	for _, v := range repo.courseLevelMap {
		for _, vv := range v {
			vv.courseRelatedRoleID.repo = &repo
			ids = append(ids, vv)
		}
	}
	for _, v := range repo.levelCourseMap {
		v.courseRelatedRoleID.repo = &repo
		ids = append(ids, v)
	}
	return &repo, nil
}

// コース関連ロールならcourseRelatedRoleIDにパースし、そうでなければnilを返す
func (m *roleIDRepository) FindID(id string) CourseRelatedRoleID {
	if m.levelCourseMap[id] != nil {
		return &CourseLevelRoleID{courseRelatedRoleID{id, m}}
	} else if m.courseLevelMap[id] != nil {
		return &CourseRoleID{courseRelatedRoleID{id, m}}
	}
	return nil
}

// IDからコース関連IDを抜き出す
func (m *roleIDRepository) FilterIDs(ids []string) []CourseRelatedRoleID {
	arr := []CourseRelatedRoleID{}
	for _, id := range ids {
		if cr := m.FindID(id); cr != nil {
			arr = append(arr, cr)
		}
	}
	return arr
}

// コース関連ロールID
type CourseRelatedRoleID interface {
	// IDを文字列で取得
	String() string
	// コース内のレベルを取得
	GetCourseLevelIDs() []*CourseLevelRoleID
	// 該当するコースを取得
	GetCourseRoleID() *CourseRoleID
}

type courseRelatedRoleID struct {
	id   string
	repo *roleIDRepository
}

func (c *courseRelatedRoleID) String() string {
	return c.id
}

// コースID
type CourseRoleID struct{ courseRelatedRoleID }

func (c *CourseRoleID) GetCourseLevelIDs() []*CourseLevelRoleID {
	return c.repo.courseLevelMap[c.id]
}
func (c *CourseRoleID) GetCourseRoleID() *CourseRoleID {
	return c
}

// コースレベルロールID
type CourseLevelRoleID struct{ courseRelatedRoleID }

func (c *CourseLevelRoleID) GetCourseLevelIDs() []*CourseLevelRoleID {
	course := c.GetCourseRoleID()
	return course.GetCourseLevelIDs()
}
func (c *CourseLevelRoleID) GetCourseRoleID() *CourseRoleID {
	return c.repo.levelCourseMap[c.id]
}

// コース関連ロールIDの比較
func Equal(a CourseRelatedRoleID, b CourseRelatedRoleID) bool {
	return a.String() == b.String()
}
