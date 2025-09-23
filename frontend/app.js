// app.js

// --- 第一部分：API通信核心 (与上一答复相同) ---

// 您的后端API的统一入口URL。
// 如果您的Go后端与前端部署在不同地址，请修改为完整URL，例如 'http://localhost:8080/api'
const API_URL = 'http://localhost:8080/api'; 

/**
 * 核心API调用函数
 * @param {string} action - 操作名称
 * @param {object} parameters - 参数对象
 * @returns {Promise<object>}
 */
async function apiCall(action, parameters) {
    const token = localStorage.getItem('authToken');
    const requestBody = {
        token: token,
        action: action,
        parameters: parameters,
        meta: {}
    };

    try {
        const response = await fetch(API_URL, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(requestBody),
        });

        if (!response.ok) {
            throw new Error(`网络错误: ${response.status} ${response.statusText}`);
        }

        const data = await response.json();

        // 您的API协议规定，业务错误通过errorMessage字段返回
        if (data.errorMessage) {
            throw new Error(data.errorMessage);
        }

        // 特殊处理登录成功的情况
        if (action === 'LogIn' && data.authToken) {
            localStorage.setItem('authToken', data.authToken);
        }
        
        // 特殊处理登出
        if (action === 'LogOut') {
            localStorage.removeItem('authToken');
        }

        return data;
    } catch (error) {
        // 将错误信息包装后抛出，以便UI层捕获
        console.error(`API调用[${action}]失败:`, error);
        throw error;
    }
}


// --- 第二部分：UI交互逻辑 ---

// 当整个页面加载完成后执行
document.addEventListener('DOMContentLoaded', () => {

    // 获取所有需要操作的HTML元素
    const actionSelector = document.getElementById('action-selector');
    const paramsInput = document.getElementById('params-input');
    const sendBtn = document.getElementById('send-request-btn');
    const responseOutput = document.getElementById('response-output');
    const tokenDisplay = document.getElementById('token-display');

    // 为方便测试，提供每个action的参数模板
    const actionTemplates = {
        Register: `{ "userInfo": { "name": "newuser", "password": "password123", "identityInfo": { "class": { "grade": 1, "class": 1 }, "privilege": 0 } } }`,
        Remove: `{ "name": "usertoremove" }`,
        LogIn: `{ "name": "testuser", "password": "password123" }`,
        LogOut: `null`,
        ModifyPassword: `{ "password": "newpassword" }`,
        GetUserInfo: `{ "name": "testuser" }`,
        GetAllUsersInfo: `null`,
        GetPartUsersInfo: `{ "way": 0, "class": { "grade": 1, "class": 1 } }`,
        AddCourse: `{ "courseInfo": { "name": "Math", "teacherName": "Mr. Smith", "maximum": 100 } }`,
        ModifyCourse: `{ "courseName": "Math", "courseInfo": { "name": "Advanced Math", "teacherName": "Dr. Jones", "maximum": 120 } }`,
        LaunchCourse: `{ "courseName": "Math" }`,
        GetAllCoursesInfo: `null`,
        SelectCourse: `{ "courseName": "Math" }`,
        DropCourse: `null`
    };

    // 更新Token显示区域的函数
    function updateTokenDisplay() {
        const token = localStorage.getItem('authToken');
        if (token) {
            // 为保护隐私，只显示部分token
            tokenDisplay.textContent = `${token.substring(0, 8)}...`;
            tokenDisplay.style.color = '#27ae60'; // 绿色表示已登录
        } else {
            tokenDisplay.textContent = '尚未登录';
            tokenDisplay.style.color = '#c0392b'; // 红色表示未登录
        }
    }

    // 当用户切换下拉菜单时，自动填充参数模板
    function populateTemplate() {
        const selectedAction = actionSelector.value;
        const template = actionTemplates[selectedAction];
        // 使用JSON.stringify美化格式
        paramsInput.value = JSON.stringify(JSON.parse(template), null, 4);
    }

    // "发送请求"按钮的点击事件处理
    sendBtn.addEventListener('click', async () => {
        const action = actionSelector.value;
        let parameters;

        // 1. 解析参数
        try {
            parameters = JSON.parse(paramsInput.value);
        } catch (error) {
            responseOutput.textContent = `参数JSON格式错误: ${error.message}`;
            responseOutput.style.color = 'red';
            return;
        }

        // 2. 发送API请求
        responseOutput.textContent = '请求发送中...';
        responseOutput.style.color = 'gray';

        try {
            const result = await apiCall(action, parameters);
            // 成功时，美化输出JSON结果
            responseOutput.textContent = JSON.stringify(result, null, 4);
            responseOutput.style.color = 'green';
        } catch (error) {
            // 失败时，显示错误信息
            responseOutput.textContent = `错误: ${error.message}`;
            responseOutput.style.color = 'red';
        }
        
        // 3. 无论成功失败，都更新一下Token的显示
        updateTokenDisplay();
    });

    // 监听下拉菜单的变换事件
    actionSelector.addEventListener('change', populateTemplate);

    // --- 初始化页面 ---
    updateTokenDisplay(); // 页面加载时立即更新一次Token显示
    populateTemplate(); // 并填充第一个选项的模板
});