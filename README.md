# requirements.txt
```
fastapi>=0.68.0
uvicorn[standard]>=0.15.0
aiohttp>=3.8.0
loguru>=0.6.0
psutil>=5.8.0
python-multipart>=0.0.5
async-property>=0.2.1
```

# HTMLnoJS

A Go-powered HTMX framework for building modern web applications without complex JavaScript.

## Features

- 🚀 Zero JavaScript required
- 🔧 Go-powered backend with Python handlers
- ⚡ HTMX for seamless interactivity
- 📁 File-based routing system
- 🎨 Automatic CSS injection
- 🔄 Hot reloading during development
- 🐍 Python functions as API endpoints

## Quick Start

### Installation

```bash
pip install htmlnojs
```

### Basic Usage

```python
from htmlnojs import htmlnojs

# Create and run your app
app = htmlnojs("./my-project")
app.run_forever()
```

### Project Structure

```
my-project/
├── templates/           # HTML files
│   ├── index.html      # → serves at /
│   └── about.html      # → serves at /about
├── css/                # CSS files  
│   ├── main.css        # Global styles
│   └── components.css  # Component styles
└── py_htmx/           # Python HTMX handlers
    ├── demo.py        # API handlers
    └── utils.py       # Utility functions
```

## How It Works

HTMLnoJS combines the speed of Go with the simplicity of Python:

1. **Go Server**: Serves HTML, CSS, and handles routing at blazing speed
2. **Python Handlers**: Process HTMX requests via FastAPI
3. **Automatic Routing**: Files become routes automatically
4. **Zero Config**: Just create files and they work

## HTML Templates

Create HTML files in the `templates/` directory:

```html
<!-- templates/index.html -->
<!DOCTYPE html>
<html>
<head>
    <title>My HTMLnoJS App</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body>
    <div class="container">
        <h1>Welcome to HTMLnoJS</h1>
        
        <!-- Interactive button -->
        <button hx-get="/api/demo/hello" hx-target="#result" class="btn">
            Click me for HTMX magic!
        </button>
        <div id="result"></div>
        
        <!-- Form submission -->
        <form hx-post="/api/demo/form" hx-target="#form-result">
            <input type="text" name="message" placeholder="Enter a message" required>
            <button type="submit">Submit</button>
        </form>
        <div id="form-result"></div>
    </div>
</body>
</html>
```

## Python Handlers

Create Python functions in the `py_htmx/` directory:

```python
# py_htmx/demo.py

def htmx_hello(request):
    """Simple HTMX handler"""
    return '<div class="alert">Hello from HTMLnoJS! 🎉</div>'

def htmx_form(request):
    """Handle form submission"""
    message = request.get('message', 'No message provided')
    return f'''
    <div class="alert success">
        <strong>Form submitted!</strong><br>
        Your message: "{message}"
    </div>
    '''

def htmx_get_users(request):
    """GET request - fetch users"""
    users = ['Alice', 'Bob', 'Charlie']
    html = '<ul>'
    for user in users:
        html += f'<li>{user}</li>'
    html += '</ul>'
    return html

def htmx_post_user(request):
    """POST request - create user"""
    name = request.get('name', '')
    if name:
        return f'<div class="success">Created user: {name}</div>'
    return '<div class="error">Name is required</div>'
```

## CSS Styling

Add CSS files in the `css/` directory for automatic injection:

```css
/* css/main.css */
.container {
    max-width: 800px;
    margin: 0 auto;
    padding: 20px;
}

.btn {
    background: #3498db;
    color: white;
    border: none;
    padding: 10px 20px;
    border-radius: 4px;
    cursor: pointer;
}

.btn:hover {
    background: #2980b9;
}

.alert {
    padding: 15px;
    margin: 10px 0;
    border-radius: 4px;
    background: #f8f9fa;
    border-left: 4px solid #3498db;
}

.alert.success {
    background: #d4edda;
    border-color: #28a745;
    color: #155724;
}

.alert.error {
    background: #f8d7da;
    border-color: #dc3545;
    color: #721c24;
}
```

## HTTP Method Control

Control HTTP methods via function naming:

```python
# Explicit method specification
def htmx_get_hello(request):      # → GET /api/demo/hello
    return '<div>Hello World!</div>'

def htmx_post_create(request):    # → POST /api/demo/create
    return '<div>Created!</div>'

def htmx_put_update(request):     # → PUT /api/demo/update
    return '<div>Updated!</div>'

def htmx_delete_remove(request):  # → DELETE /api/demo/remove
    return '<div>Deleted!</div>'

# Automatic detection
def htmx_hello(request):          # → GET (default)
    pass
def htmx_form(request):           # → POST (semantic detection)
    pass
def htmx_create_user(request):    # → POST (contains "create")
    pass
```

## Advanced Usage

### Async Context Manager

```python
async with htmlnojs("./my-project") as app:
    # Do other async work while server runs
    await some_other_async_task()
    # Server automatically stops when exiting
```

### Manual Control

```python
app = htmlnojs("./my-project", port=8080, verbose=True)

# Start the server
await app.start()

# Do other work...
await asyncio.sleep(10)

# Stop the server
await app.stop()
```

### Multiple Instances

