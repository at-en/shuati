package services

import (
	"fmt"
	"sync"
	"time"

	"quiz-system/models"
	lru "github.com/hashicorp/golang-lru/v2"
)

// CacheService 缓存服务
type CacheService struct {
	questionCache *lru.Cache[uint, *models.Question] // 题目缓存
	userSessions  sync.Map                          // 用户会话缓存
	examSessions  sync.Map                          // 考试会话缓存
	stats         *CacheStats
	mu            sync.RWMutex
}

// CacheStats 缓存统计
type CacheStats struct {
	QuestionHits   int64 `json:"question_hits"`
	QuestionMisses int64 `json:"question_misses"`
	SessionCount   int   `json:"session_count"`
	ExamCount      int   `json:"exam_count"`
}

var Cache *CacheService

// InitCache 初始化缓存服务
func InitCache() error {
	questionCache, err := lru.New[uint, *models.Question](500) // 缓存500道题目
	if err != nil {
		return fmt.Errorf("failed to create question cache: %v", err)
	}

	Cache = &CacheService{
		questionCache: questionCache,
		stats:         &CacheStats{},
	}

	return nil
}

// GetQuestion 从缓存获取题目
func (c *CacheService) GetQuestion(id uint) (*models.Question, error) {
	// 首先尝试从缓存获取
	if question, ok := c.questionCache.Get(id); ok {
		c.mu.Lock()
		c.stats.QuestionHits++
		c.mu.Unlock()
		return question, nil
	}

	// 缓存未命中，从数据库查询
	c.mu.Lock()
	c.stats.QuestionMisses++
	c.mu.Unlock()

	var question models.Question
	if err := DB.First(&question, id).Error; err != nil {
		return nil, err
	}

	// 添加到缓存
	c.questionCache.Add(id, &question)
	return &question, nil
}

// GetQuestionsByCategory 按分类获取题目
func (c *CacheService) GetQuestionsByCategory(category string, limit int) ([]models.Question, error) {
	var questions []models.Question
	query := DB.Where("category = ?", category)
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&questions).Error; err != nil {
		return nil, err
	}

	// 将查询到的题目添加到缓存
	for i := range questions {
		c.questionCache.Add(questions[i].ID, &questions[i])
	}

	return questions, nil
}

// GetRandomQuestions 获取随机题目
func (c *CacheService) GetRandomQuestions(qType string, count int) ([]models.Question, error) {
	var questions []models.Question
	
	query := DB.Order("RANDOM()").Limit(count)
	if qType != "" {
		query = query.Where("type = ?", qType)
	}
	
	if err := query.Find(&questions).Error; err != nil {
		return nil, err
	}

	// 将查询到的题目添加到缓存
	for i := range questions {
		c.questionCache.Add(questions[i].ID, &questions[i])
	}

	return questions, nil
}

// UserSession 用户会话
type UserSession struct {
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	LoginTime time.Time `json:"login_time"`
	LastSeen  time.Time `json:"last_seen"`
}

// SetUserSession 设置用户会话
func (c *CacheService) SetUserSession(sessionID string, session *UserSession) {
	c.userSessions.Store(sessionID, session)
	c.mu.Lock()
	c.stats.SessionCount++
	c.mu.Unlock()
}

// GetUserSession 获取用户会话
func (c *CacheService) GetUserSession(sessionID string) (*UserSession, bool) {
	if session, ok := c.userSessions.Load(sessionID); ok {
		userSession := session.(*UserSession)
		userSession.LastSeen = time.Now()
		c.userSessions.Store(sessionID, userSession)
		return userSession, true
	}
	return nil, false
}

// DeleteUserSession 删除用户会话
func (c *CacheService) DeleteUserSession(sessionID string) {
	c.userSessions.Delete(sessionID)
	c.mu.Lock()
	if c.stats.SessionCount > 0 {
		c.stats.SessionCount--
	}
	c.mu.Unlock()
}

// SetExamSession 设置考试会话
func (c *CacheService) SetExamSession(sessionID string, session *models.ExamSession) {
	c.examSessions.Store(sessionID, session)
	c.mu.Lock()
	c.stats.ExamCount++
	c.mu.Unlock()
}

// GetExamSession 获取考试会话
func (c *CacheService) GetExamSession(sessionID string) (*models.ExamSession, bool) {
	if session, ok := c.examSessions.Load(sessionID); ok {
		return session.(*models.ExamSession), true
	}
	return nil, false
}

// UpdateExamSession 更新考试会话
func (c *CacheService) UpdateExamSession(sessionID string, session *models.ExamSession) {
	c.examSessions.Store(sessionID, session)
}

// DeleteExamSession 删除考试会话
func (c *CacheService) DeleteExamSession(sessionID string) {
	c.examSessions.Delete(sessionID)
	c.mu.Lock()
	if c.stats.ExamCount > 0 {
		c.stats.ExamCount--
	}
	c.mu.Unlock()
}

// GetCacheStats 获取缓存统计
func (c *CacheService) GetCacheStats() *CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	// 计算当前会话数量
	sessionCount := 0
	c.userSessions.Range(func(key, value interface{}) bool {
		sessionCount++
		return true
	})
	
	examCount := 0
	c.examSessions.Range(func(key, value interface{}) bool {
		examCount++
		return true
	})
	
	return &CacheStats{
		QuestionHits:   c.stats.QuestionHits,
		QuestionMisses: c.stats.QuestionMisses,
		SessionCount:   sessionCount,
		ExamCount:      examCount,
	}
}

// CleanupExpiredSessions 清理过期会话
func (c *CacheService) CleanupExpiredSessions() {
	now := time.Now()
	expireDuration := 24 * time.Hour // 24小时过期
	
	// 清理用户会话
	c.userSessions.Range(func(key, value interface{}) bool {
		session := value.(*UserSession)
		if now.Sub(session.LastSeen) > expireDuration {
			c.userSessions.Delete(key)
		}
		return true
	})
	
	// 清理考试会话
	c.examSessions.Range(func(key, value interface{}) bool {
		session := value.(*models.ExamSession)
		if now.Sub(session.StartTime) > expireDuration {
			c.examSessions.Delete(key)
		}
		return true
	})
}

// StartCleanupTimer 启动清理定时器
func (c *CacheService) StartCleanupTimer() {
	ticker := time.NewTicker(1 * time.Hour) // 每小时清理一次
	go func() {
		for range ticker.C {
			c.CleanupExpiredSessions()
		}
	}()
}

// PreloadQuestions 预加载热点题目到缓存
func (c *CacheService) PreloadQuestions() error {
	// 预加载每个分类的前50道题目
	var categories []string
	if err := DB.Model(&models.Question{}).
		Distinct("category").
		Pluck("category", &categories).Error; err != nil {
		return err
	}
	
	for _, category := range categories {
		var questions []models.Question
		if err := DB.Where("category = ?", category).
			Limit(50).
			Find(&questions).Error; err != nil {
			continue
		}
		
		for i := range questions {
			c.questionCache.Add(questions[i].ID, &questions[i])
		}
	}
	
	return nil
}