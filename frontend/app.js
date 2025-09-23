// app.js
import {apiCall} from './api.js';

// =================================================================
// 1. 操作配置 (Action Configuration)
// 这是驱动动态表单的核心。我们为每个action定义它需要的字段。
// =================================================================
const ACTION_CONFIG = {
  // --- 用户管理 ---
  Register: {
    title: '用户注册 (管理员权限)',
    fields: [
      {
        path: 'userInfo.name',
        label: '用户名',
        type: 'text',
        placeholder: '例如: zhangsan'
      },
      {path: 'userInfo.password', label: '密码', type: 'password'}, {
        path: 'userInfo.identityInfo.class.grade',
        label: '年级',
        type: 'number',
        placeholder: '例如: 1'
      },
      {
        path: 'userInfo.identityInfo.class.class',
        label: '班级',
        type: 'number',
        placeholder: '例如: 2'
      },
      {
        path: 'userInfo.identityInfo.privilege',
        label: '权限',
        type: 'select',
        options: ['Student', 'Teacher', 'Admin']
      }
    ]
  },
  Remove: {
    title: '移除用户 (管理员权限)',
    fields: [{path: 'name', label: '要移除的用户名', type: 'text'}]
  },
  LogIn: {
    title: '用户登录',
    fields: [
      {path: 'name', label: '用户名', type: 'text'},
      {path: 'password', label: '密码', type: 'password'}
    ]
  },
  LogOut: {
    title: '用户登出',
    fields: []  // 无需参数
  },
  ModifyPassword: {
    title: '修改密码',
    fields: [{path: 'password', label: '新密码', type: 'password'}]
  },
  GetUserInfo: {
    title: '获取用户信息 (教师权限)',
    fields: [{path: 'name', label: '要查询的用户名', type: 'text'}]
  },
  GetAllUsersInfo: {title: '获取所有用户信息 (管理员权限)', fields: []},
  GetPartUsersInfo: {
    title: '获取部分用户信息 (教师权限)',
    fields: [
      {
        path: 'way',
        label: '查询方式',
        type: 'select',
        options: [{text: '按班级', value: 0}, {text: '按课程', value: 1}]
      },
      {path: 'class.grade', label: '年级 (按班级查询)', type: 'number'},
      {path: 'class.class', label: '班级 (按班级查询)', type: 'number'},
      {path: 'courseName', label: '课程名称 (按课程查询)', type: 'text'}
    ]
  },
  // --- 课程管理 ---
  AddCourse: {
    title: '添加课程 (管理员权限)',
    fields: [
      {path: 'courseInfo.name', label: '课程名称', type: 'text'},
      {path: 'courseInfo.teacherName', label: '教师姓名', type: 'text'},
      {path: 'courseInfo.maximum', label: '课程容量', type: 'number'}
    ]
  },
  ModifyCourse: {
    title: '修改课程 (管理员权限)',
    fields: [
      {path: 'courseName', label: '原课程名称', type: 'text'},
      {path: 'courseInfo.name', label: '新课程名称', type: 'text'},
      {path: 'courseInfo.teacherName', label: '新教师姓名', type: 'text'},
      {path: 'courseInfo.maximum', label: '新课程容量', type: 'number'}
    ]
  },
  LaunchCourse: {
    title: '发布课程 (管理员权限)',
    fields: [{path: 'courseName', label: '课程名称', type: 'text'}]
  },
  GetAllCoursesInfo: {title: '获取所有课程信息', fields: []},
  SelectCourse: {
    title: '选择课程 (学生权限)',
    fields: [{path: 'courseName', label: '要选择的课程名称', type: 'text'}]
  },
  DropCourse: {title: '退出课程 (学生权限)', fields: []}
};

// =================================================================
// 2. DOM 元素获取
// =================================================================
const actionSelect = document.getElementById('action-select');
const dynamicForm = document.getElementById('dynamic-form');
const submitButton = document.getElementById('submit-button');
const responseContent = document.getElementById('response-content');
const authStatus = document.getElementById('auth-status');

// =================================================================
// 3. 核心功能函数
// =================================================================

/**
 * 根据 Action 配置动态生成表单
 * @param {string} action - 当前选择的操作
 */
