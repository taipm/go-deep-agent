# Rate Limiting Implementation - Quick Reference

**File:** TODO_RATE_LIMITING.md  
**Version Target:** v0.7.3  
**Total Tasks:** 31 tasks  
**Total Time:** ~10 hours  

---

## 📊 Quick Overview

```
31 Tasks = 5 Phases
├── Phase 1: Core Implementation (6 tasks, 2h)    ⬜⬜⬜⬜⬜⬜ 0%
├── Phase 2: Builder Integration (5 tasks, 1.5h) ⬜⬜⬜⬜⬜ 0%
├── Phase 3: Testing & Docs (6 tasks, 2h)        ⬜⬜⬜⬜⬜⬜ 0%
├── Phase 4: Redis Backend (6 tasks, 2h)         ⏸️ Optional
└── Phase 5: Polish & Release (6 tasks, 1.5h)    ⬜⬜⬜⬜⬜⬜ 0%
```

---

## 🎯 Key Features to Implement

### 1. Core Rate Limiter Interface
```go
type RateLimiter interface {
    Allow(ctx, key) (bool, error)
    Wait(ctx, key) error
    Reserve(ctx, key) (*Reservation, error)
    Stats(ctx, key) (*RateLimitStats, error)
}
```

### 2. Builder API
```go
ai := agent.NewOpenAI("gpt-4o", apiKey).
    WithRateLimitPerMinute(100, 10)  // 100 req/min, burst 10
```

### 3. Token Bucket Implementation
- Using `golang.org/x/time/rate`
- Per-key rate limiting
- Thread-safe with sync.Map

### 4. Integration Points
- Ask() method
- Stream() method
- All LLM interactions

---

## 🔄 Workflow per Task

Each task follows this pattern:

```bash
# 1. Start task (mark in TODO)
# 2. Implement feature (≤20 min)
# 3. Remove hardcoded values
# 4. Write/update tests
# 5. Run tests: go test ./... -v -race
# 6. Format: go fmt ./...
# 7. Commit: git commit -m "feat(rate-limit): [description]"
# 8. Update TODO (mark task as done)
```

---

## 📝 Task Template

```markdown
#### ✅ Task X.Y: [Task Name] (20 min)
- [ ] Sub-task 1
- [ ] Sub-task 2
- [ ] Sub-task 3
- [ ] Commit: "feat(rate-limit): [description]"

Status: ⬜ Not Started | 🟡 In Progress | ✅ Done
Time: Estimated 20min | Actual: __min
```

---

## 🎯 Success Criteria

**Done when:**
- [ ] All 31 tasks completed (or Phase 4 skipped)
- [ ] Tests pass: `go test ./... -v -race -cover`
- [ ] Coverage > 70% for new code
- [ ] No hardcoded values
- [ ] Code formatted: `go fmt ./...`
- [ ] Linter passes: `golangci-lint run`
- [ ] Examples work
- [ ] Documentation updated
- [ ] CHANGELOG updated
- [ ] Version tagged: v0.7.3
- [ ] GitHub release created

---

## 🚀 Next Steps

1. **Review TODO_RATE_LIMITING.md** - Full details
2. **Start with Task 1.1** - Create core interfaces
3. **Follow the workflow** - One task at a time
4. **Update progress** - Mark tasks done as you go
5. **Commit frequently** - After each task

---

## 📚 Reference

- Main TODO: [TODO_RATE_LIMITING.md](TODO_RATE_LIMITING.md)
- Design Guide: [docs/RATE_LIMITING_GUIDE.md](docs/RATE_LIMITING_GUIDE.md)
- Production Assessment: [PRODUCTION_READINESS_ASSESSMENT.md](PRODUCTION_READINESS_ASSESSMENT.md)

---

**Ready to start? → Open TODO_RATE_LIMITING.md and begin with Task 1.1**
