package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gw31415/pgautorole/newbie"
	"github.com/robfig/cron/v3"
)

var (
	// デバッグモード
	DEBUG_MODE = os.Getenv("DEBUG_MODE")

	// Discordのトークン
	DISCORD_TOKEN = os.Getenv("DISCORD_TOKEN")

	// DiscordサーバーID
	GUILD_ID = os.Getenv("GUILD_ID")
	// 一般会員のロールID
	MEMBER_ROLE_ID = os.Getenv("MEMBER_ROLE_ID")
	// 新規会員のロールID
	NEWBIE_ROLE_ID = os.Getenv("NEWBIE_ROLE_ID")
	// 新規会員のロールをリフレッシュするスケジュール
	NEWBIE_REFRESHING_CRON = os.Getenv("NEWBIE_REFRESHING_CRON")
	// 新規会員のロールの有効期間
	NEWBIE_MAX_DURATION, _ = time.ParseDuration(os.Getenv("NEWBIE_MAX_DURATION"))
)

func main() {
	if len(os.Getenv("DEBUG_MODE")) > 0 {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Debug mode")
	}

	// 環境変数のチェック
	if DISCORD_TOKEN == "" || MEMBER_ROLE_ID == "" || NEWBIE_ROLE_ID == "" || NEWBIE_REFRESHING_CRON == "" || NEWBIE_MAX_DURATION == 0 {
		slog.Error("Please set environment variables")
		return
	}

	// Discordセッションの初期化
	discord, err := discordgo.New("Bot " + DISCORD_TOKEN)
	if err != nil {
		slog.Error("Error creating Discord session:", err)
		return
	}
	discord.Identify.Intents = discordgo.IntentsGuildMembers

	// cronの初期化
	cr := cron.New()

	// 対応外のサーバーから退出する設定
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.GuildCreate) {
		if m.Guild.ID != GUILD_ID {
			slog.Warn("Leaving guild", "GUILD_ID", m.Guild.ID)
			s.GuildLeave(m.Guild.ID)
		}
	})
	discord.AddHandler(func(s *discordgo.Session, m *discordgo.Ready) {
		for _, guild := range m.Guilds {
			if guild.ID != GUILD_ID {
				slog.Warn("Leaving guild", "GUILD_ID", guild.ID)
				s.GuildLeave(guild.ID)
			}
		}
	})

	// NewbieManagerの設定
	slog.Info("Setting up NewbieManager", "MEMBER_ROLE_ID", MEMBER_ROLE_ID, "NEWBIE_ROLE_ID", NEWBIE_ROLE_ID, "NEWBIE_MAX_DURATION", NEWBIE_MAX_DURATION)
	newbiemanager := newbie.NewNewbieManager(GUILD_ID, NEWBIE_ROLE_ID, MEMBER_ROLE_ID, NEWBIE_MAX_DURATION)
	discord.AddHandler(newbiemanager.MemberRoleUpdateHandler)
	_, err = cr.AddFunc(NEWBIE_REFRESHING_CRON, func() {
		slog.Info("Refreshing newbie roles")
		newbiemanager.RefreshNewbieRoles(discord)
	})
	if err != nil {
		slog.Error("Error adding cron job:", err)
		return
	}

	// Discordセッションの開始
	slog.Info("Opening discord connection")
	err = discord.Open()
	if err != nil {
		slog.Error("Error opening discord connection:", err)
		return
	}
	defer discord.Close()

	// cronの開始
	slog.Info("Starting cron")
	go cr.Run()
	defer cr.Stop()

	// 終了シグナルの待機
	slog.Info("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	slog.Info("Shutting down...")
}
