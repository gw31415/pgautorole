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
	// roles, roleIDs, levelCourseMapを操作するためのロック
	rw sync.RWMutex
	// サーバーID
	guildID string
	// コース関連ロール情報
	roles []*discordgo.Role
	// コース関連ロールのID一覧
	roleIDs []string
	// コースレベルロールIDからコースロールIDのマップ
	// キーとして存在すればコースレベルロールと判断する
	levelCourseMap map[string]string
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
			croles := utils.SlicesFilter(member.Roles, func(r string) bool {
				return slices.Contains(m.roleIDs, r)
			})
			if len(croles) == 0 {
				continue
			}

			for _, rid := range croles {
				cid := m.levelCourseMap[rid]
				if cid == "" {
					cid = rid

					clids := []string{}
					for v, i := range m.levelCourseMap {
						if i == cid {
							clids = append(clids, v)
						}
					}

					dups := utils.SlicesFilter(member.Roles, func(id string) bool {
						return slices.Contains(clids, id)
					})
					if len(dups) > 1 {
						course := m.getCourse(cid)
						slog.Warn("Duplicated course level roles", "USER", member.User.ID, "USER_NAME", member.User.GlobalName, "COURSE", *course)
					}
				} else {
					if !slices.Contains(member.Roles, cid) {
						course := m.getCourse(cid)
						slog.Debug("", "COURSE", course)
						slog.Info("Adding missing course role", "USER", member.User.ID, "USER_NAME", member.User.GlobalName, "COURSE", *course)
						s.GuildMemberRoleAdd(m.guildID, member.User.ID, cid)
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
	cids := []string{}
	// コースレベルロールのID
	clids := []string{}
	// コースレベルロールIDからコースロールIDのマップ
	levelCourseMap := map[string]string{}

	// サーバー内全てのロールの内からコース関連ロールを抽出
r:
	for _, r := range allroles {
		c := internal.Course(r.Name)
		clid := []string{}
		// 対応するコースレベルロールのIDを取得
		// 過不足があればスキップ
		for _, cl := range c.CourseLevels() {
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
		slog.Info("Course detected:", "COURSE_NAME", r.Name)
		cids = append(cids, r.ID)
		clids = append(clids, clid...)
		for _, cl := range clid {
			levelCourseMap[cl] = r.ID
		}
	}

	m.roleIDs = []string{}
	m.roleIDs = append(m.roleIDs, cids...)
	m.roleIDs = append(m.roleIDs, clids...)

	m.roles = utils.SlicesFilter(allroles, func(r *discordgo.Role) bool {
		return slices.Contains(m.roleIDs, r.ID)
	})

	m.levelCourseMap = levelCourseMap
}

func (m *courseManager) getID(cl *internal.CourseLevel) string {
	for _, r := range m.roles {
		if r.Name == cl.String() {
			return r.ID
		}
	}
	return ""
}
func (m *courseManager) getCourse(id string) *internal.Course {
	if m.levelCourseMap[id] != "" {
		return nil
	}
	for _, r := range m.roles {
		if r.ID == id {
			course := internal.Course(r.Name)
			return &course
		}
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

func (m *courseManager) MemberRoleUpdateHandler(s *discordgo.Session, u *discordgo.GuildMemberUpdate) {
	m.rw.RLock()
	defer m.rw.RUnlock()

	// ロール変更前後のコース関連ロールを取得
	roles := utils.SlicesFilter(u.Member.Roles, func(id string) bool {
		return slices.Contains(m.roleIDs, id)
	})

	rolesBefore := []string{}
	if u.BeforeUpdate != nil {
		rolesBefore = utils.SlicesFilter(u.BeforeUpdate.Roles, func(id string) bool {
			return slices.Contains(m.roleIDs, id)
		})
	}

	// 追加されたロールと削除されたロールを取得
	added := utils.SlicesFilter(roles, func(r string) bool {
		return !slices.Contains(rolesBefore, r)
	})
	removed := utils.SlicesFilter(rolesBefore, func(r string) bool {
		return !slices.Contains(roles, r)
	})

	for _, rid := range added {
		wasCourse := false
		cid := m.levelCourseMap[rid]
		var clids []string
		if cid == "" {
			wasCourse = true
			cid = rid
		}

		for v, i := range m.levelCourseMap {
			if i == cid {
				clids = append(clids, v)
			}
		}

		cl := utils.SlicesFilter(u.Member.Roles, func(id string) bool {
			return slices.Contains(clids, id)
		})

		if wasCourse {
			// コースロールが追加された時

			if len(cl) == 0 {
				// コースレベルロールの初期値はアプレンティス
				c := m.getCourse(cid)
				initialCourseLevel := c.With(internal.Apprentice)
				s.GuildMemberRoleAdd(u.GuildID, u.User.ID, m.getID(&initialCourseLevel))
			} else {
				// コースレベルロールが複数ある時はとりあえず1つにする
				// 再度イベント発火するので、再帰的に処理される
				s.GuildMemberRoleRemove(u.GuildID, u.User.ID, cl[0])
			}

		} else {
			// 他のコースレベルロールを削除
			idsToRemove := utils.SlicesFilter(u.Member.Roles, func(id string) bool {
				if slices.Contains(u.Member.Roles, cid) && id == rid {
					return false
				}
				return slices.Contains(clids, id)
			})
			if len(idsToRemove) > 0 {
				// コースレベルロールが複数ある時はとりあえず1つにする
				// 再度イベント発火するので、再帰的に処理される
				s.GuildMemberRoleRemove(u.GuildID, u.User.ID, idsToRemove[0])
			}
		}
	}
	for _, rid := range removed {
		wasCourse := false
		cid := m.levelCourseMap[rid]
		var clids []string
		if cid == "" {
			wasCourse = true
			cid = rid
		}

		for v, i := range m.levelCourseMap {
			if i == cid {
				clids = append(clids, v)
			}
		}

		if wasCourse {
			// コースロールが削除された時
			// 他のコースレベルロールを削除
			idsToRemove := utils.SlicesFilter(u.Member.Roles, func(id string) bool {
				return slices.Contains(clids, id)
			})
			for _, id := range idsToRemove {
				s.GuildMemberRoleRemove(u.GuildID, u.User.ID, id)
			}
		} else {
			// コースレベルロールが削除された時
			if slices.Contains(u.Member.Roles, cid) {
				cl := utils.SlicesFilter(u.Member.Roles, func(id string) bool {
					return slices.Contains(clids, id)
				})
				if len(cl) == 0 {
					// コースレベルロールが0になる時は復元する
					s.GuildMemberRoleAdd(u.GuildID, u.User.ID, rid)
				} else if len(cl) > 1 {
					// コースレベルロールが複数ある時はとりあえず1つにする
					// 再度イベント発火するので、再帰的に処理される
					s.GuildMemberRoleRemove(u.GuildID, u.User.ID, cl[0])
				}
			}
		}
	}
}
