# 🐍 Python HTMX Directory

> **Purpose**: This directory contains Python functions that handle HTMX requests from your HTML templates.

## 🚀 How It Works

Each Python file defines functions that respond to HTMX requests:

- **File naming**: Use descriptive names like `user_auth.py`, `data_api.py`
- **Function naming**: Prefix functions with `htmx_` to mark them as HTMX handlers
- **Route mapping**: File structure determines API endpoints

## 📝 Python File Guidelines

### ✅ Do:
- **Use `htmx_` prefix**: `def htmx_login(request):`
- **Return HTML fragments**: Small pieces of HTML to swap into the page
- **Handle form data**: Process POST requests with form validation
- **Keep functions focused**: One responsibility per function

### ❌ Don't:
- **Return full pages**: HTMX works with fragments, not complete HTML
- **Use complex routing**: Let the file structure handle routing
- **Mix concerns**: Separate data logic from HTML generation

## 🔗 HTMX Integration Examples

### File: `auth.py`
```python
def htmx_login(request):
    # Handle login form submission
    return "<div class='success'>Logged in!</div>"

def htmx_logout(request):
    # Handle logout request
    return "<div class='info'>Logged out</div>"
```

### Corresponding HTML:
```html
<form hx-post="/auth/login" hx-target="#message">
    <input name="username" type="text">
    <button type="submit">Login</button>
</form>
<div id="message"></div>
```

## 📂 Example Structure

```
py_htmx/
├── auth.py         # Authentication handlers
├── user_data.py    # User management
└── api/
    └── posts.py    # Blog post handlers
```

## 🎯 Best Practices

- **Small responses**: Return minimal HTML for better performance
- **Error handling**: Always handle edge cases gracefully
- **Validation**: Validate all input data before processing
- **Security**: Sanitize user input and use CSRF protection

Happy coding! 🎉