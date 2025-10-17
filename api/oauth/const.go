package oauth

const fallbackTpl = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>N Platform - SSO Login</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
            position: relative;
            overflow-x: hidden;
        }

        /* 背景动画效果 */
        body::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><defs><pattern id="grain" width="100" height="100" patternUnits="userSpaceOnUse"><circle cx="25" cy="25" r="1" fill="white" opacity="0.1"/><circle cx="75" cy="75" r="1" fill="white" opacity="0.1"/><circle cx="50" cy="10" r="0.5" fill="white" opacity="0.1"/><circle cx="10" cy="60" r="0.5" fill="white" opacity="0.1"/><circle cx="90" cy="40" r="0.5" fill="white" opacity="0.1"/></pattern></defs><rect width="100" height="100" fill="url(%23grain)"/></svg>');
            pointer-events: none;
        }

        .login-container {
            background: white;
            border-radius: 24px;
            box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
            overflow: hidden;
            width: 100%;
            max-width: 420px;
            position: relative;
            animation: slideUp 0.6s ease-out;
        }

        @keyframes slideUp {
            from {
                opacity: 0;
                transform: translateY(30px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }

        .login-header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px 30px 30px;
            text-align: center;
            position: relative;
            overflow: hidden;
        }

        .login-header::before {
            content: '';
            position: absolute;
            top: -50%;
            left: -50%;
            width: 200%;
            height: 200%;
            background: radial-gradient(circle, rgba(255,255,255,0.1) 0%, transparent 70%);
            animation: float 6s ease-in-out infinite;
        }

        @keyframes float {
            0%, 100% { transform: translateY(0px) rotate(0deg); }
            50% { transform: translateY(-20px) rotate(180deg); }
        }

        .logo {
            font-size: 32px;
            font-weight: 800;
            margin-bottom: 8px;
            letter-spacing: -0.5px;
            position: relative;
            z-index: 1;
        }

        .subtitle {
            font-size: 16px;
            opacity: 0.9;
            font-weight: 400;
            position: relative;
            z-index: 1;
        }

        .login-form {
            padding: 40px 30px;
        }

        .form-group {
            margin-bottom: 24px;
            position: relative;
        }

        .form-label {
            display: block;
            font-size: 14px;
            font-weight: 600;
            color: #374151;
            margin-bottom: 8px;
            transition: color 0.2s ease;
        }

        .form-group:focus-within .form-label {
            color: #667eea;
        }

        .form-input {
            width: 100%;
            padding: 12px 16px;
            border: 2px solid #e5e7eb;
            border-radius: 12px;
            font-size: 16px;
            transition: all 0.2s ease;
            background: #f9fafb;
            position: relative;
        }

        .form-input:focus {
            outline: none;
            border-color: #667eea;
            background: white;
            box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
            transform: translateY(-1px);
        }

        .form-input::placeholder {
            color: #9ca3af;
        }

        .form-input.error {
            border-color: #dc2626;
            box-shadow: 0 0 0 3px rgba(220, 38, 38, 0.1);
        }

        .login-btn {
            width: 100%;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            padding: 14px 24px;
            border-radius: 12px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.2s ease;
            margin-top: 8px;
            position: relative;
            overflow: hidden;
        }

        .login-btn::before {
            content: '';
            position: absolute;
            top: 0;
            left: -100%;
            width: 100%;
            height: 100%;
            background: linear-gradient(90deg, transparent, rgba(255,255,255,0.2), transparent);
            transition: left 0.5s;
        }

        .login-btn:hover::before {
            left: 100%;
        }

        .login-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(102, 126, 234, 0.4);
        }

        .login-btn:active {
            transform: translateY(0);
        }

        .login-btn:disabled {
            opacity: 0.6;
            cursor: not-allowed;
            transform: none;
        }

        .error-message {
            background: #fef2f2;
            border: 1px solid #fecaca;
            color: #dc2626;
            padding: 12px 16px;
            border-radius: 8px;
            font-size: 14px;
            margin-bottom: 20px;
            display: none;
            animation: shake 0.5s ease-in-out;
        }

        @keyframes shake {
            0%, 100% { transform: translateX(0); }
            25% { transform: translateX(-5px); }
            75% { transform: translateX(5px); }
        }

        .error-message.show {
            display: block;
        }

        .footer {
            text-align: center;
            padding: 20px 30px;
            border-top: 1px solid #f3f4f6;
            color: #6b7280;
            font-size: 14px;
        }

        .loading {
            display: inline-block;
            width: 16px;
            height: 16px;
            border: 2px solid rgba(255, 255, 255, 0.3);
            border-radius: 50%;
            border-top-color: white;
            animation: spin 1s ease-in-out infinite;
            margin-right: 8px;
        }

        @keyframes spin {
            to { transform: rotate(360deg); }
        }

        .return-url {
            font-size: 12px;
            color: #9ca3af;
            margin-top: 16px;
            padding: 8px 12px;
            background: #f3f4f6;
            border-radius: 6px;
            word-break: break-all;
            border-left: 3px solid #667eea;
        }

        .password-toggle {
            position: absolute;
            right: 12px;
            top: 50%;
            transform: translateY(-50%);
            background: none;
            border: none;
            cursor: pointer;
            color: #9ca3af;
            font-size: 14px;
            padding: 4px;
            border-radius: 4px;
            transition: color 0.2s ease;
        }

        .password-toggle:hover {
            color: #667eea;
        }

        .password-field {
            position: relative;
        }

        .strength-meter {
            height: 4px;
            background: #e5e7eb;
            border-radius: 2px;
            margin-top: 8px;
            overflow: hidden;
        }

        .strength-fill {
            height: 100%;
            transition: all 0.3s ease;
            border-radius: 2px;
        }

        .strength-weak { background: #dc2626; width: 25%; }
        .strength-medium { background: #f59e0b; width: 50%; }
        .strength-strong { background: #10b981; width: 75%; }
        .strength-very-strong { background: #059669; width: 100%; }

        @media (max-width: 480px) {
            .login-container {
                margin: 10px;
                border-radius: 16px;
            }
            
            .login-header {
                padding: 30px 20px 20px;
            }
            
            .login-form {
                padding: 30px 20px;
            }
            
            .logo {
                font-size: 28px;
            }
        }

        /* 深色模式支持 */
        @media (prefers-color-scheme: dark) {
            .login-container {
                background: #1f2937;
                color: white;
            }
            
            .form-input {
                background: #374151;
                border-color: #4b5563;
                color: white;
            }
            
            .form-input:focus {
                background: #374151;
            }
            
            .form-label {
                color: #d1d5db;
            }
            
            .footer {
                border-top-color: #374151;
                color: #9ca3af;
            }
            
            .return-url {
                background: #374151;
                color: #9ca3af;
            }
        }
    </style>
</head>
<body>
    <div class="login-container">
        <div class="login-header">
            <div class="logo">N Platform</div>
            <div class="subtitle">Single Sign-On Login</div>
        </div>
        
        <form method="post" action="/oauth/login" class="login-form" id="loginForm">
            <input type="hidden" name="return_url" value="{{.ReturnURL}}">
            
            <div class="error-message" id="errorMessage"></div>
            
            <div class="form-group">
                <label for="username" class="form-label">Username</label>
                <input 
                    type="text" 
                    id="username" 
                    name="username" 
                    class="form-input" 
                    placeholder="Enter your username"
                    required
                    autocomplete="username"
                >
            </div>
            
            <div class="form-group">
                <label for="password" class="form-label">Password</label>
                <div class="password-field">
                    <input 
                        type="password" 
                        id="password" 
                        name="password" 
                        class="form-input" 
                        placeholder="Enter your password"
                        required
                        autocomplete="current-password"
                    >
                    <button type="button" class="password-toggle" id="passwordToggle" title="Toggle password visibility">
                        👁
                    </button>
                </div>
                <div class="strength-meter" id="strengthMeter" style="display: none;">
                    <div class="strength-fill" id="strengthFill"></div>
                </div>
            </div>
            
            <button type="submit" class="login-btn" id="loginBtn">
                <span class="btn-text">Sign In</span>
                <span class="btn-loading" style="display: none;">
                    <span class="loading"></span>
                    Signing in...
                </span>
            </button>
        </form>
        
        {{if .ReturnURL}}
        <div class="footer">
            <div class="return-url">
                <strong>Return to:</strong><br>
                {{.ReturnURL}}
            </div>
        </div>
        {{end}}
    </div>

    <script>
        // SHA256 哈希函数
        async function sha256(message) {
            const msgBuffer = new TextEncoder().encode(message);
            const hashBuffer = await crypto.subtle.digest('SHA-256', msgBuffer);
            const hashArray = Array.from(new Uint8Array(hashBuffer));
            const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
            return hashHex;
        }

        // 表单元素
        const loginForm = document.getElementById('loginForm');
        const usernameInput = document.getElementById('username');
        const passwordInput = document.getElementById('password');
        const passwordToggle = document.getElementById('passwordToggle');
        const loginBtn = document.getElementById('loginBtn');
        const btnText = loginBtn.querySelector('.btn-text');
        const btnLoading = loginBtn.querySelector('.btn-loading');
        const errorMessage = document.getElementById('errorMessage');
        const strengthMeter = document.getElementById('strengthMeter');
        const strengthFill = document.getElementById('strengthFill');

        // 密码可见性切换
        passwordToggle.addEventListener('click', function() {
            const type = passwordInput.type === 'password' ? 'text' : 'password';
            passwordInput.type = type;
            passwordToggle.textContent = type === 'password' ? '👁' : '🙈';
        });

        // 密码强度检测
        function checkPasswordStrength(password) {
            let score = 0;
            
            if (password.length >= 8) score++;
            if (/[a-z]/.test(password)) score++;
            if (/[A-Z]/.test(password)) score++;
            if (/[0-9]/.test(password)) score++;
            if (/[^A-Za-z0-9]/.test(password)) score++;
            
            return score;
        }

        passwordInput.addEventListener('input', function() {
            const password = this.value;
            const strength = checkPasswordStrength(password);
            
            if (password.length > 0) {
                strengthMeter.style.display = 'block';
                strengthFill.className = 'strength-fill';
                
                if (strength <= 2) {
                    strengthFill.classList.add('strength-weak');
                } else if (strength <= 3) {
                    strengthFill.classList.add('strength-medium');
                } else if (strength <= 4) {
                    strengthFill.classList.add('strength-strong');
                } else {
                    strengthFill.classList.add('strength-very-strong');
                }
            } else {
                strengthMeter.style.display = 'none';
            }
        });



        // 显示错误消息
        function showError(message) {
            errorMessage.textContent = message;
            errorMessage.classList.add('show');
        }

        // 输入框焦点处理
        [usernameInput, passwordInput].forEach(input => {
            input.addEventListener('focus', function() {
                this.classList.remove('error');
            });
        });

        // 自动聚焦用户名字段
        usernameInput.focus();

        // 处理回车键
        passwordInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                loginForm.submit();
            }
        });

        // 显示URL参数中的错误消息
        const urlParams = new URLSearchParams(window.location.search);
        const error = urlParams.get('error');
        if (error) {
            showError(decodeURIComponent(error));
        }

        // 添加输入动画效果
        [usernameInput, passwordInput].forEach(input => {
            input.addEventListener('input', function() {
                if (this.value.length > 0) {
                    this.style.transform = 'translateY(-1px)';
                } else {
                    this.style.transform = 'translateY(0)';
                }
            });
        });

        // 防止表单重复提交
        let isSubmitting = false;
        
        // 移除重复的事件监听器，统一处理
        loginForm.addEventListener('submit', async function(e) {
            e.preventDefault(); // 阻止默认提交
            
            // 防止重复提交
            if (isSubmitting) {
                return;
            }
            isSubmitting = true;
            
            // 清除之前的错误
            errorMessage.classList.remove('show');
            usernameInput.classList.remove('error');
            passwordInput.classList.remove('error');
            
            // 验证表单
            let hasError = false;
            
            if (!usernameInput.value.trim()) {
                usernameInput.classList.add('error');
                hasError = true;
            }
            
            if (!passwordInput.value.trim()) {
                passwordInput.classList.add('error');
                hasError = true;
            }
            
            if (hasError) {
                showError('Please fill in all required fields');
                isSubmitting = false;
                return;
            }
            
            // 显示加载状态
            loginBtn.disabled = true;
            btnText.style.display = 'none';
            btnLoading.style.display = 'inline-flex';
            
            try {
                // 对密码进行SHA256哈希
                const originalPassword = passwordInput.value;
                const hashedPassword = await sha256(originalPassword);
                
                // 直接替换原始密码字段的值为哈希值
                passwordInput.value = hashedPassword;
                
                // 使用传统表单提交，避免CORS问题
                console.log('使用传统表单提交，避免CORS问题');
                loginForm.submit();
                
            } catch (error) {
                console.error('Password hashing error:', error);
                showError('处理请求时发生错误');
                
                // 恢复按钮状态
                loginBtn.disabled = false;
                btnText.style.display = 'inline';
                btnLoading.style.display = 'none';
                isSubmitting = false;
            }
        });
    </script>
</body>
</html>
`
