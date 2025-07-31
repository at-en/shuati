-- 检查数据库中题目和答案的存储格式
-- 可以使用SQLite命令行工具执行这些查询

-- 查看前5道题目的详细信息
SELECT id, type, question, options, answer, category 
FROM questions 
LIMIT 5;

-- 查看不同类型题目的答案格式
SELECT type, answer, COUNT(*) as count
FROM questions 
GROUP BY type, answer 
ORDER BY type, count DESC 
LIMIT 20;

-- 查看选项格式示例
SELECT id, type, options, answer
FROM questions 
WHERE options IS NOT NULL AND options != ''
LIMIT 10;