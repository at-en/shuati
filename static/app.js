// 全局状态管理
const AppState = {
    user: null,
    currentExamSession: null,
    currentQuestions: [],
    currentQuestionIndex: 0,
    examTimer: null,
    isExamMode: false
};

// API 基础URL
const API_BASE = '/api';

// HTML转义函数
function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// 页面初始化
document.addEventListener('DOMContentLoaded', function() {
    checkAuthStatus();
});

// 检查认证状态
async function checkAuthStatus() {
    try {
        const response = await fetch(`${API_BASE}/user/profile`, {
            credentials: 'include'
        });
        
        if (response.ok) {
            const data = await response.json();
            AppState.user = data.user;
            showMainPage();
        } else {
            showAuthPage();
        }
    } catch (error) {
        console.error('Auth check failed:', error);
        showAuthPage();
    }
}

// 显示认证页面
function showAuthPage() {
    document.getElementById('auth-page').style.display = 'block';
    document.getElementById('main-page').style.display = 'none';
    document.getElementById('nav-user').style.display = 'none';
}

// 显示主页面
function showMainPage() {
    document.getElementById('auth-page').style.display = 'none';
    document.getElementById('main-page').style.display = 'block';
    document.getElementById('nav-user').style.display = 'flex';
    document.getElementById('username').textContent = AppState.user.username;
    showWelcomeContent();
}

// 显示欢迎内容
function showWelcomeContent() {
    const contentArea = document.getElementById('content-area');
    contentArea.innerHTML = `
        <div id="welcome-content" class="text-center fade-in">
            <div class="bg-white rounded-lg shadow-md p-8">
                <h2 class="text-3xl font-bold text-gray-900 mb-4">欢迎来到刷题系统</h2>
                <p class="text-gray-600 mb-6">选择上方功能开始学习</p>
                <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mt-8">
                    <div class="bg-blue-50 p-4 rounded-lg">
                        <h3 class="font-semibold text-blue-800">分类练习</h3>
                        <p class="text-blue-600 text-sm">按知识点分类学习</p>
                    </div>
                    <div class="bg-green-50 p-4 rounded-lg">
                        <h3 class="font-semibold text-green-800">随机练习</h3>
                        <p class="text-green-600 text-sm">随机题目强化训练</p>
                    </div>
                    <div class="bg-yellow-50 p-4 rounded-lg">
                        <h3 class="font-semibold text-yellow-800">模拟考试</h3>
                        <p class="text-yellow-600 text-sm">180题180分钟考试</p>
                    </div>
                    <div class="bg-red-50 p-4 rounded-lg">
                        <h3 class="font-semibold text-red-800">错题复习</h3>
                        <p class="text-red-600 text-sm">针对性复习错题</p>
                    </div>
                </div>
            </div>
        </div>
    `;
}

// 切换到注册表单
function showRegister() {
    document.getElementById('login-form').style.display = 'none';
    document.getElementById('register-form').style.display = 'block';
}

// 切换到登录表单
function showLogin() {
    document.getElementById('register-form').style.display = 'none';
    document.getElementById('login-form').style.display = 'block';
}

