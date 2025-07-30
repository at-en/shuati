package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"quiz-system/models"
	"quiz-system/services"
	"github.com/gin-gonic/gin"
)

// GetQuestions 获取题目列表
func GetQuestions(c *gin.Context) {
	category := c.Query("category")
	limitStr := c.Query("limit")
	qType := c.Query("type")
	
	limit := 20 // 默认返回20道题
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	var questions []models.Question
	var err error

	if category != "" {
		// 按分类获取题目
		questions, err = services.Cache.GetQuestionsByCategory(category, limit)
	} else if qType != "" {
		// 按类型获取随机题目
		questions, err = services.Cache.GetRandomQuestions(qType, limit)
	} else {
		// 获取随机题目
		questions, err = services.Cache.GetRandomQuestions("", limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get questions",
		})
		return
	}

	// 处理选项字段
	for i := range questions {
		if questions[i].Options != "" {
			var options []string
			if err := json.Unmarshal([]byte(questions[i].Options), &options); err == nil {
				// 临时存储解析后的选项，用于前端显示
				questions[i].Options = strings.Join(options, "|")
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"questions": questions,
		"total": len(questions),
	})
}

// GetQuestion 获取单个题目
func GetQuestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid question ID",
		})
		return
	}

	question, err := services.Cache.GetQuestion(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Question not found",
		})
		return
	}

	// 处理选项字段
	if question.Options != "" {
		var options []string
		if err := json.Unmarshal([]byte(question.Options), &options); err == nil {
			question.Options = strings.Join(options, "|")
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"question": question,
	})
}

// GetCategories 获取所有分类
func GetCategories(c *gin.Context) {
	stats, err := services.GetQuestionStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get categories",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": stats.Categories,
		"total": stats.Total,
	})
}

// SubmitAnswer 提交答案
func SubmitAnswer(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	userSession := user.(*services.UserSession)

	var req models.AnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// 获取题目信息
	question, err := services.Cache.GetQuestion(req.QuestionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Question not found",
		})
		return
	}

	// 判断答案是否正确
	userAnswer := strings.TrimSpace(req.Answer)
	correctAnswer := strings.TrimSpace(question.Answer)
	isCorrect := false

	switch question.Type {
	case "single", "judge":
		isCorrect = strings.EqualFold(userAnswer, correctAnswer)
	case "multiple":
		// 多选题需要特殊处理
		isCorrect = compareMultipleChoiceAnswer(userAnswer, correctAnswer)
	}

	// 保存答题记录
	userAnswerRecord := models.UserAnswer{
		UserID:     userSession.UserID,
		QuestionID: req.QuestionID,
		UserAnswer: userAnswer,
		IsCorrect:  isCorrect,
		Category:   question.Category,
		AnsweredAt: time.Now(),
	}

	if err := services.DB.Create(&userAnswerRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save answer",
		})
		return
	}

	// 返回答题结果
	response := models.AnswerResponse{
		IsCorrect:     isCorrect,
		CorrectAnswer: correctAnswer,
		Explanation:   question.Explanation,
	}

	c.JSON(http.StatusOK, response)
}

// GetWrongQuestions 获取错题本
func GetWrongQuestions(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	userSession := user.(*services.UserSession)
	
	// 获取用户统计信息（包含错题）
	stats, err := services.GetUserStats(userSession.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get wrong questions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"wrong_questions": stats.WrongQuestions,
		"total": len(stats.WrongQuestions),
	})
}

// GetUserStats 获取用户统计信息
func GetUserStatsHandler(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	userSession := user.(*services.UserSession)
	
	stats, err := services.GetUserStats(userSession.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user stats",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SearchQuestions 搜索题目
func SearchQuestions(c *gin.Context) {
	keyword := c.Query("keyword")
	category := c.Query("category")
	qType := c.Query("type")
	limitStr := c.Query("limit")
	
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search keyword is required",
		})
		return
	}

	limit := 50 // 默认返回50道题
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// 构建查询
	query := services.DB.Where("question LIKE ?", "%"+keyword+"%")
	
	if category != "" {
		query = query.Where("category = ?", category)
	}
	
	if qType != "" {
		query = query.Where("type = ?", qType)
	}

	var questions []models.Question
	if err := query.Limit(limit).Find(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to search questions",
		})
		return
	}

	// 处理选项字段
	for i := range questions {
		if questions[i].Options != "" {
			var options []string
			if err := json.Unmarshal([]byte(questions[i].Options), &options); err == nil {
				questions[i].Options = strings.Join(options, "|")
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"questions": questions,
		"total": len(questions),
		"keyword": keyword,
	})
}

// compareMultipleChoiceAnswer 比较多选题答案
func compareMultipleChoiceAnswer(userAnswer, correctAnswer string) bool {
	// 将答案按逗号分割并排序比较
	userParts := strings.Split(userAnswer, ",")
	correctParts := strings.Split(correctAnswer, ",")
	
	// 去除空格并转换为小写
	for i := range userParts {
		userParts[i] = strings.ToLower(strings.TrimSpace(userParts[i]))
	}
	for i := range correctParts {
		correctParts[i] = strings.ToLower(strings.TrimSpace(correctParts[i]))
	}
	
	// 检查长度是否相同
	if len(userParts) != len(correctParts) {
		return false
	}
	
	// 检查每个选项是否都匹配
	userMap := make(map[string]bool)
	for _, part := range userParts {
		userMap[part] = true
	}
	
	for _, part := range correctParts {
		if !userMap[part] {
			return false
		}
	}
	
	return true
}