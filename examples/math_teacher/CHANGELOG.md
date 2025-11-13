# Changelog - Math Teacher Example

## [1.1.0] - 2025-11-12

### Changed

- **Simplified agent initialization**: Removed explicit `.WithMemory()` call
- Updated to use library v0.7.10+ with automatic memory in `WithDefaults()`
- Cleaner, more intuitive code

**Before:**
```go
teacher := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithDefaults().
    WithMemory().            // Phải thêm thủ công
    WithPersona(persona).
    WithTools(...)
```

**After:**
```go
teacher := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithDefaults().          // Memory đã tự động có
    WithPersona(persona).
    WithTools(...)
```

### Documentation

- Updated README.md to reflect new `WithDefaults()` behavior
- Added MEMORY_FIX.md explaining the bug and fix
- Updated all code examples

### Fixed

- Memory now works correctly in interactive mode
- Agent remembers conversation context across multiple messages

---

## [1.0.0] - 2025-11-12

### Added

- Initial release of Math Teacher example
- 6 example scenarios:
  1. Simple addition
  2. Word problems
  3. Fractions
  4. Complex multi-step problems
  5. Basic geometry
  6. Interactive chat mode
- Vietnamese persona configuration (math_teacher.yaml)
- Comprehensive documentation (README.md, QUICKSTART.md, EXAMPLE_OUTPUT.md)
- Integration with MathTool and DateTimeTool
- Production-ready configuration with WithDefaults()

### Features

- Patient, step-by-step explanations
- Real-world examples (candy, toys, money)
- Encouraging tone for children
- Memory of 20 recent messages
- Auto-execution of mathematical tools
- Detailed logging for monitoring

---

## Version History

- **v1.1.0** (2025-11-12) - Updated for library v0.7.10+ with automatic memory
- **v1.0.0** (2025-11-12) - Initial release with manual memory configuration
