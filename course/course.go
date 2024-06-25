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

type courseManager struct {
	// updatingUsersを操作するためのロック
	usersSync sync.RWMutex
	// ユーザー情報を更新中かどうか
	updatingUsers map[string]bool
	// roles, roleIDs, levelCourseMapを操作するためのロック
	guildsync sync.RWMutex
	// サーバーID
	guildID string
	// IDからロール情報へのマップ
	roles map[string]*discordgo.Role
	// コース関連ロールの情報
	*internal.RoleIDRepository
}

// コースマネージャを生成
func NewCourseManager(guildID string) CourseManager {
	return &courseManager{
		guildID:       guildID,
		updatingUsers: make(map[string]bool),
	}
}

// ロール名からコースIDを逆検索
func (m *courseManager) reverseNameToRoleID(cl *internal.CourseLevelName) string {
	for _, r := range m.roles {
		if r.Name == cl.String() {
			return r.ID
		}
	}
	return ""
}

// コース関連ロールIDからロールの表示名を取得
func (m *courseManager) getCourseName(id interface{}) *internal.CourseName {
	switch id := id.(type) {
	case internal.CourseRoleID:
		name := internal.CourseName(m.roles[id.String()].Name)
		return &name
	case internal.CourseLevelRoleID:
		return m.getCourseName(m.GetCourseRoleID(id))
	}
	return nil
}

