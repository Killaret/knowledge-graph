# Fixture catalog

| № | Обещание | Условие | Срок | Тест |
|---|----------|---------|------|------|
| P-01 | note appears on the graph | create → graph | 5 s | `promises.spec.ts` · `P-01` |
| P-02 | import lands on graph | import → node | 30 s | `other.spec.ts` · `imports a bookmark` |
| P-09 | backup restores everything | backup → restore | cycle | вне PR-CI: ручной прогон RELEASE-TEST-1 |
