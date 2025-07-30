package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"quiz-system/models"
	"quiz-system/services"
	"github.com/gin-gonic/gin"
)

// StartExam 开始考试
func StartExam(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	userSession := user.(*services.UserSession)
	examType := c.Query("type") // practice 或 mock_exam
	
	if examType == "" {
		examType = "practice"
	}

	var questions []models.Question
	var duration int // 考试时长（分钟）
	
	if examType == "mock_exam" {
		// 模拟考试：单选60题、多选60题、判断60题，共180题，180分钟
		singleQuestions, err := services.Cache.GetRandomQuestions("single", 60)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get single choice questions",
			})
			return
		}
		
		multipleQuestions, err := services.Cache.GetRandomQuestions("multiple", 60)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get multiple choice questions",
			})
			return
		}
		
		judgeQuestions, err := services.Cache.GetRandomQuestions("judge", 60)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get judge questions",
			})
			return
		}
		
		questions = append(questions, singleQuestions...)
		questions = append(questions, multipleQuestions...)
		questions = append(questions, judgeQuestions...)
		duration = 180 // 180分钟
	} else {
		// 练习模式：随机20题，不限时间
		var err error
		questions, err = services.Cache.GetRandomQuestions("", 20)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get practice questions",
			})
			return
		}
		duration = 0 // 不限时间
	}

	// 生成考试会话ID
	sessionID, err := generateExamSessionID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate exam session",
		})
		return
	}

	// 提取题目ID列表
	questionIDs := make([]uint, len(questions))
	for i, q := range questions {
		questionIDs[i] = q.ID
	}

	// 创建考试会话
	examSession := &models.ExamSession{
		ID:          sessionID,
		UserID:      userSession.UserID,
		Questions:   questionIDs,
		Answers:     make(map[uint]string),
		StartTime:   time.Now(),
		Duration:    duration,
		IsCompleted: false,
	}

	services.Cache.SetExamSession(sessionID, examSession)

	// 创建考试记录
	examRecord := models.ExamRecord{
		UserID:       userSession.UserID,
		ExamType:     examType,
		TotalCount:   len(questions),
		CorrectCount: 0,
		Score:        0,
		StartedAt:    time.Now(),
	}

	if err := services.DB.Create(&examRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create exam record",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"exam_type":  examType,
		"questions":  questions,
		"duration":   duration,
		"start_time": examSession.StartTime,
	})
}

// GetExamSession 获取考试会话
func GetExamSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	
	examSession, exists := services.Cache.GetExamSession(sessionID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Exam session not found",
		})
		return
	}

	// 检查会话是否属于当前用户
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	userSession := user.(*services.UserSession)
	if examSession.UserID != userSession.UserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// 获取题目详情
	var questions []models.Question
	for _, qid := range examSession.Questions {
		question, err := services.Cache.GetQuestion(qid)
		if err != nil {
			continue
		}
		questions = append(questions, *question)
	}

	// 计算剩余时间
	var remainingTime int
	if examSession.Duration > 0 {
		elapsed := int(time.Since(examSession.StartTime).Minutes())
		remainingTime = examSession.Duration - elapsed
		if remainingTime < 0 {
			remainingTime = 0
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"session":        examSession,
		"questions":      questions,
		"remaining_time": remainingTime,
	})
}

// SubmitExamAnswer 提交考试答案
func SubmitExamAnswer(c *gin.Context) {
	sessionID := c.Param("sessionId")
	
	var req models.AnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	examSession, exists := services.Cache.GetExamSession(sessionID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Exam session not found",
		})
		return
	}

	// 检查会话是否属于当前用户
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	userSession := user.(*services.UserSession)
	if examSession.UserID != userSession.UserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// 检查考试是否已完成
	if examSession.IsCompleted {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Exam already completed",
		})
		return
	}

	// 检查时间是否超时
	if examSession.Duration > 0 {
		elapsed := int(time.Since(examSession.StartTime).Minutes())
		if elapsed >= examSession.Duration {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Exam time expired",
			})
			return
		}
	}

	// 保存答案到会话
	examSession.Answers[req.QuestionID] = req.Answer
	services.Cache.UpdateExamSession(sessionID, examSession)

	c.JSON(http.StatusOK, gin.H{
		"message": "Answer submitted successfully",
	})
}

// CompleteExam 完成考试
func CompleteExam(c *gin.Context) {
	sessionID := c.Param("sessionId")
	
	examSession, exists := services.Cache.GetExamSession(sessionID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Exam session not found",
		})
		return
	}

	// 检查会话是否属于当前用户
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	userSession := user.(*services.UserSession)
	if examSession.UserID != userSession.UserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// 标记考试完成
	examSession.IsCompleted = true
	services.Cache.UpdateExamSession(sessionID, examSession)

	// 计算成绩
	correctCount := 0
	totalCount := len(examSession.Questions)
	
	// 保存所有答题记录
	for _, questionID := range examSession.Questions {
		userAnswer, hasAnswer := examSession.Answers[questionID]
		if !hasAnswer {
			userAnswer = "" // 未答题
		}

		// 获取题目信息
		question, err := services.Cache.GetQuestion(questionID)
		if err != nil {
			continue
		}

		// 判断答案是否正确
		isCorrect := false
		if hasAnswer {
			switch question.Type {
			case "single", "judge":
				isCorrect = userAnswer == question.Answer
			case "multiple":
				isCorrect = compareMultipleChoiceAnswer(userAnswer, question.Answer)
			}
		}

		if isCorrect {
			correctCount++
		}

		// 保存答题记录
		userAnswerRecord := models.UserAnswer{
			UserID:     userSession.UserID,
			QuestionID: questionID,
			UserAnswer: userAnswer,
			IsCorrect:  isCorrect,
			Category:   question.Category,
			AnsweredAt: time.Now(),
		}
		services.DB.Create(&userAnswerRecord)
	}

	// 计算分数
	score := float64(correctCount) / float64(totalCount) * 100

	// 更新考试记录
	var examRecord models.ExamRecord
	if err := services.DB.Where("user_id = ? AND started_at >= ?", 
		userSession.UserID, examSession.StartTime.Add(-time.Minute)).
		Order("started_at DESC").First(&examRecord).Error; err == nil {
		
		examRecord.CorrectCount = correctCount
		examRecord.Score = score
		examRecord.Duration = int(time.Since(examSession.StartTime).Seconds())
		now := time.Now()
		examRecord.CompletedAt = &now
		services.DB.Save(&examRecord)
	}

	// 生成详细结果
	result := gin.H{
		"total_questions": totalCount,
		"correct_answers": correctCount,
		"score":          score,
		"duration":       int(time.Since(examSession.StartTime).Minutes()),
		"completed_at":   time.Now(),
	}

	// 清理考试会话
	services.Cache.DeleteExamSession(sessionID)

	c.JSON(http.StatusOK, result)
}

// GetExamHistory 获取考试历史
func GetExamHistory(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	userSession := user.(*services.UserSession)
	
	limitStr := c.Query("limit")
	limit := 10 // 默认返回10条记录
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	var examRecords []models.ExamRecord
	if err := services.DB.Where("user_id = ?", userSession.UserID).
		Order("started_at DESC").
		Limit(limit).
		Find(&examRecords).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get exam history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"exam_records": examRecords,
		"total":        len(examRecords),
	})
}

// generateExamSessionID 生成考试会话ID
func generateExamSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "exam_" + hex.EncodeToString(bytes), nil
}