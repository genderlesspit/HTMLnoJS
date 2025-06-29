# 🎨 CSS Directory

> **Purpose**: This directory contains all stylesheets that will be automatically attached to your HTML templates.

## 🚀 How It Works

All CSS files in this directory are automatically included in your HTML pages:

- **Global styles**: `main.css`, `global.css` loaded first
- **Component styles**: `buttons.css`, `forms.css` for specific components
- **Theme styles**: `dark.css`, `light.css` for theming
- **Automatic loading**: No need to manually link stylesheets

## 📝 CSS File Guidelines

### ✅ Do:
- **Use class selectors**: `.button`, `.card`, `.header`
- **Follow BEM methodology**: `.button--primary`, `.card__title`
- **Mobile-first design**: Start with mobile styles, add desktop with media queries
- **Use CSS custom properties**: `--primary-color: #3498db;`

### ❌ Don't:
- **Avoid ID selectors**: Use classes instead of `#header`
- **No !important**: Write specific selectors instead
- **Avoid inline styles**: Keep all styles in CSS files
- **Don't duplicate**: Use CSS variables for repeated values

## 🎯 Recommended File Structure

### Core Files:
- **`reset.css`**: CSS reset/normalize
- **`variables.css`**: CSS custom properties
- **`main.css`**: Global styles and layout

### Component Files:
- **`buttons.css`**: Button styles
- **`forms.css`**: Form and input styles
- **`cards.css`**: Card component styles
- **`navigation.css`**: Navigation and menu styles

### Utility Files:
- **`utilities.css`**: Helper classes
- **`responsive.css`**: Media queries and responsive utilities

## 🔧 CSS Variables Example

```css
/* variables.css */
:root {
  --primary-color: #3498db;
  --secondary-color: #2ecc71;
  --text-color: #333;
  --bg-color: #f8f9fa;
  --border-radius: 8px;
  --spacing-sm: 0.5rem;
  --spacing-md: 1rem;
  --spacing-lg: 2rem;
}
```

## 📱 Responsive Design

```css
/* Mobile first approach */
.container {
  padding: var(--spacing-sm);
}

/* Tablet and up */
@media (min-width: 768px) {
  .container {
    padding: var(--spacing-md);
  }
}

/* Desktop and up */
@media (min-width: 1024px) {
  .container {
    padding: var(--spacing-lg);
    max-width: 1200px;
    margin: 0 auto;
  }
}
```

## 📂 Example Structure

```
css/
├── reset.css        # CSS reset
├── variables.css    # CSS custom properties
├── main.css         # Global styles
├── components/
│   ├── buttons.css  # Button styles
│   ├── forms.css    # Form styles
│   └── cards.css    # Card styles
└── utilities.css    # Helper classes
```

Happy styling! 🎉