```python
# Run multiple apps on different ports
app1 = htmlnojs("./frontend", port=3000)
app2 = htmlnojs("./admin", port=3001)

await app1.start()
await app2.start()

# Both servers running simultaneously
```

### Signal Handling

```python
# Graceful shutdown on Ctrl+C
app = htmlnojs("./my-project")
try:
    app.run_forever()  # Handles signals automatically
except KeyboardInterrupt:
    print("Shutting down...")
```

## Request Handling

The `request` parameter contains form data, query parameters, and more:

```python
def htmx_example(request):
    # Form data (POST requests)
    username = request.get('username', '')
    email = request.get('email', '')
    
    # Query parameters (GET requests)
    page = request.get('page', '1')
    
    # Multiple values
    tags = request.get('tags', [])  # For multiple checkboxes
    
    # Check if field exists
    if 'username' in request:
        return f'<div>Hello {username}!</div>'
    else:
        return '<div>Username required</div>'
```

## File Organization

### Large Applications

```
my-app/
├── templates/
│   ├── index.html
│   ├── auth/
│   │   ├── login.html      # → /auth/login
│   │   └── register.html   # → /auth/register
│   └── dashboard/
│       └── index.html      # → /dashboard/
├── css/
│   ├── main.css           # Global styles
│   ├── components.css     # Reusable components
│   └── pages.css          # Page-specific styles
└── py_htmx/
    ├── auth.py            # Authentication handlers
    ├── dashboard.py       # Dashboard handlers
    └── utils.py           # Utility functions
```

### Handler Organization

```python
# py_htmx/auth.py
def htmx_login(request):
    username = request.get('username')
    password = request.get('password')
    # Handle login logic
    return '<div>Login successful!</div>'

def htmx_register(request):
    # Handle registration
    return '<div>Registration complete!</div>'

# py_htmx/dashboard.py  
def htmx_get_stats(request):
    # Return dashboard statistics
    return '<div class="stats">...</div>'

def htmx_update_profile(request):
    # Handle profile updates
    return '<div>Profile updated!</div>'
```

## Error Handling

```python
def htmx_safe_handler(request):
    try:
        # Your logic here
        result = some_operation(request.get('data'))
        return f'<div class="success">{result}</div>'
    except ValueError as e:
        return f'<div class="error">Invalid input: {e}</div>'
    except Exception as e:
        return f'<div class="error">Something went wrong: {e}</div>'
```

## Best Practices

### 1. Return HTML Fragments
```python
# Good - returns HTML fragment
def htmx_get_user(request):
    return '<div class="user">John Doe</div>'

# Avoid - returns JSON (use regular API for this)
def htmx_get_user(request):
    return {"name": "John Doe"}  # Won't render in HTML
```

### 2. Use Semantic Class Names
```html
<div class="user-card">
    <button hx-delete="/api/users/delete" 
            hx-target="#user-list"
            class="btn btn-danger">
        Delete User
    </button>
</div>
```

### 3. Handle Empty States
```python
def htmx_get_items(request):
    items = get_items_from_db()
    
    if not items:
        return '<div class="empty-state">No items found</div>'
    
    html = '<div class="item-list">'
    for item in items:
        html += f'<div class="item">{item.name}</div>'
    html += '</div>'
    return html
```

### 4. Progressive Enhancement
```html
<!-- Works without JavaScript, enhanced with HTMX -->
<form action="/api/contact/send" method="post" 
      hx-post="/api/contact/send" hx-target="#result">
    <input type="email" name="email" required>
    <button type="submit">Subscribe</button>
</form>
```

## Configuration

### Port Configuration
```python
# Custom ports
app = htmlnojs(
    project_dir="./my-app",
    port=8080,           # Go server port
    verbose=True         # Enable debug logging
)
```

### Instance Management
```python
from htmlnojs import list_instances, get, stop_all

# List all running instances
instances = list_instances()

# Get specific instance by ID
app = get("my-app-id")

# Stop all instances
await stop_all()
```

## Troubleshooting

### Common Issues

1. **Port Already in Use**
   ```python
   # HTMLnoJS automatically finds available ports
   app = htmlnojs("./my-project")  # Will use 3000, 3001, etc.
   ```

2. **Go Server Not Found**
   ```bash
   # HTMLnoJS includes Go server automatically
   # No need to install Go separately
   ```

3. **Form Data Not Received**
   ```python
   # Make sure to use the correct method
   def htmx_form(request):  # POST by default
       message = request.get('message', '')  # Will work
   ```

4. **CSS Not Loading**
   ```
   # CSS files must be in css/ directory
   css/
   ├── main.css        ✅ Will load
   └── styles/
       └── theme.css   ✅ Will load as styles/theme.css
   ```

### Debug Mode
```python
# Enable verbose logging
app = htmlnojs("./my-project", verbose=True)
app.run_forever()

# Check logs for detailed request/response info
```

## Requirements

- Python 3.8+
- Go 1.19+ (automatically managed)
- Modern web browser with HTMX support

## Performance

HTMLnoJS is designed for speed:

- **Go server**: Handles static files and routing at C-like speeds
- **Python handlers**: Only process dynamic content
- **Minimal overhead**: Direct proxy between Go and Python
- **Concurrent**: Handles multiple requests simultaneously

## License

MIT License - see LICENSE file for details.