// 用户注册
async function register(event) {
    event.preventDefault();
    
    const username = document.getElementById('register-username').value;
    const password = document.getElementById('register-password').value;
    
    try {
        showLoading(true);
        const response = await fetch(`${API_BASE}/auth/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, password }),
            credentials: 'include'
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showMessage('注册成功！请登录', 'success');
            showLogin();
        } else {
            showMessage(data.error || '注册失败', 'error');
        }
    } catch (error) {
        showMessage('网络错误，请稍后重试', 'error');
    } finally {
        showLoading(false);
    }
}

// 用户登录
async function login(event) {
    event.preventDefault();
    
    const username = document.getElementById('login-username').value;
    const password = document.getElementById('login-password').value;
    
    try {
        showLoading(true);
        const response = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, password }),
            credentials: 'include'
        });
        
        const data = await response.json();
        
        if (response.ok) {
            AppState.user = data.user;
            showMessage('登录成功！', 'success');
            showMainPage();
        } else {
            showMessage(data.error || '登录失败', 'error');
        }
    } catch (error) {
        showMessage('网络错误，请稍后重试', 'error');
    } finally {
        showLoading(false);
    }
}

// 用户登出
async function logout() {
    try {
        await fetch(`${API_BASE}/auth/logout`, {
            method: 'POST',
            credentials: 'include'
        });
        
        AppState.user = null;
        showMessage('已退出登录', 'info');
        showAuthPage();
    } catch (error) {
        console.error('Logout failed:', error);
    }
}

// 显示分类列表
async function showCategories() {
    try {
        showLoading(true);
        const response = await fetch(`${API_BASE}/questions/categories`, {
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error('Failed to fetch categories');
        }
        
        const data = await response.json();
        
        const contentArea = document.getElementById('content-area');
        contentArea.innerHTML = `
            <div class="fade-in">
                <div class="bg-white rounded-lg shadow-md p-6">
                    <h2 class="text-2xl font-bold text-gray-900 mb-6">选择分类</h2>
                    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        ${data.categories.map(category => `
                            <div class="border border-gray-200 rounded-lg p-4 hover:shadow-md transition duration-200 cursor-pointer"
                                 onclick="startCategoryPractice('${category.name}')">
                                <h3 class="font-semibold text-gray-900">${category.name}</h3>
                                <p class="text-gray-600 text-sm">${category.count} 道题目</p>
                            </div>
                        `).join('')}
                    </div>
                    <div class="mt-6 text-center">
                        <button onclick="showWelcomeContent()" class="text-gray-600 hover:text-gray-800">
                            ← 返回首页
                        </button>
                    </div>
                </div>
            </div>
        `;
    } catch (error) {
        showMessage('加载分类失败', 'error');
    } finally {
        showLoading(false);
    }
}

// 开始分类练习
async function startCategoryPractice(category) {
    try {
        showLoading(true);
        const response = await fetch(`${API_BASE}/questions?category=${encodeURIComponent(category)}&limit=20`, {
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error('Failed to fetch questions');
        }
        
        const data = await response.json();
        AppState.currentQuestions = data.questions;
        AppState.currentQuestionIndex = 0;
        AppState.isExamMode = false;
        
        showQuestionPage();
    } catch (error) {
        showMessage('加载题目失败', 'error');
    } finally {
        showLoading(false);
    }
}

// 显示随机练习
async function showPractice() {
    try {
        showLoading(true);
        const response = await fetch(`${API_BASE}/questions?limit=20`, {
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error('Failed to fetch questions');
        }
        
        const data = await response.json();
        AppState.currentQuestions = data.questions;
        AppState.currentQuestionIndex = 0;
        AppState.isExamMode = false;
        
        showQuestionPage();
    } catch (error) {
        showMessage('加载题目失败', 'error');
    } finally {
        showLoading(false);
    }
}

// 开始考试
async function startExam(examType) {
    if (!confirm(examType === 'mock_exam' ? 
        '即将开始模拟考试（180题，180分钟），确定开始吗？' : 
        '即将开始练习模式，确定开始吗？')) {
        return;
    }
    
    try {
        showLoading(true);
        const response = await fetch(`${API_BASE}/exam/start?type=${examType}`, {
            method: 'POST',
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error('Failed to start exam');
        }
        
        const data = await response.json();
        AppState.currentExamSession = data.session_id;
        AppState.currentQuestions = data.questions;
        AppState.currentQuestionIndex = 0;
        AppState.isExamMode = true;
        
        if (data.duration > 0) {
            startExamTimer(data.duration);
        }
        
        showQuestionPage();
    } catch (error) {
        showMessage('开始考试失败', 'error');
    } finally {
        showLoading(false);
    }
}

// 显示题目页面
function showQuestionPage() {
    if (AppState.currentQuestions.length === 0) {
        showMessage('没有题目可显示', 'error');
        return;
    }
    
    const question = AppState.currentQuestions[AppState.currentQuestionIndex];
    const options = question.options ? question.options.split('|') : [];
    
    const contentArea = document.getElementById('content-area');
    contentArea.innerHTML = `
        <div class="fade-in">
            <div class="bg-white rounded-lg shadow-md p-6">
                <!-- 进度条 -->
                <div class="mb-6">
                    <div class="flex justify-between items-center mb-2">
                        <span class="text-sm text-gray-600">
                            第 ${AppState.currentQuestionIndex + 1} 题 / 共 ${AppState.currentQuestions.length} 题
                        </span>
                        ${AppState.isExamMode && AppState.examTimer ? 
                            '<div id="timer" class="text-sm text-red-600 font-bold"></div>' : ''}
                    </div>
                    <div class="w-full bg-gray-200 rounded-full h-2">
                        <div class="bg-primary h-2 rounded-full" 
                             style="width: ${((AppState.currentQuestionIndex + 1) / AppState.currentQuestions.length) * 100}%"></div>
                    </div>
                </div>
                
                <!-- 题目内容 -->
                <div class="mb-6">
                    <div class="flex items-center mb-4">
                        <span class="bg-primary text-white px-3 py-1 rounded-full text-sm font-medium">
                            ${getQuestionTypeText(question.type)}
                        </span>
                        <span class="ml-3 text-gray-600 text-sm">${escapeHtml(question.category)}</span>
                    </div>
                    <h3 class="text-lg font-medium text-gray-900 mb-4">${escapeHtml(question.question)}</h3>
                    
                    <!-- 选项 -->
                    <div id="options-container">
                        ${renderQuestionOptions(question, options)}
                    </div>
                </div>
                
                <!-- 答案结果区域 -->
                <div id="answer-result" style="display: none;" class="mb-6 p-4 rounded-lg"></div>
                
                <!-- 操作按钮 -->
                <div class="flex justify-between">
                    <div>
                        ${AppState.currentQuestionIndex > 0 ? 
                            '<button onclick="previousQuestion()" class="bg-gray-500 text-white px-4 py-2 rounded-md hover:bg-gray-600">上一题</button>' : ''}
                    </div>
                    <div class="space-x-2">
                        ${!AppState.isExamMode ? 
                            '<button onclick="submitAnswer()" id="submit-btn" class="bg-primary text-white px-4 py-2 rounded-md hover:bg-blue-600">提交答案</button>' :
                            '<button onclick="submitExamAnswer()" id="submit-btn" class="bg-primary text-white px-4 py-2 rounded-md hover:bg-blue-600">提交答案</button>'}
                        ${AppState.currentQuestionIndex < AppState.currentQuestions.length - 1 ? 
                            '<button onclick="nextQuestion()" class="bg-gray-500 text-white px-4 py-2 rounded-md hover:bg-gray-600">下一题</button>' :
                            (AppState.isExamMode ? 
                                '<button onclick="completeExam()" class="bg-success text-white px-4 py-2 rounded-md hover:bg-green-600">完成考试</button>' :
                                '<button onclick="showWelcomeContent()" class="bg-success text-white px-4 py-2 rounded-md hover:bg-green-600">完成练习</button>')}
                    </div>
                </div>
            </div>
        </div>
    `;
}

// 渲染题目选项
function renderQuestionOptions(question, options) {
    try {
        if (question.type === 'judge') {
            return `
                <div class="space-y-3">
                    <label class="flex items-center space-x-3 cursor-pointer">
                        <input type="radio" name="answer" value="正确" class="form-radio text-primary">
                        <span>正确</span>
                    </label>
                    <label class="flex items-center space-x-3 cursor-pointer">
                        <input type="radio" name="answer" value="错误" class="form-radio text-primary">
                        <span>错误</span>
                    </label>
                </div>
            `;
        } else if (options && options.length > 0) {
            const inputType = question.type === 'multiple' ? 'checkbox' : 'radio';
            return `
                <div class="space-y-3">
                    ${options.map((option, index) => `
                        <label class="flex items-center space-x-3 cursor-pointer">
                            <input type="${inputType}" name="answer" value="${escapeHtml(option)}" class="form-${inputType} text-primary">
                            <span>${escapeHtml(option)}</span>
                        </label>
                    `).join('')}
                </div>
            `;
        } else {
            return `
                <textarea name="answer" rows="4" placeholder="请输入答案..." 
                          class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary"></textarea>
            `;
        }
    } catch (error) {
        console.error('Error rendering question options:', error);
        return '<p class="text-red-500">选项加载失败</p>';
    }
}

// 获取题目类型文本
function getQuestionTypeText(type) {
    switch (type) {
        case 'single': return '单选题';
        case 'multiple': return '多选题';
        case 'judge': return '判断题';
        default: return '题目';
    }
}

// 获取用户答案
function getUserAnswer() {
    const question = AppState.currentQuestions[AppState.currentQuestionIndex];
    
    if (question.type === 'multiple') {
        const checked = document.querySelectorAll('input[name="answer"]:checked');
        return Array.from(checked).map(input => input.value).join(',');
    } else {
        const checked = document.querySelector('input[name="answer"]:checked');
        if (checked) {
            return checked.value;
        }
        
        const textarea = document.querySelector('textarea[name="answer"]');
        if (textarea) {
            return textarea.value.trim();
        }
        
        return '';
    }
}

// 提交答案（练习模式）
async function submitAnswer() {
    const userAnswer = getUserAnswer();
    if (!userAnswer) {
        showMessage('请选择或输入答案', 'warning');
        return;
    }
    
    const question = AppState.currentQuestions[AppState.currentQuestionIndex];
    
    try {
        const response = await fetch(`${API_BASE}/questions/submit`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                question_id: question.id,
                answer: userAnswer
            }),
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error('Failed to submit answer');
        }
        
        const data = await response.json();
        showAnswerResult(data);
        
    } catch (error) {
        showMessage('提交答案失败', 'error');
    }
}

// 提交考试答案
async function submitExamAnswer() {
    const userAnswer = getUserAnswer();
    if (!userAnswer) {
        showMessage('请选择或输入答案', 'warning');
        return;
    }
    
    const question = AppState.currentQuestions[AppState.currentQuestionIndex];
    
    try {
        const response = await fetch(`${API_BASE}/exam/${AppState.currentExamSession}/answer`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                question_id: question.id,
                answer: userAnswer
            }),
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error('Failed to submit exam answer');
        }
        
        showMessage('答案已保存', 'success');
        nextQuestion();
        
    } catch (error) {
        showMessage('提交答案失败', 'error');
    }
}

// 显示答案结果
function showAnswerResult(result) {
    const resultDiv = document.getElementById('answer-result');
    const isCorrect = result.is_correct;
    
    resultDiv.className = `mb-6 p-4 rounded-lg ${isCorrect ? 'bg-green-50 border border-green-200' : 'bg-red-50 border border-red-200'}`;
    resultDiv.innerHTML = `
        <div class="flex items-center mb-2">
            <span class="font-bold ${isCorrect ? 'text-green-800' : 'text-red-800'}">
                ${isCorrect ? '✓ 回答正确' : '✗ 回答错误'}
            </span>
        </div>
        <p class="${isCorrect ? 'text-green-700' : 'text-red-700'}">
            <strong>正确答案：</strong>${result.correct_answer}
        </p>
        ${result.explanation ? `
            <p class="mt-2 ${isCorrect ? 'text-green-700' : 'text-red-700'}">
                <strong>解析：</strong>${result.explanation}
            </p>
        ` : ''}
    `;
    resultDiv.style.display = 'block';
    
    // 禁用提交按钮
    const submitBtn = document.getElementById('submit-btn');
    if (submitBtn) {
        submitBtn.disabled = true;
        submitBtn.className = 'bg-gray-400 text-white px-4 py-2 rounded-md cursor-not-allowed';
    }
}

// 上一题
function previousQuestion() {
    if (AppState.currentQuestionIndex > 0) {
        AppState.currentQuestionIndex--;
        showQuestionPage();
    }
}

// 下一题
function nextQuestion() {
    if (AppState.currentQuestionIndex < AppState.currentQuestions.length - 1) {
        AppState.currentQuestionIndex++;
        showQuestionPage();
    }
}

// 完成考试
async function completeExam() {
    if (!confirm('确定要完成考试吗？未答题目将被视为错误。')) {
        return;
    }
    
    try {
        showLoading(true);
        const response = await fetch(`${API_BASE}/exam/${AppState.currentExamSession}/complete`, {
            method: 'POST',
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error('Failed to complete exam');
        }
        
        const data = await response.json();
        showExamResult(data);
        
        // 清理考试状态
        AppState.currentExamSession = null;
        AppState.isExamMode = false;
        if (AppState.examTimer) {
            clearInterval(AppState.examTimer);
            AppState.examTimer = null;
        }
        
    } catch (error) {
        showMessage('完成考试失败', 'error');
    } finally {
        showLoading(false);
    }
}

// 显示考试结果
function showExamResult(result) {
    const contentArea = document.getElementById('content-area');
    const accuracy = result.total_questions > 0 ? (result.correct_answers / result.total_questions * 100).toFixed(1) : 0;
    const passed = result.score >= 60;
    
    contentArea.innerHTML = `
        <div class="fade-in">
            <div class="bg-white rounded-lg shadow-md p-8 text-center">
                <div class="mb-6">
                    <div class="${passed ? 'text-green-600' : 'text-red-600'} text-6xl mb-4">
                        ${passed ? '🎉' : '😔'}
                    </div>
                    <h2 class="text-3xl font-bold ${passed ? 'text-green-800' : 'text-red-800'} mb-2">
                        ${passed ? '恭喜通过！' : '继续努力！'}
                    </h2>
                    <p class="text-gray-600">考试已完成</p>
                </div>
                
                <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
                    <div class="bg-blue-50 p-4 rounded-lg">
                        <div class="text-2xl font-bold text-blue-800">${result.total_questions}</div>
                        <div class="text-blue-600 text-sm">总题数</div>
                    </div>
                    <div class="bg-green-50 p-4 rounded-lg">
                        <div class="text-2xl font-bold text-green-800">${result.correct_answers}</div>
                        <div class="text-green-600 text-sm">正确数</div>
                    </div>
                    <div class="bg-yellow-50 p-4 rounded-lg">
                        <div class="text-2xl font-bold text-yellow-800">${accuracy}%</div>
                        <div class="text-yellow-600 text-sm">正确率</div>
                    </div>
                    <div class="bg-purple-50 p-4 rounded-lg">
                        <div class="text-2xl font-bold text-purple-800">${result.duration}</div>
                        <div class="text-purple-600 text-sm">用时(分钟)</div>
                    </div>
                </div>
                
                <div class="space-x-4">
                    <button onclick="showWelcomeContent()" class="bg-primary text-white px-6 py-3 rounded-lg hover:bg-blue-600">
                        返回首页
                    </button>
                    <button onclick="showWrongQuestions()" class="bg-danger text-white px-6 py-3 rounded-lg hover:bg-red-600">
                        查看错题
                    </button>
                </div>
            </div>
        </div>
    `;
}

// 显示错题本
async function showWrongQuestions() {
    try {
        showLoading(true);
        const response = await fetch(`${API_BASE}/questions/wrong`, {
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error('Failed to fetch wrong questions');
        }
        
        const data = await response.json();
        
        const contentArea = document.getElementById('content-area');
        
        if (data.wrong_questions.length === 0) {
            contentArea.innerHTML = `
                <div class="fade-in">
                    <div class="bg-white rounded-lg shadow-md p-8 text-center">
                        <div class="text-green-600 text-6xl mb-4">🎉</div>
                        <h2 class="text-2xl font-bold text-gray-900 mb-2">太棒了！</h2>
                        <p class="text-gray-600 mb-6">您还没有错题，继续保持！</p>
                        <button onclick="showWelcomeContent()" class="bg-primary text-white px-6 py-3 rounded-lg hover:bg-blue-600">
                            返回首页
                        </button>
                    </div>
                </div>
            `;
            return;
        }
        
        contentArea.innerHTML = `
            <div class="fade-in">
                <div class="bg-white rounded-lg shadow-md p-6">
                    <h2 class="text-2xl font-bold text-gray-900 mb-6">错题本 (${data.total} 道题目)</h2>
                    <div class="space-y-4">
                        ${data.wrong_questions.map((item, index) => `
                            <div class="border border-gray-200 rounded-lg p-4">
                                <div class="flex justify-between items-start mb-2">
                                    <span class="bg-red-100 text-red-800 px-2 py-1 rounded text-sm">${item.category}</span>
                                    <span class="text-gray-500 text-sm">${new Date(item.answered_at).toLocaleDateString()}</span>
                                </div>
                                <h3 class="font-medium text-gray-900 mb-2">${item.question}</h3>
                                <div class="text-sm">
                                    <p class="text-red-600 mb-1"><strong>您的答案：</strong>${item.user_answer || '未作答'}</p>
                                    <p class="text-green-600"><strong>正确答案：</strong>${item.correct_answer}</p>
                                </div>
                            </div>
                        `).join('')}
                    </div>
                    <div class="mt-6 text-center">
                        <button onclick="showWelcomeContent()" class="text-gray-600 hover:text-gray-800">
                            ← 返回首页
                        </button>
                    </div>
                </div>
            </div>
        `;
    } catch (error) {
        showMessage('加载错题本失败', 'error');
    } finally {
        showLoading(false);
    }
}

// 显示学习统计
async function showStats() {
    try {
        showLoading(true);
        const response = await fetch(`${API_BASE}/user/stats`, {
            credentials: 'include'
        });
        
        if (!response.ok) {
            throw new Error('Failed to fetch stats');
        }
        
        const data = await response.json();
        
        const contentArea = document.getElementById('content-area');
        contentArea.innerHTML = `
            <div class="fade-in">
                <div class="bg-white rounded-lg shadow-md p-6">
                    <h2 class="text-2xl font-bold text-gray-900 mb-6">学习统计</h2>
                    
                    <!-- 总体统计 -->
                    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
                        <div class="bg-blue-50 p-4 rounded-lg text-center">
                            <div class="text-3xl font-bold text-blue-800">${data.total_answered || 0}</div>
                            <div class="text-blue-600">已答题数</div>
                        </div>
                        <div class="bg-green-50 p-4 rounded-lg text-center">
                            <div class="text-3xl font-bold text-green-800">${data.correct_count || 0}</div>
                            <div class="text-green-600">正确题数</div>
                        </div>
                        <div class="bg-yellow-50 p-4 rounded-lg text-center">
                            <div class="text-3xl font-bold text-yellow-800">${(data.accuracy || 0).toFixed(1)}%</div>
                            <div class="text-yellow-600">总体正确率</div>
                        </div>
                    </div>
                    
                    <!-- 分类统计 -->
                    <div class="mb-6">
                        <h3 class="text-lg font-semibold text-gray-900 mb-4">分类统计</h3>
                        <div class="space-y-3">
                            ${data.category_stats && data.category_stats.length > 0 ? data.category_stats.map(stat => `
                                <div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                                    <div>
                                        <span class="font-medium text-gray-900">${stat.category}</span>
                                        <span class="text-gray-600 text-sm ml-2">(${stat.correct}/${stat.total})</span>
                                    </div>
                                    <div class="flex items-center space-x-3">
                                        <div class="w-32 bg-gray-200 rounded-full h-2">
                                            <div class="bg-primary h-2 rounded-full" style="width: ${stat.accuracy || 0}%"></div>
                                        </div>
                                        <span class="text-sm font-medium text-gray-900 w-12">${(stat.accuracy || 0).toFixed(1)}%</span>
                                    </div>
                                </div>
                            `).join('') : '<p class="text-gray-500 text-center py-4">暂无统计数据</p>'}
                        </div>
                    </div>
                    
                    <div class="text-center">
                        <button onclick="showWelcomeContent()" class="text-gray-600 hover:text-gray-800">
                            ← 返回首页
                        </button>
                    </div>
                </div>
            </div>
        `;
    } catch (error) {
        showMessage('加载统计信息失败', 'error');
    } finally {
        showLoading(false);
    }
}

// 开始考试计时器
function startExamTimer(durationMinutes) {
    let remainingSeconds = durationMinutes * 60;
    
    AppState.examTimer = setInterval(() => {
        remainingSeconds--;
        
        const minutes = Math.floor(remainingSeconds / 60);
        const seconds = remainingSeconds % 60;
        
        const timerElement = document.getElementById('timer');
        if (timerElement) {
            timerElement.textContent = `剩余时间: ${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
        }
        
        if (remainingSeconds <= 0) {
            clearInterval(AppState.examTimer);
            AppState.examTimer = null;
            alert('考试时间到！系统将自动提交。');
            completeExam();
        }
    }, 1000);
}

// 显示消息提示
function showMessage(message, type = 'info') {
    const messageContainer = document.getElementById('message-container');
    const messageDiv = document.createElement('div');
    
    const colors = {
        success: 'bg-green-500',
        error: 'bg-red-500',
        warning: 'bg-yellow-500',
        info: 'bg-blue-500'
    };
    
    messageDiv.className = `${colors[type]} text-white px-4 py-2 rounded-lg shadow-lg mb-2 fade-in`;
    messageDiv.textContent = message;
    
    messageContainer.appendChild(messageDiv);
    
    setTimeout(() => {
        messageDiv.remove();
    }, 3000);
}

// 显示/隐藏加载状态
function showLoading(show) {
    const overlay = document.getElementById('loading-overlay');
    overlay.style.display = show ? 'flex' : 'none';
}