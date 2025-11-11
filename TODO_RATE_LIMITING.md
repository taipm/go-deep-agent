# TODO: Rate Limiting Implementation (v0.7.3)

**Target Version:** v0.7.3  
**Start Date:** 2025-11-11  
**Estimated Total Time:** 8-10 hours (split into 20-min tasks)  
**Status:** 🟡 In Progress

---

## 📋 Task Breakdown (Each task ≤ 20 minutes)

### Phase 1: Core Types & Interfaces (2 hours)

#### ✅ Task 1.1: Create rate_limiter.go with core interfaces (20 min)
- [ ] Create `agent/rate_limiter.go`
- [ ] Define `RateLimiter` interface (Allow, Wait, Reserve, Stats)
- [ ] Define `RateLimitConfig` struct
- [ ] Define `RateLimitStats` struct
- [ ] Define `Reservation` struct
- [ ] Add error codes (ErrRateLimitExceeded)
- [ ] Commit: "feat(rate-limit): Add core interfaces and types"

#### ✅ Task 1.2: Create token bucket implementation (20 min)
- [ ] Create `agent/rate_limiter_token_bucket.go`
- [ ] Implement `TokenBucketLimiter` struct
- [ ] Implement `NewTokenBucketLimiter()` constructor
- [ ] Implement `Allow()` method using golang.org/x/time/rate
- [ ] Add basic validation
- [ ] Commit: "feat(rate-limit): Implement token bucket limiter"

#### ✅ Task 1.3: Implement Wait() and Reserve() methods (20 min)
- [ ] Implement `Wait(ctx, key)` in TokenBucketLimiter
- [ ] Implement `Reserve(ctx, key)` in TokenBucketLimiter
- [ ] Implement `Stats(ctx, key)` for current stats
- [ ] Handle context cancellation
- [ ] Commit: "feat(rate-limit): Add Wait and Reserve methods"

#### ✅ Task 1.4: Add helper methods and cleanup (20 min)
- [ ] Implement `getLimiter(key)` helper with sync.Map
- [ ] Add `Close()` method for cleanup
- [ ] Add `Reset(key)` method to reset specific limiter
- [ ] Remove any hardcoded values
- [ ] Commit: "feat(rate-limit): Add helper methods and cleanup"

#### ✅ Task 1.5: Write basic unit tests (20 min)
- [ ] Create `agent/rate_limiter_test.go`
- [ ] Test TokenBucketLimiter creation
- [ ] Test Allow() with single key
- [ ] Test Allow() with multiple keys
- [ ] Test rate limiting behavior
- [ ] Commit: "test(rate-limit): Add basic unit tests"

#### ✅ Task 1.6: Write advanced unit tests (20 min)
- [ ] Test Wait() with timeout
- [ ] Test Reserve() functionality
- [ ] Test Stats() accuracy
- [ ] Test concurrent access
- [ ] Commit: "test(rate-limit): Add advanced unit tests"

---

### Phase 2: Builder Integration (1.5 hours)

#### ✅ Task 2.1: Add Builder methods (20 min)
- [ ] Update `agent/builder.go` with rateLimiter field
- [ ] Add `WithRateLimit(config)` method
- [ ] Add `WithRateLimitPerSecond(rate, burst)` method
- [ ] Add `WithRateLimitPerMinute(rate, burst)` method
- [ ] Add `WithRateLimitPerHour(rate, burst)` method
- [ ] Commit: "feat(rate-limit): Add Builder API methods"

#### ✅ Task 2.2: Add rate limit checking methods (20 min)
- [ ] Add `CheckRateLimit(ctx, key)` to Builder
- [ ] Add `WaitForRateLimit(ctx, key)` to Builder
- [ ] Add `ReserveRateLimit(ctx, key)` to Builder
- [ ] Add `GetRateLimitStats(ctx, key)` to Builder
- [ ] Commit: "feat(rate-limit): Add rate limit checking methods"

#### ✅ Task 2.3: Integrate into Ask() method (20 min)
- [ ] Update `buildMessages()` or `Ask()` to check rate limit
- [ ] Add rate limit check before API call
- [ ] Return ErrRateLimitExceeded if blocked
- [ ] Add stats to response metadata
- [ ] Commit: "feat(rate-limit): Integrate rate limit into Ask()"

#### ✅ Task 2.4: Integrate into Stream() method (20 min)
- [ ] Update `Stream()` to check rate limit
- [ ] Handle rate limit before streaming
- [ ] Ensure proper error handling
- [ ] Commit: "feat(rate-limit): Integrate rate limit into Stream()"

#### ✅ Task 2.5: Add default key extraction function (20 min)
- [ ] Create `extractRateLimitKey(ctx)` helper
- [ ] Support per-IP extraction (if available in context)
- [ ] Support per-user extraction (if available)
- [ ] Fallback to "global" key
- [ ] Commit: "feat(rate-limit): Add default key extraction"

