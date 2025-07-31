package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"quiz-system/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase 初始化数据库
func InitDatabase() error {
	var err error
	
	// 配置数据库连接
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 生产环境关闭SQL日志
	}
	
	DB, err = gorm.Open(sqlite.Open("quiz.db"), config)
	if err != nil {
		return fmt.Errorf("failed to connect database: %v", err)
	}
	
	// 配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %v", err)
	}
	
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	
	// 自动迁移数据表
	err = DB.AutoMigrate(
		&models.Question{},
		&models.User{},
		&models.UserAnswer{},
		&models.ExamRecord{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}
	
	// 创建索引
	if err := createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %v", err)
	}
	
	// 检查是否需要导入题库数据
	var count int64
	DB.Model(&models.Question{}).Count(&count)
	if count == 0 {
		log.Println("Importing questions from questions.json...")
		if err := importQuestionsFromJSON(); err != nil {
			return fmt.Errorf("failed to import questions: %v", err)
		}
		// 重新计算导入后的题目数量
		DB.Model(&models.Question{}).Count(&count)
		log.Printf("Successfully imported %d questions", count)
	} else {
		log.Printf("Database already contains %d questions", count)
	}
	
	log.Println("Database initialized successfully")
	return nil
}

// createIndexes 创建数据库索引
func createIndexes() error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_questions_category ON questions(category)",
		"CREATE INDEX IF NOT EXISTS idx_questions_type ON questions(type)",
		"CREATE INDEX IF NOT EXISTS idx_user_answers_user_id ON user_answers(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_answers_category ON user_answers(category)",
		"CREATE INDEX IF NOT EXISTS idx_exam_records_user_id ON exam_records(user_id)",
	}
	
	for _, indexSQL := range indexes {
		if err := DB.Exec(indexSQL).Error; err != nil {
			return err
		}
	}
	
	return nil
}

// QuestionFile 题库文件结构
type QuestionFile struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Total       int    `json:"total_questions"`
	Questions   []QuestionData `json:"questions"`
}

// QuestionData 题目数据结构
type QuestionData struct {
	ID          int                    `json:"id"`
	Type        string                 `json:"type"`
	Question    string                 `json:"question"`
	Options     map[string]string      `json:"options,omitempty"`
	Answer      string                 `json:"answer"`
	Category    string                 `json:"category"`
	Explanation string                 `json:"explanation,omitempty"`
}

// importQuestionsFromJSON 从JSON文件导入题库数据
func importQuestionsFromJSON() error {
	// 检查文件是否存在
	if _, err := os.Stat("questions.json"); os.IsNotExist(err) {
		return fmt.Errorf("questions.json file not found")
	}
	
	// 读取JSON文件
	data, err := ioutil.ReadFile("questions.json")
	if err != nil {
		return fmt.Errorf("failed to read questions.json: %v", err)
	}
	
	log.Printf("Read questions.json file, size: %d bytes", len(data))
	
	// 解析JSON数据
	var questionFile QuestionFile
	if err := json.Unmarshal(data, &questionFile); err != nil {
		return fmt.Errorf("failed to parse questions.json: %v", err)
	}
	
	log.Printf("Parsed question file: %s, total questions: %d", questionFile.Title, len(questionFile.Questions))
	
	// 批量插入数据
	batchSize := 100
	questionsData := questionFile.Questions
	
	for i := 0; i < len(questionsData); i += batchSize {
		end := i + batchSize
		if end > len(questionsData) {
			end = len(questionsData)
		}
		
		batch := questionsData[i:end]
		questions := make([]models.Question, len(batch))
		
		for j, qd := range batch {
			// 转换选项为JSON字符串
			var optionsJSON string
			if len(qd.Options) > 0 {
				// 将map转换为数组格式，并按键排序
				var keys []string
				for key := range qd.Options {
					keys = append(keys, key)
				}
				sort.Strings(keys)
				
				var optionsList []string
				for _, key := range keys {
					optionsList = append(optionsList, fmt.Sprintf("%s. %s", key, qd.Options[key]))
				}
				optionsBytes, _ := json.Marshal(optionsList)
				optionsJSON = string(optionsBytes)
			}
			
			questions[j] = models.Question{
				Type:        normalizeQuestionType(qd.Type),
				Question:    strings.TrimSpace(qd.Question),
				Options:     optionsJSON,
				Answer:      strings.TrimSpace(qd.Answer),
				Category:    strings.TrimSpace(qd.Category),
				Explanation: strings.TrimSpace(qd.Explanation),
				CreatedAt:   time.Now(),
			}
		}
		
		if err := DB.Create(&questions).Error; err != nil {
			return fmt.Errorf("failed to insert questions batch: %v", err)
		}
	}
	
	return nil
}

