package newbie

import (
	"log/slog"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
)

type NewbieManager interface {
	// 会員ロール変化時に新規会員ロールを操作するハンドラ
	MemberRoleUpdateHandler(s *discordgo.Session, m *discordgo.GuildMemberUpdate)
	// 新規会員ロールを持っているメンバーから新規会員ロールを削除
	FilterNewbieRoles(s *discordgo.Session)
}

type newbieManager struct {
	newbieRoleID   string
	memberRoleID   string
	newbieDuration time.Duration
}

func NewNewbieManager(newbieRoleID, memberRoleID string, newbieDuration time.Duration) NewbieManager {
	return &newbieManager{
		newbieRoleID:   newbieRoleID,
		memberRoleID:   memberRoleID,
		newbieDuration: newbieDuration,
	}
}

// userが新規会員に該当するかどうか
func (n *newbieManager) checkNewbie(member *discordgo.Member) (bool, error) {
	age := time.Since(member.JoinedAt)
	// 会員でない場合は新規会員ではない
	if !slices.Contains(member.Roles, n.memberRoleID) {
		return false, nil
	}
	// JoinedAtからNEWBIE_DURAITON経っていない新規会員
	return age < n.newbieDuration, nil
}

func (n *newbieManager) MemberRoleUpdateHandler(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	// 会員ロールが付与された時
	if slices.Contains(m.Roles, n.memberRoleID) && (m.BeforeUpdate == nil || m.BeforeUpdate.Roles == nil || !slices.Contains(m.BeforeUpdate.Roles, n.memberRoleID)) {
		isNewbie, err := n.checkNewbie(m.Member)
		if err == nil && isNewbie {
			slog.Info("Add newbie role", "m.User.ID", m.User.ID)
			s.GuildMemberRoleAdd(m.GuildID, m.User.ID, n.newbieRoleID)
		}
	}
	// 新規会員ロールが手動付与された時
	if slices.Contains(m.Roles, n.newbieRoleID) && (m.BeforeUpdate == nil || m.BeforeUpdate.Roles == nil || !slices.Contains(m.BeforeUpdate.Roles, n.newbieRoleID)) {
		isNewbie, err := n.checkNewbie(m.Member)
		if err == nil && !isNewbie {
			// 条件にあてはまらない場合キャンセル
			slog.Info("Refuse newbie role", "m.User.ID", m.User.ID)
			s.GuildMemberRoleRemove(m.GuildID, m.User.ID, n.newbieRoleID)
		}
	}
}

const MEMBERS_PER_REQUEST = 1000

func (n *newbieManager) FilterNewbieRoles(s *discordgo.Session) {
	for _, guild := range s.State.Guilds {
		after := ""
		for {
			// MEMBER_PER_REQUESTずつメンバーを取得
			m, err := s.GuildMembers(guild.ID, after, MEMBERS_PER_REQUEST)
			// メンバーが取得できなかった場合は終了
			if err != nil || len(m) == 0 {
				break
			}

			// 全メンバーに対して処理
			for _, member := range m {
				// 会員ロールと新規会員ロールを持っている場合
				if slices.Contains(member.Roles, n.memberRoleID) && slices.Contains(member.Roles, n.newbieRoleID) {
					// 新規会員でない場合は新規会員ロールを削除
					isNewbie, err := n.checkNewbie(member)
					if err == nil && !isNewbie {
						slog.Info("Remove newbie role", "member.User.ID", member.User.ID)
						err := s.GuildMemberRoleRemove(guild.ID, member.User.ID, n.newbieRoleID)
						if err != nil {
							slog.Error("Failed to remove newbie role", err)
						}
					}
				}
			}

			after = m[len(m)-1].User.ID
		}
	}
}
