# Regression Testing Plan Summary

**Version:** 1.1  
**Last Updated:** 2026-08-03  
**Comprehensive 25-step isolated regression testing plan for production readiness:**

### Regression Test Cycle (Isolated Model)
- **Document:** `docs/REGRESSION_TEST_PLAN.md`
- **Script:** `scripts/testing/run-full-test-cycle.ps1` (Windows) or `scripts/testing/run-full-test-cycle.sh` (Linux/Mac)
- **Identity Check:** `scripts/ci/check-stacks-identity.ps1` (verifies dev/personal/test consistency)
- **Health Check:** `scripts/ci/check-stacks-health.ps1 -Stack <dev|personal|test|all>`

### Isolated Testing Model
**⚠️ IMPORTANT:** The test cycle uses an isolated model where dev and personal stacks are stopped during testing to prevent resource conflicts and ensure accurate test results.

### Test Steps (25 total)
1. **Step 0:** Capture dev stack state snapshot (containers, health, API)
2. **Step 1:** Stop dev stack (`docker compose down`)
3. **Step 2:** Stop personal stack (`docker compose -f docker-compose.personal.yml down`)
4. **Step 3:** Check stacks identity (dev/personal/test consistency)
5. **Step 4:** Start test stack (`start-test.ps1`)
6. **Step 5:** Seed test data (`seed-test-data.ps1`)
7. **Step 6:** Docker build verification
8. **Step 7:** NLP service tests
9. **Step 8:** Backend unit tests (`go test ./...`)
10. **Step 9:** Backend integration tests (`go test -tags=integration ./...` on Linux/WSL Docker)
11. **Step 10:** Backend API verification (manual/contract)
12. **Step 11:** Asynchronous tasks verification
13. **Step 12:** PGVECTOR verification
14. **Step 13:** Redis & MongoDB verification
15. **Step 14:** Frontend unit tests (`npm run test:unit`)
16. **Step 15:** Manual testing instructions
17. **Step 16:** Public graph verification
18. **Step 17:** CI/CD verification
19. **Step 18:** Documentation verification (`docs/AGENTS.md`, `.windsurfrules`, architecture diagrams)
20. **Step 19:** Stop test stack (`stop-test.ps1`)
21. **Step 20:** Cleanup temporary files (`coverage.out`, `*.cov`, `*.tmp`, `*.log`)
22. **Step 21:** Start dev stack (`docker compose up -d --wait`)
23. **Step 22:** Start personal stack (`docker compose -f docker-compose.personal.yml up -d --wait`)
24. **Step 23:** Compare dev stack state with pre-test snapshot, compare dev and personal stacks for identity, check health
25. **Step 24:** Auto-commit with test success marker (if all checks pass)

### Automatic State Verification
- **Pre-test snapshot:** Captures dev stack state before testing
- **Post-test comparison:** Compares dev stack state after testing
- **Dev/Personal identity:** Verifies dev and personal stacks are identical
- **Auto-commit:** Only if dev state unchanged and dev/personal identical
- **Failure handling:** Stops with exit code 1 if differences found

### Frequency
- **Full Regression:** Before each production deployment
- **Quick Regression:** Before each major feature release (Steps 0-6, 8-14, 19-23)
- **Smoke Regression:** After each minor feature release (Steps 0-6, 8-9, 13-14, 19-23)
- **Identity Check:** Before each manual testing session

### Exit Criteria
- **PASS:** Dev state unchanged, dev/personal identical, all stacks healthy, all tests pass
- **FAIL:** Dev state changed, dev/personal not identical, stacks not healthy, test failure, infrastructure failure, data leakage

### See Also
- [REGRESSION_TEST_PLAN.md](REGRESSION_TEST_PLAN.md) — Complete regression testing procedures
- [TESTING.md](TESTING.md) — Testing infrastructure and procedures
- [FINAL_TEST_REPORT.md](archive/FINAL_TEST_REPORT.md) — Latest test results

---
