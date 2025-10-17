# SSO Login Template

这是一个现代化的SSO（单点登录）界面模板，专为N Platform设计。

## 特性

### 🎨 现代化设计
- **渐变背景**：使用紫色渐变背景，与N Platform品牌一致
- **圆角设计**：24px圆角，现代化视觉效果
- **阴影效果**：柔和的阴影和悬停效果
- **响应式设计**：完美适配移动端和桌面端

### ✨ 交互体验
- **动画效果**：页面加载动画、按钮悬停效果
- **密码可见性切换**：点击眼睛图标切换密码显示/隐藏
- **密码强度检测**：实时显示密码强度指示器
- **表单验证**：前端验证和错误提示
- **加载状态**：提交时显示加载动画

### 🔧 功能特性
- **自动聚焦**：页面加载时自动聚焦用户名字段
- **键盘支持**：支持回车键提交表单
- **错误处理**：显示URL参数中的错误信息
- **防重复提交**：防止表单重复提交
- **深色模式支持**：自动适配系统深色模式

### 🛡️ 安全特性
- **CSRF保护**：使用隐藏字段传递return_url
- **输入验证**：前端和后端双重验证
- **安全提示**：密码强度实时检测
- **密码哈希**：前端SHA256哈希处理，提高传输安全性

## 文件结构

```
chaos_api/
├── templates/
│   ├── login.html          # 主登录页面模板
│   └── README.md           # 本文档
└── api/oauth/handler.go    # 后端处理逻辑
```

## 使用方法

### 1. 后端集成

在 `chaos_api/api/oauth/handler.go` 中，`LoginPage` 函数会自动加载模板：

```go
func LoginPage(c *gin.Context) {
    // 读取HTML模板文件
    tmpl, err := template.ParseFiles("templates/login.html")
    if err != nil {
        // 如果模板文件不存在，使用内联模板作为后备
        // ...
    }
    
    err = tmpl.Execute(c.Writer, map[string]any{
        "ReturnURL": c.Query("return_url"),
    })
    // ...
}
```

### 2. 路由配置

确保OAuth路由正确配置：

```go
// 在router.go中
oauthGroup.GET("/login", oauth.LoginPage)
oauthGroup.POST("/login", oauth.LoginPost)
```

### 3. 访问登录页面

访问登录页面时，可以通过URL参数传递return_url：

```
GET /oauth/login?return_url=https://example.com/callback
```

## 自定义配置

### 修改品牌信息

在 `login.html` 中修改以下部分：

```html
<div class="logo">N Platform</div>
<div class="subtitle">Single Sign-On Login</div>
```

### 修改颜色主题

修改CSS中的颜色变量：

```css
/* 主色调 */
background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);

/* 按钮颜色 */
background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
```

### 添加Logo

可以在logo区域添加图片：

```html
<div class="logo">
    <img src="/static/logo.png" alt="N Platform" style="height: 40px;">
    N Platform
</div>
```

## 浏览器兼容性

- ✅ Chrome 60+
- ✅ Firefox 55+
- ✅ Safari 12+
- ✅ Edge 79+
- ✅ 移动端浏览器

## 性能优化

- **CSS内联**：所有样式内联，减少HTTP请求
- **JavaScript优化**：最小化DOM操作
- **动画优化**：使用CSS3动画，GPU加速
- **响应式图片**：使用SVG背景，无额外加载

## 安全注意事项

1. **HTTPS**：生产环境必须使用HTTPS
2. **Cookie安全**：设置HttpOnly和Secure标志
3. **CSRF保护**：验证return_url的合法性
4. **输入验证**：前后端都要进行输入验证
5. **错误信息**：避免泄露敏感信息

## 故障排除

### 模板文件未找到

如果出现模板文件未找到的错误，检查：

1. 文件路径是否正确：`chaos_api/templates/login.html`
2. 工作目录是否正确
3. 文件权限是否正确

### 样式不显示

如果样式没有正确显示：

1. 检查CSS语法是否正确
2. 确认浏览器支持CSS3特性
3. 检查是否有其他CSS冲突

### JavaScript错误

如果JavaScript功能不正常：

1. 检查浏览器控制台错误
2. 确认DOM元素ID是否正确
3. 检查事件监听器是否正确绑定

## 更新日志

### v1.1.0 (当前版本)
- ✅ 前端SHA256密码哈希处理
- ✅ 提高密码传输安全性
- ✅ 防止密码明文传输

### v1.0.0
- ✅ 现代化设计界面
- ✅ 响应式布局
- ✅ 密码强度检测
- ✅ 深色模式支持
- ✅ 动画效果
- ✅ 表单验证
- ✅ 错误处理

## 贡献

欢迎提交Issue和Pull Request来改进这个登录界面！

## 许可证

本项目遵循MIT许可证。
