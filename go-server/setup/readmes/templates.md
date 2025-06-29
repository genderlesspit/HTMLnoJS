# 📁 Templates Directory

> **Important**: This directory should only contain HTML files. Other file types are not supported and will generate warnings.

## 🚀 How It Works

Each HTML file in this directory automatically generates a corresponding route on the Go server:

- `index.html` → serves at `/` (root)
- `about.html` → serves at `/about`
- `contact.html` → serves at `/contact`

## 📝 HTML File Guidelines

### ✅ Do:
- **Classes only**: Use CSS class names for styling
- **HTMX attributes**: Add `hx-get`, `hx-post`, etc. for interactivity
- **Semantic HTML**: Use proper HTML5 elements

### ❌ Don't:
- **No inline CSS**: All styling is attached automatically
- **No `<style>` tags**: CSS belongs in the `/css` directory
- **No JavaScript**: Use HTMX for dynamic behavior

## 🔗 HTMX Integration

Route your HTMX requests to Python functions based on the file structure in `/py_htmx`:

```html
<!-- This will call a Python function in py_htmx/ -->
<button hx-post="/api/submit" hx-target="#result">
    Submit Form
</button>
```

## 📂 Example Structure

```
templates/
├── index.html      # Homepage (/)
├── login.html      # Login page (/login)
└── dashboard.html  # Dashboard (/dashboard)
```

Happy coding! 🎉