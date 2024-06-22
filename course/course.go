package course

import (
	"log/slog"
	"slices"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/gw31415/pgautorole/course/internal"
	"github.com/gw31415/pgautorole/internal/utils"
)

// コースマネージャ
type CourseManager interface {
	// ロール情報を同期
	ReadyHandler(s *discordgo.Session, u *discordgo.Ready)
	// ロール情報を同期
	GuildCreateHandler(s *discordgo.Session, u *discordgo.GuildCreate)
	// ロール情報を同期
	GuildRoleCreateHandler(s *discordgo.Session, u *discordgo.GuildRoleCreate)
	// ロール情報を同期
	GulidRoleUpdateHandler(s *discordgo.Session, u *discordgo.GuildRoleUpdate)
	// ロール情報を同期
	GuildRoleDeleteHandler(s *discordgo.Session, u *discordgo.GuildRoleDelete)
	// ロール変更時にコース関連ロールを操作するハンドラ
	MemberRoleUpdateHandler(s *discordgo.Session, m *discordgo.GuildMemberUpdate)

	// WARN: MemberRoleUpdateHandlerのハンドラを切って使わなければならない
	// メンバーのコース関連ロールの競合状態をチェックし、更新可能なら更新する
	UnsafeCheckCourseRoles(s *discordgo.Session)
}

// コース関連ロールのID
type CourseRelatedRoleID string

// コースID
type CourseID CourseRelatedRoleID

// コースレベルロールID
type CourseLevelID CourseRelatedRoleID

type courseManager struct {
	// roles, roleIDs, levelCourseMapを操作するためのロック
	rw sync.RWMutex
	// サーバーID
	guildID string
	// コース関連ロール情報
	roles []*discordgo.Role
	// コース関連ロールのID一覧
	roleIDs []CourseRelatedRoleID
	// コースレベルロールIDからコースロールIDのマップ
	// キーとして存在すればコースレベルロールと判断する
	levelCourseMap map[CourseLevelID]CourseID
}

// コース関連ロールならparseCourseRelatedIDにパースし、そうでなければnilを返す
func (m *courseManager) parseCourseRelatedID(id string) *CourseRelatedRoleID {
	if slices.Contains(m.roleIDs, (CourseRelatedRoleID)(id)) {
		return (*CourseRelatedRoleID)(&id)
	}
	return nil
}

// コースIDまたはコースレベルロールIDを分類する
func (m *courseManager) classifyCourseRelatedID(id CourseRelatedRoleID) interface{} {
	if m.levelCourseMap[(CourseLevelID)(id)] != "" {
		return CourseLevelID(id)
	} else {
		return CourseID(id)
	}
}

// IDからコース関連IDを抜き出す
func (m *courseManager) filterCourseRelatedRoleIDs(roles []string) []CourseRelatedRoleID {
	return utils.SlicesFilterMap(roles, m.parseCourseRelatedID)
}

// コースIDからコースレベルロールIDを取得
func (m *courseManager) getSameCourseLevels(cid CourseID) (ids []CourseLevelID) {
	for v, i := range m.levelCourseMap {
		if i == cid {
			ids = append(ids, v)
		}
	}
	return
}

// コースマネージャを生成
func NewCourseManager(guildID string) CourseManager {
	return &courseManager{
		guildID: guildID,
	}
}

func (m *courseManager) UnsafeCheckCourseRoles(s *discordgo.Session) {
	m.syncRoles(s)
	m.rw.RLock()
	defer m.rw.RUnlock()

	slog.Info("Refreshing members' course roles...")
	after := ""
	for {
		members, err := s.GuildMembers(m.guildID, after, 1000)
		if err != nil || len(members) == 0 {
			break
		}
		slog.Debug("Paging members", "COUNT", len(members))
		for _, member := range members {
			croles := m.filterCourseRelatedRoleIDs(member.Roles)
			if len(croles) == 0 {
				continue
			}

			for _, id := range croles {
				switch cid := m.classifyCourseRelatedID(id).(type) {
				case CourseID:
					clids := m.getSameCourseLevels(cid)

					dups := FilterMemberRoles(member, clids)
					if len(dups) > 1 {
						course := m.getCourse(cid)
						slog.Warn("Duplicated course level roles", "USER", member.User.ID, "USER_NAME", member.User.GlobalName, "COURSE", *course)
					}
				case CourseLevelID:
					if !slices.Contains(member.Roles, string(cid)) {
						course := m.getCourse(cid)
						slog.Info("Adding missing course role", "USER", member.User.ID, "USER_NAME", member.User.GlobalName, "COURSE", *course)
						s.GuildMemberRoleAdd(m.guildID, member.User.ID, string(cid))
					}
				}
			}
		}
		after = members[len(members)-1].User.ID
	}
}

