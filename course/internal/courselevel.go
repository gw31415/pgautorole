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
type Course string

// コース配下のコースレベルを取得
func (c *Course) CourseLevels() []CourseLevel {
	return utils.SlicesMap(levels, func(l Level) CourseLevel {
		return CourseLevel{
			Course: *c,
			Level:  l,
		}
	})
}

// 指定したレベルのコースレベルを取得
func (c *Course) With(level Level) CourseLevel {
	return CourseLevel{
		Course: *c,
		Level:  level,
	}
}

// コースとレベルの組み合わせ
type CourseLevel struct {
	Course Course
	Level  Level
}

// コースレベルのロール名からコースとレベルを取得
func ParseCourseLevel(s string) *CourseLevel {
	for _, l := range levels {
		if strings.HasSuffix(s, "-"+string(l)) {
			return &CourseLevel{
				Course: Course(strings.TrimSuffix(s, "-"+string(l))),
				Level:  l,
			}
		}
	}
	return nil
}

// 対応するロール名を取得
func (cl *CourseLevel) String() string {
	return string(cl.Course) + "-" + string(cl.Level)
}
