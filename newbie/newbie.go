package newbie

import (
	"log/slog"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gw31415/pgautorole/internal/utils"
)

// 新規会員マネージャ
type NewbieManager interface {
	// 会員ロール変化時に新規会員ロールを操作するハンドラ
	MemberRoleUpdateHandler(s *discordgo.Session, m *discordgo.GuildMemberUpdate)
	// 新規会員ロールを更新
	RefreshNewbieRoles(s *discordgo.Session)
}

type newbieManager struct {
	// サーバーID
	guildID string
	// 新規会員ロールID
	newbieRoleID string
	// 会員ロールID
	memberRoleID string
	// 新規会員とみなす期間
	newbieDuration time.Duration
}

// 新規会員マネージャを作成
func NewNewbieManager(guildID, newbieRoleID, memberRoleID string, newbieDuration time.Duration) NewbieManager {
	return &newbieManager{
		guildID:        guildID,
		newbieRoleID:   newbieRoleID,
		memberRoleID:   memberRoleID,
		newbieDuration: newbieDuration,
	}
}

// userが新規会員に該当するかどうか
func (n *newbieManager) checkNewbie(member *discordgo.Member) (bool, error) {
	// 会員でない場合は新規会員ではない
	if !slices.Contains(member.Roles, n.memberRoleID) {
		return false, nil
	}
	// JoinedAtからNEWBIE_DURAITON経っていない新規会員
	return time.Since(member.JoinedAt) < n.newbieDuration, nil
}

func (n *newbieManager) MemberRoleUpdateHandler(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	// イベントが発生したサーバーが異なる場合は無視
	if m.GuildID != n.guildID {
		return
	}

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
	// 会員ロールが剥奪され、新規会員ロールが存在する場合
	if !slices.Contains(m.Roles, n.memberRoleID) && slices.Contains(m.Roles, n.newbieRoleID) {
		slog.Info("Remove newbie role", "m.User.ID", m.User.ID)
		s.GuildMemberRoleRemove(m.GuildID, m.User.ID, n.newbieRoleID)
	}
}

// 一度に取得するメンバー数
const MEMBERS_PER_REQUEST = 1000

func (n *newbieManager) RefreshNewbieRoles(s *discordgo.Session) {
	guildIsOnline := utils.SlicesHas(s.State.Guilds, func(g *discordgo.Guild) bool {
		return g.ID == n.guildID
	})
	if !guildIsOnline {
		return
	}

	after := ""
	for {
		// MEMBER_PER_REQUESTずつメンバーを取得
		m, err := s.GuildMembers(n.guildID, after, MEMBERS_PER_REQUEST)
		// メンバーが取得できなかった場合は終了
		if err != nil || len(m) == 0 {
			break
		}

		// 全メンバーに対して処理
		for _, member := range m {
			isNewbie, err := n.checkNewbie(member)
			if err == nil {
				if isNewbie {
					// 新規会員の場合は新規会員ロールを付与
					slog.Info("Add newbie role", "member.User.ID", member.User.ID)
					err := s.GuildMemberRoleAdd(n.guildID, member.User.ID, n.newbieRoleID)
					if err != nil {
						slog.Error("Failed to add newbie role", err)
					}
				} else if slices.Contains(member.Roles, n.newbieRoleID) {
					// 新規会員でない新規会員ロールがいた場合は新規会員ロールを削除
					slog.Info("Remove newbie role", "member.User.ID", member.User.ID)
					err := s.GuildMemberRoleRemove(n.guildID, member.User.ID, n.newbieRoleID)
					if err != nil {
						slog.Error("Failed to remove newbie role", err)
					}
				}
			}
		}

		after = m[len(m)-1].User.ID
	}
}
