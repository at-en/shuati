package models

import (
	"time"
)

// Question 题目模型
type Question struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Type        string    `json:"type" gorm:"not null"`        // single, multiple, judge
	Question    string    `json:"question" gorm:"not null"`
	Options     string    `json:"options,omitempty"`           // JSON格式存储选项
	Answer      string    `json:"answer" gorm:"not null"`
	Category    string    `json:"category" gorm:"not null"`
	Explanation string    `json:"explanation,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// QuestionJSON 用于解析questions.json的结构
type QuestionJSON struct {
	Type        string   `json:"type"`
	Question    string   `json:"question"`
	Options     []string `json:"options,omitempty"`
	Answer      string   `json:"answer"`
	Category    string   `json:"category"`
	Explanation string   `json:"explanation,omitempty"`
}

// Category 分类统计
type Category struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// QuestionStats 题目统计
type QuestionStats struct {
	Total      int        `json:"total"`
	Categories []Category `json:"categories"`
}