// サーバーのロール情報を同期
func (m *courseManager) syncRoles(s *discordgo.Session) {
	slog.Info("Syncing roles...")
	m.rw.Lock()
	defer m.rw.Unlock()

	// ロールの取得
	allroles, err := s.GuildRoles(m.guildID)
	if err != nil {
		slog.Error("failed to get roles:", err)
		return
	}

	// コースロールのID
	cids := []CourseID{}
	// コースレベルロールのID
	clids := []CourseLevelID{}
	// コースレベルロールIDからコースロールIDのマップ
	levelCourseMap := map[CourseLevelID]CourseID{}

	// サーバー内全てのロールの内からコース関連ロールを抽出
r:
	for _, r := range allroles {
		c := internal.CourseName(r.Name)
		clid := []CourseLevelID{}
		// 対応するコースレベルロールのIDを取得
		// 過不足があればスキップ
		for _, cl := range c.CourseLevelNames() {
			name := cl.String()
			clr := utils.SlicesFilter(allroles, func(r *discordgo.Role) bool {
				return r.Name == name
			})
			if len(clr) != 1 {
				continue r
			}
			clid = append(clid, CourseLevelID(clr[0].ID))
		}
		// 対応するコースレベルロールが過不足なければコースとして登録
		slog.Info("Course detected:", "COURSE_NAME", r.Name)
		cids = append(cids, CourseID(r.ID))
		clids = append(clids, clid...)
		for _, cl := range clid {
			levelCourseMap[cl] = CourseID(r.ID)
		}
	}

	m.roleIDs = []CourseRelatedRoleID{}
	for _, id := range cids {
		m.roleIDs = append(m.roleIDs, CourseRelatedRoleID(id))
	}
	for _, id := range clids {
		m.roleIDs = append(m.roleIDs, CourseRelatedRoleID(id))
	}

	m.roles = utils.SlicesFilter(allroles, func(r *discordgo.Role) bool {
		return slices.Contains(m.roleIDs, CourseRelatedRoleID(r.ID))
	})

	m.levelCourseMap = levelCourseMap
}

func (m *courseManager) getID(cl *internal.CourseLevelName) string {
	for _, r := range m.roles {
		if r.Name == cl.String() {
			return r.ID
		}
	}
	return ""
}

func (m *courseManager) getCourse(id interface{}) *internal.CourseName {
	switch id := id.(type) {
	case CourseID:
		for _, r := range m.roles {
			if r.ID == string(id) {
				course := internal.CourseName(r.Name)
				return &course
			}
		}
	case CourseLevelID:
		return m.getCourse(m.levelCourseMap[id])
	}
	return nil
}

func (m *courseManager) ReadyHandler(s *discordgo.Session, u *discordgo.Ready) {
	m.syncRoles(s)
}
func (m *courseManager) GuildCreateHandler(s *discordgo.Session, u *discordgo.GuildCreate) {
	m.syncRoles(s)
}
func (m *courseManager) GuildRoleCreateHandler(s *discordgo.Session, u *discordgo.GuildRoleCreate) {
	m.syncRoles(s)
}
func (m *courseManager) GulidRoleUpdateHandler(s *discordgo.Session, u *discordgo.GuildRoleUpdate) {
	m.syncRoles(s)
}
func (m *courseManager) GuildRoleDeleteHandler(s *discordgo.Session, u *discordgo.GuildRoleDelete) {
	m.syncRoles(s)
}

