package internal

import (
	"strings"

	"github.com/gw31415/pgautorole/internal/utils"
)

// コース内レベル
type Level string

const (
	// アプレンティス
	Apprentice Level = "アプレンティス"
	// アシスタント
	Assistant Level = "アシスタント"
	// ノーマル
	Normal Level = "ノーマル"
	// リード
	Lead Level = "リード"
)

var levels = []Level{
	Apprentice,
	Assistant,
	Normal,
	Lead,
}

// コース
type CourseName string

// コース配下のコースレベルを取得
func (c *CourseName) CourseLevelNames() []CourseLevelName {
	return utils.SlicesMap(levels, func(l Level) CourseLevelName {
		return CourseLevelName{
			Course: *c,
			Level:  l,
		}
	})
}

// 指定したレベルのコースレベルを取得
func (c *CourseName) With(level Level) CourseLevelName {
	return CourseLevelName{
		Course: *c,
		Level:  level,
	}
}

// コースとレベルの組み合わせ
type CourseLevelName struct {
	Course CourseName
	Level  Level
}

// コースレベルのロール名からコースとレベルを取得
func ParseCourseLevel(s string) *CourseLevelName {
	for _, l := range levels {
		if strings.HasSuffix(s, "-"+string(l)) {
			return &CourseLevelName{
				Course: CourseName(strings.TrimSuffix(s, "-"+string(l))),
				Level:  l,
			}
		}
	}
	return nil
}

// 対応するロール名を取得
func (cl *CourseLevelName) String() string {
	return string(cl.Course) + "-" + string(cl.Level)
}
