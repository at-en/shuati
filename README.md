# Go语言高效刷题系统

基于现有的 `questions.json` 题库文件（包含4615道商用密码应用安全性评估考试题目），使用Go语言开发的高性能、零配置、单一可执行文件的刷题系统。

## 🚀 特性

- **零配置启动**: 双击或命令行直接运行，无需复杂配置
- **单一文件部署**: 编译后只需一个可执行文件
- **高性能缓存**: 内存占用<20MB，毫秒级响应
- **自动数据转换**: 启动时自动处理questions.json
- **现代化界面**: 响应式Web界面，支持移动设备
- **完整功能**: 分类练习、随机练习、模拟考试、错题本
- **Alpine优化**: 专门针对Alpine Linux环境优化

## 📋 系统要求

### 开发环境
- **操作系统**: Alpine Linux 3.15+
- **开发语言**: Go 1.21+
- **编译依赖**: gcc, musl-dev, sqlite-dev, git

### 运行环境
- **内存**: 20MB+
- **磁盘**: 50MB+
- **网络**: 端口50442

## 🛠️ 技术栈

- **Web框架**: Gin (高性能HTTP框架)
- **数据库**: SQLite3 + GORM (内嵌数据库)
- **缓存**: golang-lru + sync.Map (内存缓存)
- **前端**: 原生JavaScript + TailwindCSS
- **静态资源**: Go embed (内嵌到可执行文件)
- **密码加密**: golang.org/x/crypto/bcrypt

## 📦 安装和编译

### 1. 安装依赖

```bash
# Alpine Linux
apk add --no-cache go gcc musl-dev sqlite-dev git

# 其他Linux发行版
# Ubuntu/Debian: apt-get install golang gcc sqlite3 libsqlite3-dev git
# CentOS/RHEL: yum install golang gcc sqlite-devel git
```

### 2. 编译应用

```bash
# 使用提供的编译脚本
chmod +x build.sh
./build.sh

# 或手动编译
export CGO_ENABLED=1
export GOOS=linux
export GOARCH=amd64

go build -ldflags="-w -s -linkmode external -extldflags '-static'" \
    -tags="sqlite_omit_load_extension netgo osusergo static_build" \
    -o quiz-system main.go
```

### 3. 验证编译

```bash
# 检查文件
ls -lh quiz-system
file quiz-system

# 测试运行
./quiz-system --help || echo "Ready to run"
```

## 🚀 部署和运行

### 1. 准备文件

确保以下文件在同一目录：
- `quiz-system` (可执行文件)
- `questions.json` (题库文件)

### 2. 启动系统

```bash
# 直接运行
./quiz-system

# 后台运行
nohup ./quiz-system > quiz.log 2>&1 &

# 使用systemd (可选)
sudo systemctl start quiz-system
```

### 3. 访问系统

打开浏览器访问: http://localhost:50442

## 📚 使用指南

### 用户注册和登录

1. 首次访问需要注册账户
2. 用户名: 3-20个字符
3. 密码: 至少6个字符
4. 连续3次登录失败将锁定5分钟

### 功能模块

#### 1. 分类练习
- 按知识点分类学习
- 支持所有题目分类
- 立即显示答案和解析

#### 2. 随机练习  
- 随机抽取20道题目
- 强化训练模式
- 不限时间练习

#### 3. 模拟考试
- 严格按考试标准: 180题180分钟
- 单选60题 + 多选60题 + 判断60题
- 自动计时和提交
- 详细成绩分析

#### 4. 错题本
- 自动收集错题
- 按分类显示
- 支持重复练习

#### 5. 学习统计
- 总体学习进度
- 各分类正确率
- 详细数据分析

## 🏗️ 项目结构

```
quiz-system/
├── main.go              # 主程序入口
├── models/              # 数据模型
│   ├── question.go      # 题目模型
│   ├── user.go          # 用户模型
│   └── answer.go        # 答题记录模型
├── handlers/            # HTTP处理器
│   ├── question.go      # 题目相关API
│   ├── user.go          # 用户相关API
│   └── exam.go          # 考试相关API
├── services/            # 业务逻辑
│   ├── question.go      # 题目服务
│   ├── cache.go         # 缓存服务
│   └── database.go      # 数据库服务
├── static/              # 静态文件 (embed)
│   ├── index.html       # 单页面应用
│   └── app.js           # 前端逻辑
├── questions.json       # 题库文件 (4615道题目)
├── go.mod               # Go模块依赖
├── build.sh             # Alpine编译脚本
└── README.md            # 使用说明
```