// normalizeQuestionType 标准化题目类型
func normalizeQuestionType(qType string) string {
	qType = strings.ToLower(strings.TrimSpace(qType))
	switch qType {
	case "单选", "单选题", "single", "single_choice":
		return "single"
	case "多选", "多选题", "multiple", "multiple_choice":
		return "multiple"
	case "判断", "判断题", "judge", "true_false":
		return "judge"
	default:
		return "single" // 默认为单选题
	}
}

// GetQuestionStats 获取题目统计信息
func GetQuestionStats() (*models.QuestionStats, error) {
	var total int64
	if err := DB.Model(&models.Question{}).Count(&total).Error; err != nil {
		return nil, err
	}
	
	var categories []models.Category
	if err := DB.Model(&models.Question{}).
		Select("category as name, count(*) as count").
		Group("category").
		Find(&categories).Error; err != nil {
		return nil, err
	}
	
	return &models.QuestionStats{
		Total:      int(total),
		Categories: categories,
	}, nil
}

// GetUserStats 获取用户统计信息
func GetUserStats(userID uint) (*models.UserStats, error) {
	var totalAnswered int64
	var correctCount int64
	
	// 获取总答题数
	if err := DB.Model(&models.UserAnswer{}).
		Where("user_id = ?", userID).
		Count(&totalAnswered).Error; err != nil {
		return nil, err
	}
	
	// 获取正确答题数
	if err := DB.Model(&models.UserAnswer{}).
		Where("user_id = ? AND is_correct = ?", userID, true).
		Count(&correctCount).Error; err != nil {
		return nil, err
	}
	
	// 计算准确率
	var accuracy float64
	if totalAnswered > 0 {
		accuracy = float64(correctCount) / float64(totalAnswered) * 100
	}
	
	// 获取分类统计
	var categoryStats []models.CategoryStats
	if err := DB.Model(&models.UserAnswer{}).
		Select("category, count(*) as total, sum(case when is_correct then 1 else 0 end) as correct").
		Where("user_id = ?", userID).
		Group("category").
		Find(&categoryStats).Error; err != nil {
		return nil, err
	}
	
	// 计算每个分类的准确率
	for i := range categoryStats {
		if categoryStats[i].Total > 0 {
			categoryStats[i].Accuracy = float64(categoryStats[i].Correct) / float64(categoryStats[i].Total) * 100
		}
	}
	
	// 获取错题
	var wrongQuestions []models.WrongQuestion
	if err := DB.Table("user_answers ua").
		Select("ua.question_id, q.question, ua.user_answer, q.answer as correct_answer, ua.category, ua.answered_at").
		Joins("JOIN questions q ON ua.question_id = q.id").
		Where("ua.user_id = ? AND ua.is_correct = ?", userID, false).
		Order("ua.answered_at DESC").
		Limit(50). // 限制返回最近50道错题
		Find(&wrongQuestions).Error; err != nil {
		return nil, err
	}
	
	return &models.UserStats{
		TotalAnswered:  int(totalAnswered),
		CorrectCount:   int(correctCount),
		Accuracy:       accuracy,
		CategoryStats:  categoryStats,
		WrongQuestions: wrongQuestions,
	}, nil
}