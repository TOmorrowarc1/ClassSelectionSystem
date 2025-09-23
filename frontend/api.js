// api.js

const API_URL = 'http://localhost:8080/api';

/**
 * 核心API调用函数
 * @param {string} action - 要执行的操作名称
 * @param {object} parameters - 构造好的参数对象
 * @returns {Promise<object>} - 返回一个Promise，成功时解析为后端返回的数据
 * @throws {Error} - 当网络请求或业务逻辑失败时抛出错误
 */
export async function apiCall(action, parameters) {
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
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(requestBody),
        });

        const data = await response.json();

        if (!response.ok) {
            // 对于HTTP层面的错误 (如 500 Internal Server Error), 优先使用后端返回的错误信息
            throw new Error(data.errorMessage || `网络错误: ${response.status}`);
        }

        if (data.errorMessage) {
            throw new Error(data.errorMessage);
        }

        // 登录成功，存储token
        if (action === 'LogIn' && data.authToken) {
            localStorage.setItem('authToken', data.authToken);
        }

        // 登出成功，移除token
        if (action === 'LogOut') {
            localStorage.removeItem('authToken');
        }

        return data;

    } catch (error) {
        console.error(`API调用 [${action}] 失败:`, error);
        throw error; // 将错误继续向上抛出，以便UI层捕获
    }
}