# AgentFlow UI - Modular Architecture

A modern, modular web-based chat interface for agentic systems with comprehensive separation of concerns.

## 🚀 Overview

This is a complete refactor of the AgentFlow UI from a monolithic Go template to a modular, maintainable architecture. The new structure separates HTML, CSS, and JavaScript while solving the baseURL injection problem with minimal template code.

## 📁 Project Structure

```
ui/agentflow/
├── static/
│   ├── css/
│   │   └── styles.css          # Complete UI stylesheet
│   ├── js/
│   │   └── chat.js            # Main application logic
│   ├── images/
│   │   ├── favicon.ico        # Browser favicon
│   │   ├── favicon.svg        # Vector favicon
│   │   ├── favicon-96x96.png  # 96x96 PNG favicon
│   │   ├── apple-touch-icon*.png # Apple touch icons
│   │   └── web-app-manifest-*.png # PWA icons
│   └── site.webmanifest       # PWA manifest
├── templates/
│   └── chat-ui.html.tmpl      # Minimal Go template
└── README.md                  # This file
```

## 🔧 Key Features

### ✅ Modular Architecture
- **Separation of Concerns**: HTML, CSS, and JavaScript are in separate files
- **Maintainable**: Each component can be edited independently
- **Scalable**: Easy to add new features without touching template logic

### ✅ Minimal Template
- **Single Variable**: Only `{{.BaseURL}}` is templated
- **Small File**: Template is now ~180 lines vs. previous ~4500 lines
- **Clean Code**: No embedded CSS or JavaScript

### ✅ BaseURL Solution
- **Global Variable**: `window.AGENTFLOW_BASE_URL` injected once in template
- **JavaScript Access**: All API calls use the global variable
- **Path Resolution**: Static assets use proper baseURL paths

### ✅ Progressive Web App (PWA)
- **Web App Manifest**: Complete PWA configuration
- **Mobile Optimized**: Responsive design with touch support
- **App Icons**: Multiple icon sizes for different devices

## 🛠 Usage

### Basic Setup

1. **Serve Static Files**: Ensure your Go server serves the `/static/` directory
   ```go
   http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("ui/agentflow/static/"))))
   ```

2. **Load Template**: Use the minimal template file
   ```go
   tmpl, err := template.ParseFiles("ui/agentflow/templates/chat-ui.html.tmpl")
   ```

3. **Render with BaseURL**: Pass the baseURL parameter
   ```go
   data := struct{ BaseURL string }{BaseURL: "/api"}
   tmpl.Execute(w, data)
   ```

### Deployment Scenarios

#### Embedded Mode (Same Server)
```go
// BaseURL is empty - same server serves both UI and API
data := struct{ BaseURL string }{BaseURL: ""}
```

#### Standalone Mode (Separate Servers)
```go
// BaseURL points to API server
data := struct{ BaseURL string }{BaseURL: "http://localhost:4000"}
```

## 🔄 Migration from Original

### Before (Monolithic)
```html
<!-- 4500+ lines in single template -->
<style>/* 1700+ lines of CSS */</style>
<script>/* 2500+ lines of JavaScript */</script>
```

### After (Modular)
```html
<!-- ~180 lines minimal template -->
<script>window.AGENTFLOW_BASE_URL = '{{.BaseURL}}';</script>
<link rel="stylesheet" href="{{.BaseURL}}/static/css/styles.css">
<script src="{{.BaseURL}}/static/js/chat.js"></script>
```

## 📱 Mobile & PWA Features

- **Responsive Design**: Optimized for mobile devices
- **Touch Support**: Proper touch targets and interactions
- **Offline Capable**: PWA manifest for app-like experience
- **Safe Areas**: iPhone X/notch support
- **Viewport**: Dynamic viewport height for mobile browsers

## 🎨 Styling

The CSS is organized into logical sections:

- **Base Styles**: Reset, typography, layout
- **Side Menu**: Navigation and conversations
- **Chat Interface**: Messages, input, typing indicators
- **Tool System**: Tool selection and popups
- **Mobile**: Responsive design and touch optimization
- **Accessibility**: High contrast, reduced motion support

## 💻 JavaScript Architecture

The JavaScript is structured as a single `ChatUI` class with:

- **Modular Methods**: Each feature is a separate method
- **Event Handling**: Centralized event listener setup
- **State Management**: Conversation and UI state
- **API Integration**: Streaming chat with tool support
- **File Handling**: Multi-modal content support

## 🔧 Development

### Adding New Features

1. **Styles**: Add CSS to `static/css/styles.css`
2. **Logic**: Add methods to the `ChatUI` class in `static/js/chat.js`
3. **HTML**: Add elements to the template if needed

### Debugging

- Global `chatUI` instance available in browser console
- Debug function `testToolPopup()` for testing tool interactions
- Comprehensive console logging for streaming events

### Building Icons

To generate actual icon files from the SVG:

```bash
# Install ImageMagick
brew install imagemagick  # macOS
apt-get install imagemagick  # Ubuntu

# Generate icon files
cd static/images/
convert favicon.svg -resize 32x32 favicon.ico
convert favicon.svg -resize 96x96 favicon-96x96.png
convert favicon.svg -resize 180x180 apple-touch-icon.png
convert favicon.svg -resize 180x180 apple-touch-icon-180x180.png
convert favicon.svg -resize 192x192 web-app-manifest-192x192.png
convert favicon.svg -resize 512x512 web-app-manifest-512x512.png
```

## 🚀 Performance Benefits

- **Cacheable Assets**: CSS/JS files can be cached by browsers
- **Parallel Loading**: Stylesheets and scripts load in parallel
- **Smaller Templates**: Faster template parsing and rendering
- **CDN Ready**: Static assets can be served from CDN

## 🔒 Security

- **CSP Ready**: External CSS/JS for Content Security Policy
- **XSS Protection**: Proper input sanitization maintained
- **CORS Friendly**: Separate static assets support CORS

## 🐛 Troubleshooting

### BaseURL Issues
- **Empty BaseURL**: For same-server deployment, use empty string
- **API Calls Failing**: Check `window.AGENTFLOW_BASE_URL` in browser console
- **Asset 404s**: Verify static file server is configured correctly

### Mobile Issues
- **Viewport Problems**: Check viewport meta tag in template
- **Touch Targets**: Ensure buttons meet 44px minimum size
- **iOS Quirks**: Test in iOS Safari for specific mobile issues

---

## 📄 License

This follows the same license as the parent gomcptest project.