func (m *courseManager) UnsafeCheckCourseRoles(s *discordgo.Session) {
	m.syncRoles(s)
	m.guildsync.RLock()
	defer m.guildsync.RUnlock()

	slog.Info("Refreshing members' course roles...")
	after := ""
	for {
		members, err := s.GuildMembers(m.guildID, after, 1000)
		if err != nil || len(members) == 0 {
			break
		}
		slog.Debug("Paging members", "COUNT", len(members))
		for _, member := range members {
			croles := m.FilterCourseRelatedRoleIDs(member.Roles)
			if len(croles) == 0 {
				continue
			}

			for _, id := range croles {
				switch cid := m.ClassifyCourseRelatedID(id).(type) {
				case internal.CourseRoleID:
					clids := m.GetSameCourseLevels(cid)

					dups := FilterMemberRoles(member, clids)
					if len(dups) > 1 {
						cname := m.getCourseName(cid)
						slog.Warn("Duplicated course level roles", "USER", member.User.ID, "USER_NAME", member.User.GlobalName, "COURSE", *cname)
					}
				case internal.CourseLevelRoleID:
					if !slices.Contains(member.Roles, cid.String()) {
						cname := m.getCourseName(cid)
						slog.Info("Adding missing course role", "USER", member.User.ID, "USER_NAME", member.User.GlobalName, "COURSE", *cname)
						s.GuildMemberRoleAdd(m.guildID, member.User.ID, cid.String())
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
	m.guildsync.Lock()
	defer m.guildsync.Unlock()

	// ロールの取得
	allroles, err := s.GuildRoles(m.guildID)
	if err != nil {
		slog.Error("failed to get roles:", err)
		return
	}

	roles := make(map[string]*discordgo.Role)
	// コースレベルロールIDからコースロールIDのマップ
	c2lMap := make(map[string][]string)

	// サーバー内全てのロールの内からコース関連ロールを抽出
r:
	for _, r := range allroles {
		c := internal.CourseName(r.Name)
		clid := []string{}
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
			clid = append(clid, clr[0].ID)
		}
		// 対応するコースレベルロールが過不足なければコースとして登録
		slog.Info("Course detected:", "COURSE_NAME", c)
		c2lMap[r.ID] = clid

		roles[r.ID] = r
		for _, id := range clid {
			idx := slices.IndexFunc(allroles, func(r *discordgo.Role) bool {
				return r.ID == id
			})
			roles[id] = allroles[idx]
		}
	}

	repo, err := internal.NewRoleIDRepository(c2lMap)
	if err != nil {
		slog.Error("failed to create role id repository:", err)
		return
	}
	m.RoleIDRepository = repo
	m.roles = roles
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

func FilterMemberRoles(member *discordgo.Member, list []internal.CourseLevelRoleID) []internal.CourseLevelRoleID {
	return utils.SlicesFilter(list, func(id internal.CourseLevelRoleID) bool {
		return slices.Contains(member.Roles, id.String())
	})
}

func (m *courseManager) MemberRoleUpdateHandler(s *discordgo.Session, u *discordgo.GuildMemberUpdate) {
	m.guildsync.RLock()
	defer m.guildsync.RUnlock()

	// ユーザー更新中かどうか取得
	m.usersSync.RLock()
	updating := m.updatingUsers[u.User.ID]
	m.usersSync.RUnlock()

	if updating {
		// 他のタスクで更新中のユーザーはスキップ
		return
	}

	// ロール変更前後のコース関連ロールを取得
	roles := m.FilterCourseRelatedRoleIDs(u.Member.Roles)

	rolesBefore := []internal.CourseRelatedRoleID{}
	if u.BeforeUpdate != nil {
		rolesBefore = m.FilterCourseRelatedRoleIDs(u.BeforeUpdate.Roles)
	}
	// } else {
	// 	return // TODO: ロール変更前の情報がない場合は追加ロールを正しく処理できない
	// }

	// 追加されたロールと削除されたロールを取得
	added := utils.SlicesDifference(roles, rolesBefore)
	removed := utils.SlicesDifference(rolesBefore, roles)

	// ユーザー更新中に設定
	m.usersSync.Lock()
	m.updatingUsers[u.User.ID] = true
	m.usersSync.Unlock()
	defer func() {
		m.usersSync.Lock()
		delete(m.updatingUsers, u.User.ID)
		m.usersSync.Unlock()
	}()

	for _, id := range added {
		switch id := m.ClassifyCourseRelatedID(id).(type) {
		case internal.CourseRoleID:
			// コースロールが追加された時
			course := id
			sameCourseLevels := m.GetSameCourseLevels(id)
			dups := FilterMemberRoles(u.Member, sameCourseLevels)
			// hasCourse := slices.Contains(u.Member.Roles, course.String())

			if len(dups) == 0 {
				// コースレベルロールの初期値はアプレンティス
				cname := m.getCourseName(course)
				initialCourseLevel := cname.With(internal.Apprentice)

				s.GuildMemberRoleAdd(u.GuildID, u.User.ID, m.reverseNameToRoleID(&initialCourseLevel))
			} else {
				// コースレベルロールが既にある時はとりあえず1つにする
				for _, i := range dups[1:] {
					s.GuildMemberRoleRemove(u.GuildID, u.User.ID, i.String())
				}
			}
		case internal.CourseLevelRoleID:
			// コースレベルロールが追加された時
			course := m.GetCourseRoleID(id)
			sameCourseLevels := m.GetSameCourseLevels(course)
			dups := FilterMemberRoles(u.Member, sameCourseLevels)
			hasCourse := slices.Contains(u.Member.Roles, course.String())

			// 他のコースレベルロールを削除
			for _, i := range dups {
				if i != id || !hasCourse {
					s.GuildMemberRoleRemove(u.GuildID, u.User.ID, i.String())
				}
			}
		}
	}
	for _, id := range removed {
		switch id := m.ClassifyCourseRelatedID(id).(type) {
		case internal.CourseRoleID:
			// コースロールが削除された時
			// course := id
			sameCourseLevels := m.GetSameCourseLevels(id)
			dups := FilterMemberRoles(u.Member, sameCourseLevels)
			// hasCourse := slices.Contains(u.Member.Roles, string(course))

			// 他のコースレベルロールを削除
			if len(dups) > 0 {
				for _, i := range dups {
					s.GuildMemberRoleRemove(u.GuildID, u.User.ID, i.String())
				}
			}
		case internal.CourseLevelRoleID:
			// コースレベルロールが削除された時
			course := m.GetCourseRoleID(id)
			sameCourseLevels := m.GetSameCourseLevels(course)
			dups := FilterMemberRoles(u.Member, sameCourseLevels)
			hasCourse := slices.Contains(u.Member.Roles, course.String())

			if len(dups) == 0 && hasCourse {
				// コースレベルロールが0になる時は復元する
				s.GuildMemberRoleAdd(u.GuildID, u.User.ID, id.String())
			} else if len(dups) > 1 {
				// コースレベルロールが複数ある時はとりあえず1つにする
				// またはコースロールがない時は完全に削除する
				for ii, i := range dups {
					if ii == 0 || !hasCourse {
						continue
					}
					s.GuildMemberRoleRemove(u.GuildID, u.User.ID, i.String())
				}
			}
		}
	}
}