function renderFormForAction(action) {
  dynamicForm.innerHTML = '';  // 清空旧表单
  const config = ACTION_CONFIG[action];

  if (!config || !config.fields) return;

  if (config.fields.length === 0) {
    dynamicForm.innerHTML = '<p>此操作无需额外参数。</p>';
    return;
  }

  config.fields.forEach(field => {
    const group = document.createElement('div');
    group.className = 'form-group';

    const label = document.createElement('label');
    label.setAttribute('for', field.path);
    label.textContent = field.label;

    let input;
    if (field.type === 'select') {
      input = document.createElement('select');
      field.options.forEach(opt => {
        const option = document.createElement('option');
        if (typeof opt === 'object') {
          option.value = opt.value;
          option.textContent = opt.text;
        } else {
          option.value = opt;
          option.textContent = opt;
        }
        input.appendChild(option);
      });
    } else {
      input = document.createElement('input');
      input.type = field.type || 'text';
      input.placeholder = field.placeholder || '';
    }

    input.id = field.path;
    input.name = field.path;

    group.appendChild(label);
    group.appendChild(input);
    dynamicForm.appendChild(group);
  });
}

/**
 * 从当前表单隐式构造 JSON 报文
 * @returns {object} - 构造好的参数对象
 */
function buildParametersFromForm() {
  const params = {};
  const inputs = dynamicForm.querySelectorAll('input, select');

  inputs.forEach(input => {
    const path = input.name;
    let value = input.value;

    // 类型转换
    if (input.type === 'number') {
      value = value ? Number(value) : null;
    }

    // 使用辅助函数设置嵌套值
    setNestedValue(params, path, value);
  });

  return params;
}

/**
 * 辅助函数：根据路径字符串设置对象中的嵌套值
 * e.g., setNestedValue(obj, 'user.address.city', 'New York')
 * @param {object} obj - 目标对象
 * @param {string} path - 路径字符串
 * @param {*} value - 要设置的值
 */
function setNestedValue(obj, path, value) {
  const keys = path.split('.');
  let current = obj;
  for (let i = 0; i < keys.length - 1; i++) {
    const key = keys[i];
    if (!current[key] || typeof current[key] !== 'object') {
      current[key] = {};
    }
    current = current[key];
  }
  current[keys[keys.length - 1]] = value;
}

/**
 * 更新认证状态显示
 */
function updateAuthStatus() {
  const token = localStorage.getItem('authToken');
  if (token) {
    authStatus.textContent = `已登录 (Token: ${token.substring(0, 8)}...)`;
    authStatus.style.color = 'var(--success-color)';
  } else {
    authStatus.textContent = '未登录';
    authStatus.style.color = 'var(--secondary-color)';
  }
}

/**
 * 显示操作结果 (带有诊断日志的版本)
 * @param {string} type - 'success', 'error', or 'info'
 * @param {string} message - 要显示的主要信息
 * @param {object|null} data - 要额外展示的数据
 */
function displayResponse(type, message, data = null) {
  // BREADCRUMB A: 检查 displayResponse 函数是否被调用
  console.log(`[BREADCRUMB A] displayResponse called with type: "${type}"`);

  responseContent.className = `content ${type}`;
  let content = `<strong>${message}</strong>`;
  if (data && Object.keys(data).length > 0) {
    content += `\n\n--- 返回数据 ---\n${JSON.stringify(data, null, 2)}`;
  }
  responseContent.textContent = content;

  // BREADCRUMB B: 检查 displayResponse 函数是否成功执行完毕
  console.log('[BREADCRUMB B] displayResponse finished successfully.');
}

// =================================================================
// 4. 事件监听与初始化 (最终的、简化的版本)
// =================================================================

// 当选择的操作变化时，重新渲染表单
actionSelect.addEventListener('change', (e) => {
  console.log('[BREADCRUMB 1] Refresh');
  renderFormForAction(e.target.value);
});

// 【核心修改】我们只监听按钮的 "click" 事件，这是用户唯一的交互点。
submitButton.addEventListener('click', async (event) => {
  console.log('[BREADCRUMB 2] listen click.');
  // 作为双重保险，我们在这里阻止点击事件可能附带的任何默认行为。
  event.preventDefault();

  const action = actionSelect.value;
  const parameters = buildParametersFromForm();

  displayResponse('info', '正在执行操作...');
  console.log('[BREADCRUMB 3] 正在执行操作...');
  try {
    const result = await apiCall(action, parameters);
    displayResponse('success', '操作成功！', result);
    updateAuthStatus();
  } catch (error) {
    displayResponse('error', `操作失败: ${error.message}`);
  }
  console.log('[BREADCRUMB 3] 操作执行完毕...');
});


// --- 应用初始化 ---
function initialize() {
  // 填充操作选择下拉框
  Object.keys(ACTION_CONFIG).forEach((actionKey, index) => {
    const option = document.createElement('option');
    option.value = actionKey;
    option.textContent = `${index + 1}. ${ACTION_CONFIG[actionKey].title}`;
    actionSelect.appendChild(option);
  });

  // 初始化认证状态和第一个表单
  updateAuthStatus();
  renderFormForAction(actionSelect.value);
}

initialize();