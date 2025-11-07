# How to Add COMPARISON.md to GitHub Wiki

## Method 1: Using GitHub Web UI (Easiest)

1. **Navigate to Wiki**
   - Go to: https://github.com/taipm/go-deep-agent/wiki
   - Click "Create the first page" or "New Page" button

2. **Create Comparison Page**
   - Page title: `Why go-deep-agent vs openai-go`
   - Copy content from `docs/COMPARISON.md`
   - Paste into wiki editor
   - Click "Save Page"

3. **Update Home Page**
   - Go to wiki Home page
   - Add link:
     ```markdown
     ## 📚 Documentation
     
     - [Why go-deep-agent vs openai-go](Why-go-deep-agent-vs-openai-go) - Detailed comparison
     - [Getting Started](#)
     - [API Reference](#)
     - [Examples](#)
     ```

## Method 2: Using Git (Advanced)

GitHub wikis are Git repositories. You can clone and manage them locally:

```bash
# Clone the wiki repository
git clone https://github.com/taipm/go-deep-agent.wiki.git

# Enter wiki directory
cd go-deep-agent.wiki

# Copy comparison document
cp ../docs/COMPARISON.md Why-go-deep-agent-vs-openai-go.md

# Create/Update Home page
cat > Home.md << 'EOF'
# go-deep-agent Wiki

Welcome to the go-deep-agent documentation!

## 📚 Documentation

- **[Why go-deep-agent vs openai-go](Why-go-deep-agent-vs-openai-go)** - Detailed comparison with code examples
- **[Getting Started](Getting-Started)** - Installation and quick start
- **[API Reference](API-Reference)** - Complete API documentation
- **[Examples](Examples)** - Code examples and tutorials
- **[Migration Guide](Migration-Guide)** - From v0.2.0 to v0.3.0

## 🚀 Quick Links

- [GitHub Repository](https://github.com/taipm/go-deep-agent)
- [Release Notes](https://github.com/taipm/go-deep-agent/releases)
- [Issues](https://github.com/taipm/go-deep-agent/issues)
EOF

# Commit and push
git add .
git commit -m "Add comprehensive comparison page"
git push origin master
```

## Method 3: Create Additional Wiki Pages

### Suggested Wiki Structure:

```
📖 go-deep-agent Wiki
├── Home.md
├── Why-go-deep-agent-vs-openai-go.md (from COMPARISON.md)
├── Getting-Started.md
├── API-Reference.md
├── Examples.md
│   ├── Basic-Usage.md
│   ├── Streaming.md
│   ├── Tool-Calling.md
│   ├── Multimodal.md
│   └── Error-Handling.md
├── Migration-Guide.md
├── Best-Practices.md
└── FAQ.md
```

### Content Sources:

- **Why-go-deep-agent-vs-openai-go.md**: Copy from `docs/COMPARISON.md`
- **Getting-Started.md**: Extract from README.md "Quick Start" section
- **API-Reference.md**: From `agent/README.md` or create comprehensive guide
- **Examples.md**: Link to examples/ directory with descriptions
- **Migration-Guide.md**: Extract from CHANGELOG.md
- **Best-Practices.md**: Production tips, error handling, performance
- **FAQ.md**: Common questions and answers

## Quick Wiki Setup Script

Save this as `setup-wiki.sh`:

```bash
#!/bin/bash

# Clone wiki
git clone https://github.com/taipm/go-deep-agent.wiki.git
cd go-deep-agent.wiki

# Copy comparison
cp ../docs/COMPARISON.md Why-go-deep-agent-vs-openai-go.md

# Create Home page
cat > Home.md << 'EOF'
# Welcome to go-deep-agent! 🚀

A powerful yet simple LLM agent library for Go with a modern Fluent Builder API.

## 🆚 Why Choose go-deep-agent?

**60-80% less code** than openai-go with **10x better developer experience**.

👉 **[See detailed comparison →](Why-go-deep-agent-vs-openai-go)**

## 📚 Documentation

- **[Why go-deep-agent vs openai-go](Why-go-deep-agent-vs-openai-go)** - Detailed comparison
- **[Getting Started](Getting-Started)** - Installation and quick start
- **[API Reference](API-Reference)** - Complete API documentation
- **[Examples](Examples)** - Code examples and tutorials
- **[Migration Guide](Migration-Guide)** - Version migration guides

## ✨ Key Features

- 🎯 **Fluent Builder API** - Natural method chaining
- 🖼️ **Multimodal Support** - GPT-4 Vision integration
- 🛠️ **Tool Calling** - Auto-execution support
- 🧠 **Auto Memory** - Conversation history management
- ⚡ **Error Recovery** - Built-in retry & backoff

## 🚀 Quick Example

\`\`\`go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are helpful").
    WithTemperature(0.7).
    WithMemory().
    Ask(ctx, "What is Go?")
\`\`\`

## 🔗 Links

- [GitHub Repository](https://github.com/taipm/go-deep-agent)
- [Latest Release](https://github.com/taipm/go-deep-agent/releases/latest)
- [Report Issues](https://github.com/taipm/go-deep-agent/issues)
EOF

# Create Getting Started page
cat > Getting-Started.md << 'EOF'
# Getting Started with go-deep-agent

## Installation

\`\`\`bash
go get github.com/taipm/go-deep-agent
\`\`\`

## Quick Start

### 1. Simple Chat

\`\`\`go
import "github.com/taipm/go-deep-agent/agent"

response, err := agent.NewOpenAI("gpt-4o-mini", "your-api-key").
    Ask(ctx, "What is Go?")
\`\`\`

### 2. With Configuration

\`\`\`go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSystem("You are a helpful assistant").
    WithTemperature(0.7).
    WithMaxTokens(500).
    Ask(ctx, "Explain quantum computing")
\`\`\`

### 3. Streaming

\`\`\`go
response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
    OnStream(func(content string) {
        fmt.Print(content)
    }).
    Stream(ctx, "Write a haiku")
\`\`\`

See [Examples](Examples) for more!
EOF

# Commit and push
git add .
git commit -m "Initialize wiki with comparison and getting started"
git push origin master

cd ..
echo "✅ Wiki setup complete!"
echo "Visit: https://github.com/taipm/go-deep-agent/wiki"
```

Make executable and run:
```bash
chmod +x setup-wiki.sh
./setup-wiki.sh
```

## After Setup

1. **Enable Wiki** (if not already enabled):
   - Go to repository Settings
   - Scroll to "Features"
   - Check "Wikis"

2. **Set Wiki Permissions**:
   - Settings → Manage access
   - Configure who can edit wiki

3. **Promote Wiki**:
   - Add wiki link to README.md
   - Pin important wiki pages
   - Reference in issues/PRs

## Verification

After setup, verify at:
- **Wiki Home**: https://github.com/taipm/go-deep-agent/wiki
- **Comparison Page**: https://github.com/taipm/go-deep-agent/wiki/Why-go-deep-agent-vs-openai-go

---

## Quick Manual Steps (5 minutes)

1. Go to https://github.com/taipm/go-deep-agent/wiki
2. Click "Create the first page" or "New Page"
3. Title: `Why go-deep-agent vs openai-go`
4. Copy entire content from `docs/COMPARISON.md`
5. Paste and click "Save Page"
6. Done! ✅

Your comparison is now accessible via wiki with better discoverability!
