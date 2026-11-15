# CONFIG_REGISTRY — реестр ключей config/*.json

Сгенерировано `scripts/testing/generate-config-registry.mjs`; проверка дрейфа — `check-config-registry.mjs`.
Порядок приоритета: **env > knowledge-graph.config.json (или config/*.json) > дефолт в коде** — backend,
graph-service и NLP (nlp.* → env через `app/config.py` при старте); фронтенд: `config/*.json` вшиваются при сборке
(`npm run build-config` + `npm run build`), переменных окружения в рантайме нет — смена `config/*.json`
без пересборки фронтенда не действует.

| Ключ | Файл | Читает | Env-переопределение | Compose | В CONFIGURATION | Статус |
|---|---|---|---|---|---|---|
| `backend.app_env` | backend.json | backend config.go → AppEnv | `APP_ENV` | compose.deploy, compose.personal, compose.test, compose | + | ok |
| `backend.server.rate_limit.enabled` | backend.json | backend config.go → ServerRateLimitEnabled | `SERVER_RATE_LIMIT_ENABLED` | — | + | ok |
| `backend.server.rate_limit.requests` | backend.json | backend config.go → ServerRateLimitRequests | `SERVER_RATE_LIMIT_REQUESTS` | — | + | ok |
| `backend.server.rate_limit.window_seconds` | backend.json | backend config.go → ServerRateLimitWindowSeconds | `SERVER_RATE_LIMIT_WINDOW_SECONDS` | — | + | ok |
| `backend.server.rate_limit.endpoints.notes_create` | backend.json | backend config.go → ServerRateLimitEndpoints (через backend.server.rate_limit.endpoints) | — | — | — | ok |
| `backend.server.rate_limit.endpoints.links_create` | backend.json | backend config.go → ServerRateLimitEndpoints (через backend.server.rate_limit.endpoints) | — | — | — | ok |
| `backend.server.rate_limit.endpoints.notes_update` | backend.json | backend config.go → ServerRateLimitEndpoints (через backend.server.rate_limit.endpoints) | — | — | — | ok |
| `backend.server.rate_limit.endpoints.notes_delete` | backend.json | backend config.go → ServerRateLimitEndpoints (через backend.server.rate_limit.endpoints) | — | — | — | ok |
| `backend.server.rate_limit.fallback_ports` | backend.json | backend config.go → ServerFallbackPorts | — | — | + | ok |
| `backend.database.retry_max_attempts` | backend.json | backend config.go → DatabaseRetryMaxAttempts | `DATABASE_RETRY_MAX_ATTEMPTS` | — | + | ok |
| `backend.database.retry_delay_seconds` | backend.json | backend config.go → DatabaseRetryDelaySeconds | `DATABASE_RETRY_DELAY_SECONDS` | — | + | ok |
| `backend.database.migrations_fail_on_error` | backend.json | backend config.go → MigrationsFailOnError | `MIGRATIONS_FAIL_ON_ERROR` | — | + | ok |
| `backend.database.pool.max_open_conns` | backend.json | backend config.go → DatabasePoolMaxOpenConns | `POSTGRES_MAX_OPEN_CONNS` | — | + | ok |
| `backend.database.pool.max_idle_conns` | backend.json | backend config.go → DatabasePoolMaxIdleConns | `POSTGRES_MAX_IDLE_CONNS` | — | + | ok |
| `backend.database.pool.conn_max_lifetime_seconds` | backend.json | backend config.go → DatabasePoolConnMaxLifetimeSeconds | `POSTGRES_CONN_MAX_LIFETIME_SECONDS` | — | + | ok |
| `backend.database.pool.conn_max_idle_time_seconds` | backend.json | backend config.go → DatabasePoolConnMaxIdleTimeSeconds | `POSTGRES_CONN_MAX_IDLE_TIME_SECONDS` | — | + | ok |
| `backend.database.pool.stats_interval_seconds` | backend.json | backend config.go → DatabasePoolStatsIntervalSeconds | `POSTGRES_POOL_STATS_INTERVAL_SECONDS` | — | + | ok |
| `backend.search.fulltext_languages` | backend.json | backend config.go → SearchFulltextLanguages | — | — | — | ok |
| `backend.search.ranking_weights.russian` | backend.json | backend config.go → SearchRankingWeights (через backend.search.ranking_weights) | — | — | — | ok |
| `backend.search.ranking_weights.simple` | backend.json | backend config.go → SearchRankingWeights (через backend.search.ranking_weights) | — | — | — | ok |
| `backend.search.fallback_to_ilike` | backend.json | backend config.go → SearchFallbackToILike | `SEARCH_FALLBACK_TO_ILIKE` | — | + | ok |
| `backend.recommendation.depth` | backend.json | backend config.go → RecommendationDepth | `RECOMMENDATION_DEPTH` | — | + | ok |
| `backend.recommendation.decay` | backend.json | backend config.go → RecommendationDecay | `RECOMMENDATION_DECAY` | — | + | ok |
| `backend.recommendation.top_n` | backend.json | backend config.go → RecommendationTopN | `RECOMMENDATION_TOP_N` | compose.personal | + | ok |
| `backend.recommendation.alpha` | backend.json | backend config.go → RecommendationAlpha | `RECOMMENDATION_ALPHA` | — | + | ok |
| `backend.recommendation.beta` | backend.json | backend config.go → RecommendationBeta | `RECOMMENDATION_BETA` | — | + | ok |
| `backend.recommendation.gamma` | backend.json | backend config.go → RecommendationGamma | `RECOMMENDATION_GAMMA` | — | + | ok |
| `backend.recommendation.cache_ttl_seconds` | backend.json | backend config.go → RecommendationCacheTTL | `RECOMMENDATION_CACHE_TTL_SECONDS` | — | + | ok |
| `backend.recommendation.task_delay_seconds` | backend.json | backend config.go → RecommendationTaskDelaySeconds | `RECOMMENDATION_TASK_DELAY_SECONDS` | compose.personal | + | ok |
| `backend.recommendation.batch_rate_limit` | backend.json | backend config.go → RecommendationBatchRateLimit | `RECOMMENDATION_BATCH_RATE_LIMIT` | — | + | ok |
| `backend.recommendation.fallback_enabled` | backend.json | backend config.go → RecommendationFallbackEnabled | `RECOMMENDATION_FALLBACK_ENABLED` | — | + | ok |
| `backend.recommendation.fallback_ttl_seconds` | backend.json | backend config.go → RecommendationFallbackTTL | `RECOMMENDATION_FALLBACK_TTL_SECONDS` | — | + | ok |
| `backend.recommendation.fallback_semantic_enabled` | backend.json | backend config.go → RecommendationFallbackSemanticEnabled | `RECOMMENDATION_FALLBACK_SEMANTIC_ENABLED` | — | + | ok |
| `backend.recommendation.bfs_aggregation` | backend.json | backend config.go → BFSAggregation | `BFS_AGGREGATION` | — | + | ok |
| `backend.recommendation.bfs_normalize` | backend.json | backend config.go → BFSNormalize | `BFS_NORMALIZE` | — | + | ok |
| `backend.recommendation.keyword_similarity_method` | backend.json | backend config.go → RecommendationKeywordSimilarityMethod | `RECOMMENDATION_KEYWORD_SIMILARITY_METHOD` | — | — | ok |
| `backend.recommendation.keyword_tversky_alpha` | backend.json | backend config.go → RecommendationKeywordTverskyAlpha | `RECOMMENDATION_KEYWORD_TVERSKY_ALPHA` | — | — | ok |
| `backend.recommendation.keyword_tversky_beta` | backend.json | backend config.go → RecommendationKeywordTverskyBeta | `RECOMMENDATION_KEYWORD_TVERSKY_BETA` | — | — | ok |
| `backend.recommendation.gamma_link_min_score` | backend.json | backend config.go → resolveGammaLinkMinScore | — | — | — | ok |
| `backend.pagination.default_limit` | backend.json | backend config.go → PaginationDefaultLimit | `PAGINATION_DEFAULT_LIMIT` | — | + | ok |
| `backend.pagination.max_limit` | backend.json | backend config.go → PaginationMaxLimit | `PAGINATION_MAX_LIMIT` | — | + | ok |
| `backend.graph.load_depth` | backend.json | backend config.go → GraphLoadDepth | `GRAPH_LOAD_DEPTH` | — | + | ok |
| `backend.graph.default_limit` | backend.json | backend config.go → GraphDefaultLimit | `GRAPH_DEFAULT_LIMIT` | — | + | ok |
| `backend.graph.max_limit` | backend.json | backend config.go → GraphMaxLimit | `GRAPH_MAX_LIMIT` | — | + | ok |
| `backend.graph.link_default_limit` | backend.json | backend config.go → GraphLinkDefaultLimit | `GRAPH_LINK_DEFAULT_LIMIT` | — | + | ok |
| `backend.graph.link_max_limit` | backend.json | backend config.go → GraphLinkMaxLimit | `GRAPH_LINK_MAX_LIMIT` | — | + | ok |
| `backend.embedding.similarity_limit` | backend.json | backend config.go → EmbeddingSimilarityLimit | `EMBEDDING_SIMILARITY_LIMIT` | — | + | ok |
| `backend.asynq.concurrency` | backend.json | backend config.go → AsynqConcurrency | `ASYNQ_CONCURRENCY` | — | + | ok |
| `backend.asynq.queue_default` | backend.json | backend config.go → AsynqQueueDefault | `ASYNQ_QUEUE_DEFAULT` | — | + | ok |
| `backend.asynq.queue_max_len` | backend.json | backend config.go → AsynqQueueMaxLen | `ASYNQ_QUEUE_MAX_LEN` | — | + | ok |
| `backend.redis.flush_on_startup` | backend.json | backend config.go → RedisFlushOnStartup | `REDIS_FLUSH_ON_STARTUP` | compose.deploy, compose.personal, compose.test, compose | — | ok |
| `backend.outbox.relay_interval_ms` | backend.json | backend config.go → OutboxRelayInterval | `OUTBOX_RELAY_INTERVAL_MS` | — | + | ok |
| `backend.outbox.batch_size` | backend.json | backend config.go → OutboxBatchSize | `OUTBOX_BATCH_SIZE` | — | + | ok |
| `backend.outbox.sent_retention_days` | backend.json | backend config.go → OutboxSentRetentionDays | `OUTBOX_SENT_RETENTION_DAYS` | — | + | ok |
| `backend.auth.jwt_secret` | backend.json | backend config.go → JWTSecret | `JWT_SECRET` | compose.test | + | ok |
| `backend.auth.jwt_access_ttl_seconds` | backend.json | backend config.go → JWTAccessTTL | `JWT_ACCESS_TTL_SECONDS` | — | + | ok |
| `backend.auth.jwt_refresh_ttl_seconds` | backend.json | backend config.go → JWTRefreshTTL | `JWT_REFRESH_TTL_SECONDS` | — | + | ok |
| `backend.auth.argon2_time` | backend.json | backend config.go → Argon2Time | `ARGON2_TIME` | — | — | ok |
| `backend.auth.argon2_memory` | backend.json | backend config.go → Argon2Memory | `ARGON2_MEMORY` | — | — | ok |
| `backend.auth.argon2_threads` | backend.json | backend config.go → Argon2Threads | `ARGON2_THREADS` | — | — | ok |
| `backend.auth.api_key_enabled` | backend.json | backend config.go → APIKeyEnabled | `API_KEY_ENABLED` | — | + | ok |
| `backend.auth.static_api_key` | backend.json | backend config.go → StaticAPIKey | `STATIC_API_KEY` | — | + | ok |
| `backend.auth.skip_auth` | backend.json | backend config.go → SkipAuth | `SKIP_AUTH` | compose.deploy, compose.personal, compose.test, compose | + | ok |
| `backend.auth.yandex_client_id` | backend.json | backend config.go → YandexClientID | `YANDEX_CLIENT_ID` | compose.deploy, compose.personal, compose.test, compose | + | ok |
| `backend.auth.yandex_client_secret` | backend.json | backend config.go → YandexClientSecret | `YANDEX_CLIENT_SECRET` | compose.deploy, compose.personal, compose.test, compose | + | ok |
| `backend.auth.pkce_enabled` | backend.json | backend config.go → PKCEEnabled | `PKCE_ENABLED` | — | + | ok |
| `backend.auth.pkce_code_challenge_length` | backend.json | backend config.go → PKCECodeChallengeLength | `PKCE_CODE_CHALLENGE_LENGTH` | — | — | ok |
| `backend.auth.smtp_host` | backend.json | backend config.go → SMTPHost | `SMTP_HOST` | — | + | ok |
| `backend.auth.smtp_port` | backend.json | backend config.go → SMTPPort | `SMTP_PORT` | — | + | ok |
| `backend.auth.smtp_user` | backend.json | backend config.go → SMTPUser | `SMTP_USER` | — | + | ok |
| `backend.auth.smtp_password` | backend.json | backend config.go → SMTPPassword | `SMTP_PASSWORD` | — | + | ok |
| `backend.auth.smtp_from` | backend.json | backend config.go → SMTPFrom | `SMTP_FROM` | — | — | ok |
| `backend.auth.password_reset_ttl_seconds` | backend.json | backend config.go → PasswordResetTTL | `PASSWORD_RESET_TTL_SECONDS` | — | — | ok |
| `backend.auth.password_policy_min_length` | backend.json | backend config.go → PasswordPolicyMinLength | `PASSWORD_POLICY_MIN_LENGTH` | — | — | ok |
| `backend.auth.password_policy_require_upper` | backend.json | backend config.go → PasswordPolicyRequireUpper | `PASSWORD_POLICY_REQUIRE_UPPER` | — | — | ok |
| `backend.auth.password_policy_require_lower` | backend.json | backend config.go → PasswordPolicyRequireLower | `PASSWORD_POLICY_REQUIRE_LOWER` | — | — | ok |
| `backend.auth.password_policy_require_digit` | backend.json | backend config.go → PasswordPolicyRequireDigit | `PASSWORD_POLICY_REQUIRE_DIGIT` | — | — | ok |
| `backend.auth.password_policy_require_special` | backend.json | backend config.go → PasswordPolicyRequireSpecial | `PASSWORD_POLICY_REQUIRE_SPECIAL` | — | — | ok |
| `backup.enabled` | backup.json | backend config.go → BackupEnabled | `BACKUP_ENABLED` | compose.deploy, compose.personal, compose.test, compose | — | ok |
| `backup.local_path` | backup.json | backend config.go → BackupLocalPath | `BACKUP_LOCAL_PATH` | compose.personal | + | ok |
| `backup.cloud.enabled` | backup.json | backend config.go → BackupCloudEnabled | `BACKUP_CLOUD_ENABLED` | compose.deploy, compose.personal, compose.test, compose | + | ok |
| `backup.cloud.provider` | backup.json | backend config.go → BackupCloudProvider | `BACKUP_CLOUD_PROVIDER` | — | + | ok |
| `backup.cloud.yandex.oauth_token` | backup.json | backend config.go → BackupYandexOAuthToken | `BACKUP_YANDEX_OAUTH_TOKEN` | compose.deploy, compose.personal, compose.test, compose | + | ok |
| `backup.cloud.yandex.backup_folder` | backup.json | backend config.go → BackupYandexFolder | `BACKUP_YANDEX_FOLDER` | compose.personal | + | ok |
| `backup.cloud.yandex.max_backups` | backup.json | backend config.go → BackupYandexMaxBackups | `BACKUP_YANDEX_MAX_BACKUPS` | — | + | ok |
| `backup.schedule` | backup.json | backend config.go → BackupSchedule | `BACKUP_SCHEDULE` | — | + | ok |
| `backup.retention_days` | backup.json | backend config.go → BackupRetentionDays | `BACKUP_RETENTION_DAYS` | — | + | ok |
| `frontend.graph.label_hub_count` | frontend.json | src/entities/graph-canvas/lib/labels.test.ts, src/entities/graph-canvas/lib/labels.ts, src/features/graph-3d/lib/labels.test.ts | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.dependency_highlight_depth` | frontend.json | src/features/graph-3d/lib/engine.ts, src/widgets/graph-canvas/GraphCanvas.svelte | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.style` | frontend.json | src/entities/graph-canvas/lib/light/style.ts | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.recommendations_on_hover` | frontend.json | src/widgets/graph-canvas/GraphCanvas.svelte | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.ambient_max_nodes` | frontend.json | src/entities/graph-canvas/lib/light/style.ts, src/widgets/graph-canvas/GraphCanvas.svelte | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.2d.shadows_threshold` | frontend.json | src/entities/graph-canvas/lib/node-renderers.ts, src/entities/graph-canvas/lib/node-renderers.ts (через graphConfig2D), src/entities/graph-canvas/lib/renderer-orchestrator.ts, src/entities/graph-canvas/lib/renderer-orchestrator.ts (через graphConfig2D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.2d.animated_links_threshold` | frontend.json | src/entities/graph-canvas/lib/link-renderers.ts, src/entities/graph-canvas/lib/link-renderers.ts (через graphConfig2D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.2d.gravity_nodes_threshold` | frontend.json | src/entities/graph-canvas/lib/gravity-system.test.ts, src/entities/graph-canvas/lib/gravity-system.test.ts (через graphConfig2D), src/entities/graph-canvas/lib/gravity-system.ts, src/entities/graph-canvas/lib/gravity-system.ts (через graphConfig2D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.2d.gravity_max_distance` | frontend.json | src/entities/graph-canvas/lib/gravity-system.ts, src/entities/graph-canvas/lib/gravity-system.ts (через graphConfig2D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.2d.hover_delay_ms` | frontend.json | src/features/graph-interaction/event-bridge.ts, src/features/graph-interaction/event-bridge.ts (через graphConfig2D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.2d.visual_fx_threshold` | frontend.json | src/entities/graph-canvas/lib/background.ts, src/entities/graph-canvas/lib/background.ts (через graphConfig2D), src/entities/graph-canvas/lib/node-renderers.ts, src/entities/graph-canvas/lib/node-renderers.ts (через graphConfig2D), src/entities/graph-canvas/lib/particle-system.test.ts, src/entities/graph-canvas/lib/particle-system.test.ts (через graphConfig2D), src/entities/graph-canvas/lib/particle-system.ts, src/entities/graph-canvas/lib/particle-system.ts (через graphConfig2D), src/entities/graph-canvas/lib/renderer.test.ts, src/entities/graph-canvas/lib/renderer.test.ts (через graphConfig2D), src/shared/lib/graph/glow-intensity.ts, src/shared/lib/graph/glow-intensity.ts (через graphConfig2D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.2d.idle_fps` | frontend.json | src/widgets/graph-canvas/GraphCanvas.svelte, src/widgets/graph-canvas/GraphCanvas.svelte (через graphConfig2D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.2d.lod_simplify_zoom` | frontend.json | src/entities/graph-canvas/lib/renderer-orchestrator.ts, src/entities/graph-canvas/lib/renderer-orchestrator.ts (через graphConfig2D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.2d.fog.enabled` | frontend.json | src/entities/graph-canvas/lib/fog.ts, src/entities/graph-canvas/lib/fog.ts (через graphConfig2D), src/entities/graph-canvas/lib/particle-system.test.ts (через graphConfig2D), src/entities/graph-canvas/lib/particle-system.ts (через graphConfig2D), src/features/graph-canvas/fog-state.svelte.test.ts, src/features/graph-canvas/fog-state.svelte.test.ts (через graphConfig2D), src/features/graph-canvas/fog-state.svelte.ts, src/features/graph-canvas/fog-state.svelte.ts (через graphConfig2D), src/widgets/graph-canvas/GraphCanvas.svelte (через graphConfig2D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.2d.fog.atmospheric` | frontend.json | src/entities/graph-canvas/lib/fog.ts, src/entities/graph-canvas/lib/fog.ts (через graphConfig2D), src/features/graph-canvas/fog-state.svelte.test.ts, src/features/graph-canvas/fog-state.svelte.test.ts (через graphConfig2D), src/features/graph-canvas/fog-state.svelte.ts, src/features/graph-canvas/fog-state.svelte.ts (через graphConfig2D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.2d.fog.adaptive` | frontend.json | src/entities/graph-canvas/lib/fog.ts, src/entities/graph-canvas/lib/fog.ts (через graphConfig2D), src/features/graph-canvas/fog-state.svelte.test.ts, src/features/graph-canvas/fog-state.svelte.test.ts (через graphConfig2D), src/features/graph-canvas/fog-state.svelte.ts, src/features/graph-canvas/fog-state.svelte.ts (через graphConfig2D), src/widgets/graph-canvas/GraphCanvas.svelte (через graphConfig2D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.2d.fog.radius_min` | frontend.json | src/entities/graph-canvas/lib/fog.ts, src/features/graph-canvas/fog-state.svelte.test.ts, src/features/graph-canvas/fog-state.svelte.test.ts (через graphConfig2D), src/features/graph-canvas/fog-state.svelte.ts, src/features/graph-canvas/fog-state.svelte.ts (через graphConfig2D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.2d.fog.radius_max` | frontend.json | src/entities/graph-canvas/lib/fog.ts, src/entities/graph-canvas/lib/fog.ts (через graphConfig2D), src/features/graph-canvas/fog-state.svelte.test.ts, src/features/graph-canvas/fog-state.svelte.test.ts (через graphConfig2D), src/features/graph-canvas/fog-state.svelte.ts, src/features/graph-canvas/fog-state.svelte.ts (через graphConfig2D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.2d.fog.fps_low` | frontend.json | src/entities/graph-canvas/lib/fog.ts, src/features/graph-canvas/fog-state.svelte.test.ts, src/features/graph-canvas/fog-state.svelte.ts, src/features/graph-canvas/fog-state.svelte.ts (через graphConfig2D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.2d.fog.fps_high` | frontend.json | src/entities/graph-canvas/lib/fog.ts, src/features/graph-canvas/fog-state.svelte.test.ts, src/features/graph-canvas/fog-state.svelte.ts, src/features/graph-canvas/fog-state.svelte.ts (через graphConfig2D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.2d.fog.warning_threshold` | frontend.json | src/entities/graph-canvas/lib/fog.ts, src/features/graph-canvas/fog-state.svelte.test.ts, src/features/graph-canvas/fog-state.svelte.ts, src/features/graph-canvas/fog-state.svelte.ts (через graphConfig2D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.2d.fog.transition_ms` | frontend.json | src/entities/graph-canvas/lib/fog.ts, src/features/graph-canvas/fog-state.svelte.test.ts, src/features/graph-canvas/fog-state.svelte.ts, src/features/graph-canvas/fog-state.svelte.ts (через graphConfig2D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.2d.fog.color` | frontend.json | src/entities/graph-canvas/lib/fog.ts, src/entities/graph-canvas/lib/fog.ts (через graphConfig2D), src/entities/graph-canvas/lib/node-renderers.ts (через graphConfig2D), src/entities/graph-canvas/lib/particle-system.ts (через graphConfig2D), src/entities/graph-canvas/lib/renderer-orchestrator.ts (через graphConfig2D), src/entities/graph-canvas/lib/renderer.test.ts (через graphConfig2D), src/features/graph-canvas/fog-state.svelte.test.ts, src/features/graph-canvas/fog-state.svelte.ts, src/features/graph-canvas/fog-state.svelte.ts (через graphConfig2D), src/widgets/graph-canvas/GraphCanvas.svelte (через graphConfig2D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.2d.fog.edge_feather` | frontend.json | src/entities/graph-canvas/lib/fog.ts, src/entities/graph-canvas/lib/fog.ts (через graphConfig2D), src/features/graph-canvas/fog-state.svelte.test.ts, src/features/graph-canvas/fog-state.svelte.ts, src/features/graph-canvas/fog-state.svelte.ts (через graphConfig2D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.3d.enabled` | frontend.json | src/features/graph-ui/GraphTopBar.spec.ts, src/features/graph-ui/GraphTopBar.svelte | нет (вшивается при сборке) | — | + | ok |
| `frontend.graph.3d.max_nodes` | frontend.json | src/features/graph-3d/lib/engine.ts, src/features/graph-3d/lib/engine.ts (через graphConfig3D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.3d.layout_provider` | frontend.json | src/features/graph-3d/config.ts, src/features/graph-3d/config.ts (через graphConfig3D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.3d.fog.presets.birth.density_initial` | frontend.json | src/features/graph-3d/lib/engine.ts, src/features/graph-3d/lib/scene.ts | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.3d.fog.presets.birth.density_final` | frontend.json | src/features/graph-3d/lib/engine.ts, src/features/graph-3d/lib/scene.ts | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.3d.fog.presets.nebula.density_initial` | frontend.json | src/features/graph-3d/lib/engine.ts, src/features/graph-3d/lib/scene.ts | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.3d.fog.presets.nebula.density_final` | frontend.json | src/features/graph-3d/lib/engine.ts, src/features/graph-3d/lib/scene.ts | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.3d.fog.presets.deep-space.density_initial` | frontend.json | src/features/graph-3d/lib/engine.ts, src/features/graph-3d/lib/scene.ts | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.3d.fog.presets.deep-space.density_final` | frontend.json | src/features/graph-3d/lib/engine.ts, src/features/graph-3d/lib/scene.ts | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.3d.fog.default_preset` | frontend.json | src/features/graph-3d/lib/engine.ts, src/features/graph-3d/lib/engine.ts (через graphConfig3D), src/features/graph-3d/lib/scene.ts, src/features/graph-3d/lib/scene.ts (через graphConfig3D) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.3d.performance.fps_threshold_low` | frontend.json | src/features/graph-3d/lib/engine.ts (через graphConfig3D), src/features/graph-3d/lib/engine.ts (через graphPerformanceConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.3d.performance.fps_threshold_high` | frontend.json | src/features/graph-3d/lib/engine.ts (через graphConfig3D), src/features/graph-3d/lib/engine.ts (через graphPerformanceConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.3d.performance.low_fps_sample_count` | frontend.json | src/features/graph-3d/lib/engine.ts (через graphConfig3D), src/features/graph-3d/lib/engine.ts (через graphPerformanceConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.3d.performance.starfield_counts.high` | frontend.json | src/features/graph-3d/lib/engine.ts (через graphConfig3D), src/features/graph-3d/lib/engine.ts (через graphPerformanceConfig), src/features/graph-3d/lib/scene.ts (через graphConfig3D), src/features/graph-3d/lib/scene.ts (через graphPerformanceConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.3d.performance.starfield_counts.medium` | frontend.json | src/features/graph-3d/lib/engine.ts (через graphConfig3D), src/features/graph-3d/lib/engine.ts (через graphPerformanceConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.3d.performance.starfield_counts.low` | frontend.json | src/features/graph-3d/lib/engine.ts (через graphConfig3D), src/features/graph-3d/lib/engine.ts (через graphPerformanceConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.reality_rift.core_color` | frontend.json | src/app.d.ts (через anomalyConfig), src/shared/lib/graph/renderer/anomalies/helpers.ts, src/shared/lib/graph/renderer/anomalies/reality-rift.ts (через anomalyConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.reality_rift.glow_color` | frontend.json | src/app.d.ts (через anomalyConfig), src/shared/lib/graph/renderer/anomalies/helpers.ts, src/shared/lib/graph/renderer/anomalies/reality-rift.ts (через anomalyConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.reality_rift.crack_count_min` | frontend.json | src/shared/lib/graph/renderer/anomalies/helpers.ts, src/shared/lib/graph/renderer/anomalies/helpers.ts (через anomalyConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.reality_rift.crack_count_max` | frontend.json | src/shared/lib/graph/renderer/anomalies/helpers.ts, src/shared/lib/graph/renderer/anomalies/helpers.ts (через anomalyConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.reality_rift.deform_amount_min` | frontend.json | src/shared/lib/graph/renderer/anomalies/helpers.ts, src/shared/lib/graph/renderer/anomalies/helpers.ts (через anomalyConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.reality_rift.deform_amount_max` | frontend.json | src/shared/lib/graph/renderer/anomalies/helpers.ts, src/shared/lib/graph/renderer/anomalies/helpers.ts (через anomalyConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.chromatic_maw.tentacle_count_min` | frontend.json | src/shared/lib/graph/renderer/anomalies/helpers.ts, src/shared/lib/graph/renderer/anomalies/helpers.ts (через anomalyConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.chromatic_maw.tentacle_count_max` | frontend.json | src/shared/lib/graph/renderer/anomalies/helpers.ts, src/shared/lib/graph/renderer/anomalies/helpers.ts (через anomalyConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.chromatic_maw.hue_shift_base` | frontend.json | src/shared/lib/graph/renderer/anomalies/helpers.ts | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.chromatic_maw.hue_shift_range` | frontend.json | src/shared/lib/graph/renderer/anomalies/helpers.ts, src/shared/lib/graph/renderer/anomalies/helpers.ts (через anomalyConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.void_whisper.particle_count_min` | frontend.json | src/shared/lib/graph/renderer/anomalies/helpers.ts, src/shared/lib/graph/renderer/anomalies/helpers.ts (через anomalyConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.void_whisper.particle_count_max` | frontend.json | src/shared/lib/graph/renderer/anomalies/helpers.ts, src/shared/lib/graph/renderer/anomalies/helpers.ts (через anomalyConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.void_whisper.hue_shift_base` | frontend.json | src/shared/lib/graph/renderer/anomalies/helpers.ts | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.void_whisper.hue_shift_range` | frontend.json | src/shared/lib/graph/renderer/anomalies/helpers.ts, src/shared/lib/graph/renderer/anomalies/helpers.ts (через anomalyConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.void_whisper.connection_distance_threshold` | frontend.json | src/shared/lib/graph/renderer/anomalies/helpers.ts | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.cosmic_abomination.particle_count_min` | frontend.json | src/shared/lib/graph/renderer/anomalies/helpers.ts, src/shared/lib/graph/renderer/anomalies/helpers.ts (через anomalyConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.cosmic_abomination.particle_count_max` | frontend.json | src/shared/lib/graph/renderer/anomalies/helpers.ts, src/shared/lib/graph/renderer/anomalies/helpers.ts (через anomalyConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.cosmic_abomination.tentacle_count_min` | frontend.json | src/shared/lib/graph/renderer/anomalies/helpers.ts, src/shared/lib/graph/renderer/anomalies/helpers.ts (через anomalyConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.cosmic_abomination.tentacle_count_max` | frontend.json | src/shared/lib/graph/renderer/anomalies/helpers.ts, src/shared/lib/graph/renderer/anomalies/helpers.ts (через anomalyConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.cosmic_abomination.crack_count_min` | frontend.json | src/shared/lib/graph/renderer/anomalies/helpers.ts, src/shared/lib/graph/renderer/anomalies/helpers.ts (через anomalyConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.graph.anomaly.cosmic_abomination.crack_count_max` | frontend.json | src/shared/lib/graph/renderer/anomalies/helpers.ts, src/shared/lib/graph/renderer/anomalies/helpers.ts (через anomalyConfig) | нет (вшивается при сборке) | — | — | ok |
| `frontend.achievements.poll_interval_ms` | frontend.json | src/entities/achievement/model/store.svelte.test.ts, src/entities/achievement/model/store.svelte.ts | нет (вшивается при сборке) | — | + | ok |
| `graph_service.grpc_port` | graph_service.json | graph-service config.go → GRPCPort | `GRPC_PORT` | compose.deploy, compose.personal, compose.test, compose | + | ok |
| `graph_service.http_port` | graph_service.json | graph-service config.go → HTTPPort | `HTTP_PORT` | compose.deploy, compose.personal, compose.test, compose | + | ok |
| `graph_service.full_limit` | graph_service.json | graph-service config.go → FullLimit | `GRAPH_FULL_LIMIT` | compose.deploy, compose.test, compose | + | ok |
| `graph_service.default_depth` | graph_service.json | graph-service config.go → DefaultDepth | — | — | + | ok |
| `graph_service.event_channel` | graph_service.json | graph-service config.go → EventChannel | `EVENT_CHANNEL` | compose.deploy, compose.personal, compose.test, compose | + | ok |
| `graph_service.cache.note_layout_ttl_seconds` | graph_service.json | graph-service config.go → NoteLayoutTTL | `CACHE_NOTE_TTL_SECONDS` | — | + | ok |
| `graph_service.cache.full_layout_ttl_seconds` | graph_service.json | graph-service config.go → FullLayoutTTL | `CACHE_FULL_TTL_SECONDS` | — | + | ok |
| `graph_service.cache.delta_ttl_seconds` | graph_service.json | graph-service config.go → DeltaTTL | `CACHE_DELTA_TTL_SECONDS` | — | + | ok |
| `graph_service.cache.snapshot_ttl_seconds` | graph_service.json | graph-service config.go → SnapshotTTL | `CACHE_SNAPSHOT_TTL_SECONDS` | — | + | ok |
| `graph_service.layout.2d_radius` | graph_service.json | graph-service config.go → Layout2DRadius | — | — | — | ok |
| `graph_service.layout.3d_radius` | graph_service.json | graph-service config.go → Layout3DRadius | — | — | — | ok |
| `graph_service.layout.3d_z_step` | graph_service.json | graph-service config.go → Layout3DZStep | — | — | — | ok |
| `graph_service.layout.default_node_size` | graph_service.json | graph-service config.go → DefaultNodeSize | — | — | — | ok |
| `graph_service.stream_chunk_size` | graph_service.json | graph-service config.go → StreamChunkSize | — | — | + | ok |
| `graph_service.event_tracking_ttl_hours` | graph_service.json | graph-service config.go → Load | — | — | — | ok |
| `graph_service.unprocessed_event_check_interval_minutes` | graph_service.json | graph-service config.go → Load | — | — | — | ok |
| `mongodb.url` | mongodb.json | backend config.go → MongoDBURL | `MONGO_URL` | compose.deploy, compose.personal, compose.test, compose | + | ok |
| `mongodb.database` | mongodb.json | backend config.go → MongoDBDatabase | `MONGO_DATABASE` | compose.deploy, compose.personal, compose.test, compose | + | ok |
| `nlp.model_name` | nlp.json | nlp-service app/config.py → NLP_MODEL_NAME | `NLP_MODEL_NAME` | compose.deploy, compose.personal, compose.test, compose | + | ok |
| `nlp.max_text_length` | nlp.json | nlp-service app/config.py → NLP_MAX_TEXT_LENGTH | `NLP_MAX_TEXT_LENGTH` | — | + | ok |
| `nlp.hf_home` | nlp.json | nlp-service app/config.py → HF_HOME | `HF_HOME` | compose.deploy, compose.personal, compose.test, compose | + | ok |
| `nlp.hf_hub_disable_telemetry` | nlp.json | nlp-service app/config.py → HF_HUB_DISABLE_TELEMETRY | `HF_HUB_DISABLE_TELEMETRY` | compose.deploy, compose.personal, compose.test, compose | + | ok |
| `nlp.hf_hub_offline` | nlp.json | nlp-service app/config.py → HF_HUB_OFFLINE | `HF_HUB_OFFLINE` | compose.deploy, compose.personal, compose.test, compose | + | ok |
| `nlp.pipeline.enabled` | nlp.json | backend config.go → NLPPipelineEnabled | `NLP_PIPELINE_ENABLED` | compose.deploy, compose.personal, compose.test, compose | + | ok |
| `nlp.quality.enabled` | nlp.json | backend config.go → NLPQualityEnabled | `NLP_QUALITY_ENABLED` | compose.deploy, compose.personal, compose.test, compose | + | ok |
| `nlp.quality.collection_prose_share` | nlp.json | backend config.go → NLPQualityCollectionProseShare | `NLP_QUALITY_COLLECTION_PROSE_SHARE` | — | + | ok |
| `nlp.quality.collection_min_links` | nlp.json | backend config.go → NLPQualityCollectionMinLinks | `NLP_QUALITY_COLLECTION_MIN_LINKS` | — | + | ok |
| `nlp.quality.sentence_min_words` | nlp.json | backend config.go → NLPQualitySentenceMinWords | `NLP_QUALITY_SENTENCE_MIN_WORDS` | — | + | ok |
| `nlp.quality.fragment_max_words` | nlp.json | backend config.go → NLPQualityFragmentMaxWords | `NLP_QUALITY_FRAGMENT_MAX_WORDS` | — | + | ok |
| `nlp.quality.mojibake_share` | nlp.json | backend config.go → NLPQualityMojibakeShare | `NLP_QUALITY_MOJIBAKE_SHARE` | — | + | ok |
| `nlp.quality.legacy_truncated_runes` | nlp.json | backend config.go → NLPQualityLegacyTruncatedRunes | `NLP_QUALITY_LEGACY_TRUNCATED_RUNES` | — | + | ok |
