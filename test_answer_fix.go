package main

import (
	"fmt"
	"strings"
)

// extractAnswerLetter 从完整答案中提取选项字母
func extractAnswerLetter(answer string) string {
	// 如果答案格式是 "A. 内容" 或 "A" 或 "A."，提取字母部分
	answer = strings.TrimSpace(answer)
	if len(answer) == 0 {
		return ""
	}
	
	// 检查是否是 "A. 内容" 格式
	if len(answer) >= 2 && answer[1] == '.' {
		return strings.ToUpper(string(answer[0]))
	}
	
	// 检查是否是单个字母
	if len(answer) == 1 && answer[0] >= 'A' && answer[0] <= 'Z' {
		return answer
	}
	if len(answer) == 1 && answer[0] >= 'a' && answer[0] <= 'z' {
		return strings.ToUpper(answer)
	}
	
	// 对于判断题
	if strings.EqualFold(answer, "正确") || strings.EqualFold(answer, "true") {
		return "正确"
	}
	if strings.EqualFold(answer, "错误") || strings.EqualFold(answer, "false") {
		return "错误"
	}
	
	// 如果都不匹配，返回原答案的第一个字符（转大写）
	return strings.ToUpper(string(answer[0]))
}

func main() {
	// 测试用例
	testCases := []struct {
		userAnswer    string
		correctAnswer string
		expected      bool
	}{
		{"A. 10Gbps", "A", true},
		{"B. 20Gbps", "A", false},
		{"A", "A", true},
		{"a", "A", true},
		{"正确", "正确", true},
		{"错误", "正确", false},
	}

	fmt.Println("测试答案提取和比较功能:")
	for i, tc := range testCases {
		userLetter := extractAnswerLetter(tc.userAnswer)
		isCorrect := strings.EqualFold(userLetter, tc.correctAnswer)
		
		status := "✓"
		if isCorrect != tc.expected {
			status = "✗"
		}
		
		fmt.Printf("%s 测试 %d: 用户答案='%s' -> 提取='%s', 正确答案='%s', 结果=%v (期望=%v)\n", 
			status, i+1, tc.userAnswer, userLetter, tc.correctAnswer, isCorrect, tc.expected)
	}
}