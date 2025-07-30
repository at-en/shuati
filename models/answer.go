package models

import (
	"time"
)

// UserAnswer 用户答题记录
type UserAnswer struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"not null;index"`
	QuestionID uint      `json:"question_id" gorm:"not null;index"`
	UserAnswer string    `json:"user_answer" gorm:"not null"`
	IsCorrect  bool      `json:"is_correct" gorm:"not null"`
	Category   string    `json:"category" gorm:"not null;index"`
	AnsweredAt time.Time `json:"answered_at"`
}

// ExamRecord 考试记录
type ExamRecord struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"not null;index"`
	ExamType    string    `json:"exam_type" gorm:"not null"` // practice, mock_exam
	TotalCount  int       `json:"total_count" gorm:"not null"`
	CorrectCount int      `json:"correct_count" gorm:"not null"`
	Score       float64   `json:"score" gorm:"not null"`
	Duration    int       `json:"duration"` // 答题用时（秒）
	StartedAt   time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// AnswerRequest 答题请求
type AnswerRequest struct {
	QuestionID uint   `json:"question_id" binding:"required"`
	Answer     string `json:"answer" binding:"required"`
}

// AnswerResponse 答题响应
type AnswerResponse struct {
	IsCorrect   bool   `json:"is_correct"`
	CorrectAnswer string `json:"correct_answer"`
	Explanation string `json:"explanation,omitempty"`
}

// UserStats 用户统计
type UserStats struct {
	TotalAnswered int              `json:"total_answered"`
	CorrectCount  int              `json:"correct_count"`
	Accuracy      float64          `json:"accuracy"`
	CategoryStats []CategoryStats  `json:"category_stats"`
	WrongQuestions []WrongQuestion `json:"wrong_questions"`
}

// CategoryStats 分类统计
type CategoryStats struct {
	Category     string  `json:"category"`
	Total        int     `json:"total"`
	Correct      int     `json:"correct"`
	Accuracy     float64 `json:"accuracy"`
}

// WrongQuestion 错题信息
type WrongQuestion struct {
	QuestionID  uint   `json:"question_id"`
	Question    string `json:"question"`
	UserAnswer  string `json:"user_answer"`
	CorrectAnswer string `json:"correct_answer"`
	Category    string `json:"category"`
	AnsweredAt  time.Time `json:"answered_at"`
}

// ExamSession 考试会话
type ExamSession struct {
	ID          string    `json:"id"`
	UserID      uint      `json:"user_id"`
	Questions   []uint    `json:"questions"`   // 题目ID列表
	Answers     map[uint]string `json:"answers"` // 已答题目
	StartTime   time.Time `json:"start_time"`
	Duration    int       `json:"duration"`    // 考试时长（分钟）
	IsCompleted bool      `json:"is_completed"`
}