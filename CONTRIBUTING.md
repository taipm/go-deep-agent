# Contributing to go-deep-agent

Thank you for your interest in contributing! This document provides guidelines and information for contributors.

## 🚀 Current Status

**We are currently implementing Option B: Full Rewrite with Builder Pattern**

See [TODO.md](TODO.md) for the roadmap and [DESIGN_DECISIONS.md](DESIGN_DECISIONS.md) for design rationale.

## 📋 How to Contribute

### Reporting Issues

- Check if the issue already exists
- Provide clear reproduction steps
- Include Go version, OS, and relevant code snippets
- Tag appropriately (bug, enhancement, question, etc.)

### Suggesting Features

- Check [TODO.md](TODO.md) - it might already be planned!
- Open an issue with tag `enhancement`
- Describe the use case and expected behavior
- Consider backward compatibility

### Submitting Pull Requests

1. **Check TODO.md first** - coordinate to avoid duplicate work
2. **Open an issue** to discuss major changes
3. **Follow the style guide** (see below)
4. **Write tests** for new functionality
5. **Update documentation** as needed
6. **Run tests** before submitting

## 🏗️ Development Setup

### Prerequisites

- Go 1.23.3 or higher
- OpenAI API key (for integration tests)
- Ollama installed (for Ollama tests)

### Clone and Setup

```bash
git clone https://github.com/taipm/go-deep-agent.git
cd go-deep-agent

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build ./...
```

### Running Integration Tests

```bash
# Set API key
export OPENAI_API_KEY=your-key-here

# Run all tests including integration
go test ./... -tags=integration

# Run only unit tests (no API calls)
go test ./... -short
```

## 🎨 Code Style

### General Guidelines

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Run `gofmt` before committing
- Run `go vet` to catch common errors
- Use `golangci-lint` for comprehensive linting

### Naming Conventions

- **Exported types**: PascalCase
- **Unexported types**: camelCase
- **Methods**: PascalCase (exported) or camelCase (unexported)
- **Variables**: camelCase
- **Constants**: PascalCase or UPPER_CASE for package-level

### Comments

- All exported types and functions MUST have doc comments
- Start with the name: `// Builder creates...`
- Keep it concise but informative
- Add examples for complex functionality

Example:
```go
// WithTemperature sets the sampling temperature (0.0 to 2.0).
// Higher values make output more random, lower values more deterministic.
// Default is 1.0.
func (b *Builder) WithTemperature(t float64) *Builder {
    b.temperature = &t
    return b
}
```

### Error Handling

- Always wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
- Use custom error types for specific cases (see DESIGN_DECISIONS.md)
- Return errors rather than panicking (except in `Ask()` which is intentionally panic-based)

### Testing

- Write table-driven tests when appropriate
- Use subtests with `t.Run()`
- Mock external dependencies
- Aim for >80% code coverage

Example:
```go
func TestBuilder_WithTemperature(t *testing.T) {
    tests := []struct {
        name        string
        temperature float64
        want        float64
    }{
        {"zero", 0.0, 0.0},
        {"default", 1.0, 1.0},
        {"high", 2.0, 2.0},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            b := New("gpt-4o-mini").WithTemperature(tt.temperature)
            if *b.temperature != tt.want {
                t.Errorf("got %v, want %v", *b.temperature, tt.want)
            }
        })
    }
}
```

## 📝 Documentation Guidelines

### Code Documentation

- Document all exported symbols
- Include examples in godoc
- Keep documentation up-to-date with code

### README and Guides

- Update README.md for user-facing changes
- Update QUICK_REFERENCE.md for new features
- Update agent/README.md for API changes
- Add examples to examples/ directory

### Changelog

- Add entries to CHANGELOG.md
- Follow [Keep a Changelog](https://keepachangelog.com/) format
- Include migration notes for breaking changes

## 🧪 Testing Strategy

### Unit Tests

- Test individual functions/methods
- Mock external dependencies
- Fast execution (<1s for all unit tests)
- No network calls

### Integration Tests

- Test with real OpenAI API
- Test with real Ollama instance
- Tag with `// +build integration`
- Document required setup in test file

### Example Tests

```go
// Unit test - no external dependencies
func TestBuilder_New(t *testing.T) {
    b := New("gpt-4o-mini")
    if b.model != "gpt-4o-mini" {
        t.Errorf("expected gpt-4o-mini, got %s", b.model)
    }
}

// Integration test - requires API key
// +build integration
func TestBuilder_Ask_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        t.Skip("OPENAI_API_KEY not set")
    }
    
    b := NewOpenAI("gpt-4o-mini", apiKey)
    response, err := b.AskE(context.Background(), "Say hello")
    if err != nil {
        t.Fatalf("Ask failed: %v", err)
    }
    if response == "" {
        t.Error("expected non-empty response")
    }
}
```

## 🔄 Git Workflow

### Branching

- `main` - stable, released code
- `develop` - integration branch (currently for Builder API)
- `feature/*` - feature branches
- `bugfix/*` - bug fix branches

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance tasks

Examples:
```
feat(builder): add WithTemperature method

Adds temperature control to Builder API.
Temperature range is 0.0 to 2.0.

Closes #123
```

```
fix(streaming): handle empty chunks correctly

Previously empty chunks would cause nil pointer panic.
Now safely ignored.

Fixes #456
```

### Pull Request Process

1. **Create feature branch** from `develop`
   ```bash
   git checkout -b feature/add-temperature
   ```

2. **Make changes** with clear commits

3. **Push and create PR** against `develop`
   ```bash
   git push origin feature/add-temperature
   ```

4. **PR Template** (fill in):
   ```markdown
   ## Description
   Brief description of changes
   
   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update
   
   ## Checklist
   - [ ] Tests added/updated
   - [ ] Documentation updated
   - [ ] Code formatted (gofmt)
   - [ ] All tests passing
   - [ ] CHANGELOG.md updated
   
   ## Related Issues
   Closes #123
   ```

5. **Code review** - address feedback

6. **Merge** after approval

## 🎯 Areas Where Help is Needed

See [TODO.md](TODO.md) for detailed tasks. High priority areas:

### Phase 1-2 (Core Implementation)
- [ ] Message helpers implementation
- [ ] Builder core structure
- [ ] Basic execution methods
- [ ] Advanced parameters

### Phase 3 (Streaming)
- [ ] ChatCompletionAccumulator integration
- [ ] Streaming callbacks
- [ ] Refusal handling

### Phase 4 (Tools)
- [ ] Tool definition API
- [ ] Auto-execution loop
- [ ] Tool result handling

### Documentation
- [ ] API examples
- [ ] Tutorial blog posts
- [ ] Video tutorials
- [ ] Comparison with other libraries

### Testing
- [ ] Unit test coverage
- [ ] Integration tests
- [ ] Benchmarks
- [ ] Edge case testing

## 🐛 Known Issues

See [GitHub Issues](https://github.com/taipm/go-deep-agent/issues) for current bugs and enhancement requests.

## 💬 Communication

- **GitHub Issues** - Bug reports, feature requests
- **GitHub Discussions** - General questions, ideas
- **Pull Requests** - Code contributions

## 📜 License

By contributing, you agree that your contributions will be licensed under the MIT License.

## 🙏 Recognition

Contributors will be recognized in:
- README.md contributors section
- Release notes
- CHANGELOG.md

Thank you for contributing to go-deep-agent! 🚀