## 🔧 配置说明

### 数据库配置

系统使用SQLite数据库，自动创建以下表：

```sql
-- 题目表
CREATE TABLE questions (
    id INTEGER PRIMARY KEY,
    type TEXT NOT NULL,           -- single, multiple, judge
    question TEXT NOT NULL,
    options TEXT,                 -- JSON格式存储选项
    answer TEXT NOT NULL,
    category TEXT NOT NULL,
    explanation TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 用户表
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    login_attempts INTEGER DEFAULT 0,
    locked_until DATETIME
);

-- 答题记录表
CREATE TABLE user_answers (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    question_id INTEGER NOT NULL,
    user_answer TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL,
    category TEXT NOT NULL,
    answered_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### 缓存配置

- LRU缓存容量: 500道题目
- 会话过期时间: 24小时
- 自动清理间隔: 1小时

### 服务器配置

- 默认端口: 50442
- 静态文件: 内嵌到可执行文件
- 日志级别: Release模式

## 🔍 API接口

### 认证接口

```bash
# 用户注册
POST /api/auth/register
Content-Type: application/json
{
    "username": "testuser",
    "password": "password123"
}

# 用户登录
POST /api/auth/login
Content-Type: application/json
{
    "username": "testuser", 
    "password": "password123"
}

# 用户登出
POST /api/auth/logout
```

### 题目接口

```bash
# 获取题目列表
GET /api/questions?category=分类&limit=20&type=single

# 获取单个题目
GET /api/questions/1

# 获取分类列表
GET /api/questions/categories

# 提交答案
POST /api/questions/submit
Content-Type: application/json
{
    "question_id": 1,
    "answer": "A"
}

# 获取错题本
GET /api/questions/wrong
```

### 考试接口

```bash
# 开始考试
POST /api/exam/start?type=mock_exam

# 提交考试答案
POST /api/exam/{sessionId}/answer
Content-Type: application/json
{
    "question_id": 1,
    "answer": "A"
}

# 完成考试
POST /api/exam/{sessionId}/complete

# 获取考试历史
GET /api/exam/history?limit=10
```

### 系统状态

```bash
# 健康检查
GET /api/health
```

## 🔧 故障排除

### 常见问题

1. **编译失败**
   ```bash
   # 检查Go版本
   go version
   
   # 安装缺失依赖
   apk add --no-cache gcc musl-dev sqlite-dev
   
   # 清理模块缓存
   go clean -modcache
   go mod download
   ```

2. **运行时错误**
   ```bash
   # 检查文件权限
   chmod +x quiz-system
   
   # 检查questions.json是否存在
   ls -la questions.json
   
   # 查看详细日志
   ./quiz-system 2>&1 | tee quiz.log
   ```

3. **端口占用**
   ```bash
   # 检查端口占用
   netstat -tlnp | grep 50442
   
   # 杀死占用端口的进程
   pkill -f quiz-system
   ```

4. **内存不足**
   ```bash
   # 检查内存使用
   free -h
   
   # 监控应用内存
   top -p $(pgrep quiz-system)
   ```

### 性能优化

1. **缓存命中率优化**
   - 预加载热点题目
   - 调整LRU缓存大小
   - 监控缓存统计

2. **数据库优化**
   - 定期清理过期数据
   - 优化查询索引
   - 使用批量操作

3. **内存优化**
   - 定期清理过期会话
   - 控制并发连接数
   - 监控内存使用

## 📊 监控和维护

### 系统监控

```bash
# 检查系统状态
curl http://localhost:50442/api/health

# 监控资源使用
htop
iostat 1

# 查看应用日志
tail -f quiz.log
```

### 数据备份

```bash
# 备份数据库
cp quiz.db quiz.db.backup.$(date +%Y%m%d)

# 备份题库
cp questions.json questions.json.backup
```

### 更新部署

```bash
# 停止服务
pkill quiz-system

# 备份当前版本
cp quiz-system quiz-system.old

# 部署新版本
cp quiz-system.new quiz-system
chmod +x quiz-system

# 启动服务
./quiz-system &
```

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 📞 支持

如有问题或建议，请通过以下方式联系：

- 提交 Issue
- 发送邮件
- 技术交流群

---

**享受学习，祝您考试顺利！** 🎉