func FilterMemberRoles[T ~string](member *discordgo.Member, list []T) []T {
	return utils.SlicesFilter(list, func(id T) bool {
		return slices.Contains(member.Roles, string(id))
	})
}

func (m *courseManager) MemberRoleUpdateHandler(s *discordgo.Session, u *discordgo.GuildMemberUpdate) {
	m.rw.RLock()
	defer m.rw.RUnlock()

	// ロール変更前後のコース関連ロールを取得
	roles := m.filterCourseRelatedRoleIDs(u.Member.Roles)

	rolesBefore := []CourseRelatedRoleID{}
	if u.BeforeUpdate != nil {
		rolesBefore = m.filterCourseRelatedRoleIDs(u.BeforeUpdate.Roles)
	}

	// 追加されたロールと削除されたロールを取得
	added := utils.SlicesDifference(rolesBefore, roles)
	removed := utils.SlicesDifference(roles, rolesBefore)

	for _, id := range added {
		switch id := m.classifyCourseRelatedID(id).(type) {
		case CourseID:
			// コースロールが追加された時
			course := id
			sameCourseLevels := m.getSameCourseLevels(id)
			dups := FilterMemberRoles(u.Member, sameCourseLevels)
			// hasCourse := slices.Contains(u.Member.Roles, string(course))

			if len(dups) == 0 {
				// コースレベルロールの初期値はアプレンティス
				c := m.getCourse(course)
				initialCourseLevel := c.With(internal.Apprentice)
				s.GuildMemberRoleAdd(u.GuildID, u.User.ID, m.getID(&initialCourseLevel))
			} else {
				// コースレベルロールが複数ある時はとりあえず1つにする
				// NOTE: 再度イベント発火するので、再帰的に処理される
				s.GuildMemberRoleRemove(u.GuildID, u.User.ID, string(dups[0]))
			}
		case CourseLevelID:
			// コースレベルロールが追加された時
			course := m.levelCourseMap[id]
			sameCourseLevels := m.getSameCourseLevels(course)
			dups := FilterMemberRoles(u.Member, sameCourseLevels)
			hasCourse := slices.Contains(u.Member.Roles, string(course))

			// 他のコースレベルロールを削除
			for _, i := range dups {
				if i != id || !hasCourse {
					s.GuildMemberRoleRemove(u.GuildID, u.User.ID, string(id))
					// 再度イベント発火するので、再帰的に処理される
					break
				}
			}
		}
	}
	for _, id := range removed {
		switch id := m.classifyCourseRelatedID(id).(type) {
		case CourseID:
			// コースロールが削除された時
			// course := id
			sameCourseLevels := m.getSameCourseLevels(id)
			dups := FilterMemberRoles(u.Member, sameCourseLevels)
			// hasCourse := slices.Contains(u.Member.Roles, string(course))

			// 他のコースレベルロールを削除
			// NOTE: 再度イベント発火するので、再帰的に処理される
			if len(dups) > 0 {
				s.GuildMemberRoleRemove(u.GuildID, u.User.ID, string(dups[0]))
			}
		case CourseLevelID:
			// コースレベルロールが削除された時
			course := m.levelCourseMap[id]
			sameCourseLevels := m.getSameCourseLevels(course)
			dups := FilterMemberRoles(u.Member, sameCourseLevels)
			hasCourse := slices.Contains(u.Member.Roles, string(course))

			if len(dups) == 0 && hasCourse {
				// コースレベルロールが0になる時は復元する
				s.GuildMemberRoleAdd(u.GuildID, u.User.ID, string(id))
			} else if !hasCourse && len(dups) == 1 || len(dups) > 1 {
				// コースレベルロールが複数ある時はとりあえず1つにする
				// またはコースロールがない時は完全に削除する
				// NOTE: 再度イベント発火するので、再帰的に処理される
				s.GuildMemberRoleRemove(u.GuildID, u.User.ID, string(dups[0]))
			}
		}
	}
}