---

### Phase 3: Testing & Documentation (2 hours)

#### ✅ Task 3.1: Write Builder integration tests (20 min)
- [ ] Create `agent/builder_rate_limit_test.go`
- [ ] Test WithRateLimitPerSecond()
- [ ] Test WithRateLimitPerMinute()
- [ ] Test rate limit in Ask()
- [ ] Commit: "test(rate-limit): Add Builder integration tests"

#### ✅ Task 3.2: Write concurrent tests (20 min)
- [ ] Test concurrent Ask() calls with rate limit
- [ ] Test multiple users with different keys
- [ ] Test rate limit stats accuracy under load
- [ ] Commit: "test(rate-limit): Add concurrent tests"

#### ✅ Task 3.3: Write edge case tests (20 min)
- [ ] Test zero rate (should error)
- [ ] Test negative burst (should error)
- [ ] Test context cancellation during Wait()
- [ ] Test rate limit with nil context
- [ ] Commit: "test(rate-limit): Add edge case tests"

#### ✅ Task 3.4: Create basic example (20 min)
- [ ] Create `examples/rate_limit_basic.go`
- [ ] Example 1: Basic rate limiting
- [ ] Example 2: Per-user rate limiting
- [ ] Example 3: Graceful degradation
- [ ] Add comments and explanations
- [ ] Commit: "docs(rate-limit): Add basic examples"

#### ✅ Task 3.5: Create advanced example (20 min)
- [ ] Create `examples/rate_limit_advanced.go`
- [ ] Example 4: Multi-tier rate limiting
- [ ] Example 5: Rate limit with metrics
- [ ] Example 6: Rate limit statistics
- [ ] Test examples work
- [ ] Commit: "docs(rate-limit): Add advanced examples"

#### ✅ Task 3.6: Update README.md (20 min)
- [ ] Add Rate Limiting section to README
- [ ] Add quick start example
- [ ] Add to Features list
- [ ] Update Builder API methods list
- [ ] Commit: "docs: Add Rate Limiting to README"

---

### Phase 4: Redis Backend (Optional, 2 hours)

#### ⏸️ Task 4.1: Create Redis rate limiter (20 min)
- [ ] Create `agent/rate_limiter_redis.go`
- [ ] Define `RedisRateLimiter` struct
- [ ] Implement constructor with Redis client
- [ ] Add basic validation
- [ ] Commit: "feat(rate-limit): Add Redis rate limiter structure"

#### ⏸️ Task 4.2: Implement Redis Allow() with Lua script (20 min)
- [ ] Write Lua script for atomic rate limit check
- [ ] Implement `Allow()` using Redis EVAL
- [ ] Use sliding window algorithm
- [ ] Handle Redis errors gracefully
- [ ] Commit: "feat(rate-limit): Implement Redis Allow() method"

#### ⏸️ Task 4.3: Implement Redis Wait() and Stats() (20 min)
- [ ] Implement `Wait()` with retry logic
- [ ] Implement `Stats()` to get current counts
- [ ] Add TTL management for keys
- [ ] Commit: "feat(rate-limit): Add Redis Wait and Stats"

#### ⏸️ Task 4.4: Add Redis tests (20 min)
- [ ] Create `agent/rate_limiter_redis_test.go`
- [ ] Use miniredis for testing
- [ ] Test Redis Allow() behavior
- [ ] Test Redis Stats()
- [ ] Test Redis connection failures
- [ ] Commit: "test(rate-limit): Add Redis rate limiter tests"

#### ⏸️ Task 4.5: Add Builder support for Redis (20 min)
- [ ] Add `WithRedisRateLimit(client, rate, window)` method
- [ ] Update Builder to support Redis limiter
- [ ] Add example with Redis
- [ ] Commit: "feat(rate-limit): Add Redis rate limiter to Builder"

#### ⏸️ Task 4.6: Document Redis rate limiting (20 min)
- [ ] Update RATE_LIMITING_GUIDE.md with Redis examples
- [ ] Add Redis setup instructions
- [ ] Add distributed rate limiting best practices
- [ ] Commit: "docs(rate-limit): Add Redis documentation"

---

### Phase 5: Polish & Release (1.5 hours)

#### ✅ Task 5.1: Code review and cleanup (20 min)
- [ ] Remove all hardcoded values
- [ ] Check for code duplication
- [ ] Ensure consistent error messages
- [ ] Run `go fmt ./...`
- [ ] Run `golangci-lint run`
- [ ] Commit: "refactor(rate-limit): Code cleanup and formatting"

#### ✅ Task 5.2: Run full test suite (20 min)
- [ ] Run `go test ./... -v -race -cover`
- [ ] Ensure coverage > 70% for new code
- [ ] Fix any failing tests
- [ ] Verify no race conditions
- [ ] Commit: "test(rate-limit): Ensure full test coverage"

#### ✅ Task 5.3: Update CHANGELOG.md (20 min)
- [ ] Add v0.7.3 section to CHANGELOG
- [ ] Document all new features
- [ ] Add usage examples
- [ ] Add breaking changes (if any)
- [ ] Commit: "docs: Update CHANGELOG for v0.7.3"

#### ✅ Task 5.4: Update version and dependencies (20 min)
- [ ] Add `golang.org/x/time` to go.mod if needed
- [ ] Run `go mod tidy`
- [ ] Update any version references
- [ ] Commit: "chore: Update dependencies for v0.7.3"

#### ✅ Task 5.5: Create release notes (20 min)
- [ ] Create RELEASE_NOTES_v0.7.3.md
- [ ] Summarize rate limiting features
- [ ] Add installation instructions
- [ ] Add migration guide
- [ ] Commit: "docs: Add release notes for v0.7.3"

#### ✅ Task 5.6: Tag and release (20 min)
- [ ] Create git tag v0.7.3
- [ ] Push tag to remote
- [ ] Create GitHub release
- [ ] Verify module on pkg.go.dev
- [ ] Commit: "release: v0.7.3 - Rate Limiting"

---

## 📊 Progress Tracking

### Statistics
- **Total Tasks:** 31 tasks
- **Completed:** 0 tasks (0%)
- **In Progress:** 0 tasks
- **Remaining:** 31 tasks (100%)
- **Estimated Time:** 10.3 hours
- **Actual Time:** 0 hours

### Phase Status
- Phase 1 (Core): ⬜⬜⬜⬜⬜⬜ 0/6 (0%)
- Phase 2 (Builder): ⬜⬜⬜⬜⬜ 0/5 (0%)
- Phase 3 (Testing): ⬜⬜⬜⬜⬜⬜ 0/6 (0%)
- Phase 4 (Redis): ⏸️⏸️⏸️⏸️⏸️⏸️ 0/6 (Optional)
- Phase 5 (Release): ⬜⬜⬜⬜⬜⬜ 0/6 (0%)

---

## 🎯 Current Task

**Next Task:** Task 1.1 - Create rate_limiter.go with core interfaces  
**Estimated Time:** 20 minutes  
**Status:** Ready to start

---

## 📝 Notes & Decisions

### Design Decisions
1. **Use Token Bucket Algorithm**: Best balance between simplicity and flexibility
2. **Use golang.org/x/time/rate**: Battle-tested, maintained by Go team
3. **Per-key Rate Limiting**: Support multiple users with separate limits
4. **Optional Redis Backend**: For distributed deployments (Phase 4)
5. **Backward Compatible**: No breaking changes, all features opt-in

### Dependencies
- `golang.org/x/time/rate`: Token bucket implementation
- `github.com/redis/go-redis/v9`: Redis client (Phase 4 only)
- `github.com/alicebob/miniredis/v2`: Redis mock for testing (Phase 4 only)

### Testing Strategy
- Unit tests for each component
- Integration tests for Builder API
- Concurrent tests for race conditions
- Edge case tests for error handling
- Example programs as documentation tests

### Commit Strategy
- Commit after each task (small, focused commits)
- Use conventional commit format: feat/test/docs/refactor/chore
- Each commit should be self-contained and buildable

---

## 🚀 Quick Start Commands

```bash
# Start implementation
git checkout -b feature/rate-limiting

# After each task
git add .
git commit -m "feat(rate-limit): [task description]"

# Run tests
go test ./... -v -race -cover

# Check coverage
go test ./agent -coverprofile=coverage.out
go tool cover -html=coverage.out

# Format code
go fmt ./...

# Lint
golangci-lint run

# Final release
git tag v0.7.3
git push origin v0.7.3
gh release create v0.7.3 --notes-file RELEASE_NOTES_v0.7.3.md
```

---

## 📚 Reference Documents

- [RATE_LIMITING_GUIDE.md](docs/RATE_LIMITING_GUIDE.md) - Detailed guide
- [PRODUCTION_READINESS_ASSESSMENT.md](PRODUCTION_READINESS_ASSESSMENT.md) - Why we need this
- [golang.org/x/time/rate docs](https://pkg.go.dev/golang.org/x/time/rate)

---

## ✅ Completion Checklist

Before marking as complete:
- [ ] All unit tests pass
- [ ] Coverage > 70% for new code
- [ ] No race conditions detected
- [ ] All examples run successfully
- [ ] Documentation updated
- [ ] CHANGELOG updated
- [ ] No hardcoded values
- [ ] Code formatted (go fmt)
- [ ] Linter passes (golangci-lint)
- [ ] Module published to pkg.go.dev
- [ ] GitHub release created

---

**Last Updated:** 2025-11-11  
**Maintainer:** taipm  
**Target Completion:** 2025-11-12
