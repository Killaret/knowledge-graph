# ADR 015: Galactic Lexicon and Achievements System

## Status
Accepted (with polling implementation)

## Context
Knowledge Graph uses a space/cosmic theme throughout the UI:
- Notes displayed as celestial bodies
- Connections shown as orbits/constellations
- Types represented by different celestial body types (stars, planets, etc.)
- User achievements framed as space exploration milestones

As the system grew, several challenges emerged:
- **Inconsistent terminology**: Different terms used for same concepts across UI
- **Hard-coded strings**: Internationalization (i18n) support difficult
- **Achievement notifications**: Users not immediately notified of new achievements
- **Gamification complexity**: Achievement logic scattered across codebase

### Current State Analysis
The system has:
- Basic achievement system with database-backed definitions
- Mixed terminology (note/celestial body, link/orbit, type/celestial type)
- **Polling-based achievement checks** (7000ms interval) - SSE not yet implemented
- Galactic lexicon message system (standard/galactic modes)

Current challenges:
- **UX inconsistency**: Users see different terms for same concepts
- **Maintenance burden**: Adding new languages requires finding all hardcoded strings
- **Polling inefficiency**: 7-second polling interval instead of real-time SSE
- **Scattered logic**: Achievement rules mixed with business logic

## Problem Statement
How do we create a consistent, extensible terminology system with proper internationalization support and real-time achievement notifications?

## Decision Drivers
- **Consistency**: Uniform terminology across entire application
- **Internationalization**: Support for multiple languages
- **Extensibility**: Easy to add new terms/languages/achievements
- **Real-time feedback**: Users should see achievements immediately
- **Performance**: Achievement checks should not impact system performance
- **Maintainability**: Centralized terminology and achievement definitions

## Considered Options

### Option 1: Continue with Hardcoded Strings
Keep current approach with hardcoded strings throughout the codebase.

**Pros:**
- ✅ Simplest implementation
- ✅ No additional infrastructure
- ✅ Fast (no lookup overhead)

**Cons:**
- ❌ No i18n support
- ❌ Terminology inconsistencies
- ❌ Difficult to maintain
- ❌ Adding languages requires code changes
- ❌ Cannot dynamically update terminology

### Option 2: JSON-based Lexicon Files
Store terminology in JSON files, load at runtime.

**Pros:**
- ✅ Easy to edit (text files)
- ✅ Supports multiple languages (separate files per language)
- ✅ No database dependency
- ✅ Can be version-controlled

**Cons:**
- ❌ Requires application restart to update
- ❌ No dynamic updates
- ❌ File I/O overhead
- ❌ Synchronization issues across multiple instances
- ❌ No UI for non-technical users

### Option 3: Database-stored Lexicon
Store terminology and achievements in database tables.

**Pros:**
- ✅ Dynamic updates (no restart required)
- ✅ Centralized management
- ✅ Can build UI for editing
- ✅ Syncs across all instances
- ✅ Queryable and searchable
- ✅ Can track changes/audit trail

**Cons:**
- ❌ Additional database queries (performance impact)
- ❌ More complex infrastructure
- ❌ Need caching to avoid performance issues
- ❌ Database dependency for UI strings
- ❌ Schema changes required for new fields

### Option 4: Frontend i18n Library (i18next)
Use standard frontend i18n library for terminology, keep achievements in database.

**Pros:**
- ✅ Industry standard for frontend i18n
- ✅ Great tooling and ecosystem
- ✅ Supports pluralization, formatting, etc.
- ✅ Lazy loading of language files
- ✅ Can use JSON files (easy to edit)

**Cons:**
- ❌ Only frontend (backend still has strings)
- ❌ Need to manage terminology in two places
- ❌ Achievement notifications still need solution
- ❌ No centralized terminology management

### Option 5: Hybrid Approach - Frontend i18n + Database Achievements + SSE Notifications
Combine multiple approaches:
- Frontend: i18next for UI terminology
- Backend: Database for achievements with SSE for real-time notifications

**Pros:**
- ✅ Best of both worlds (frontend i18n + backend flexibility)
- ✅ Real-time achievement notifications via SSE
- ✅ Industry-standard frontend i18n
- ✅ Database for dynamic achievement rules
- ✅ Separation of concerns (UI vs game logic)

**Cons:**
- ❌ More complex (multiple systems to manage)
- ❌ Terminology still in two places (frontend JSON + backend)
- ❌ SSE requires connection management
- ❌ Need to handle SSE reconnection

## Decision
**Chosen Approach: Option 5 - Hybrid Approach (Partial Implementation)**

We combine frontend terminology system with database-backed achievements. **Current implementation uses polling** instead of SSE for real-time notifications.

### Current Implementation Status

#### ✅ Implemented
- **Database schema**: achievements and user_achievements tables
- **Achievement engine**: Rules engine with count/streak conditions
- **HTTP handlers**: GET /api/v1/achievements, GET /api/v1/users/me/achievements
- **Frontend polling**: 7-second interval via PreloadService
- **Galactic lexicon**: Message system with standard/galactic modes
- **User settings**: galactic_mode, show_achievement_notifications, preferred_language

#### ⏳ TODO (Not Yet Implemented)
- **SSE notifications**: Server-Sent Events for real-time achievement pushes
- **i18next integration**: Frontend internationalization library
- **Redis streak tracking**: Real-time streak counters
- **SSE endpoint**: /api/v1/achievements/stream

### Architecture (Current - Polling)

```
┌─────────────────────────────────────────────────────────────┐
│                      Frontend (Svelte)                       │
│  ┌─────────────────────────────────────────────────────┐   │
│  │          PreloadService (Polling)                    │   │
│  │  ┌───────────────────────────────────────────────┐  │   │
│  │  │  setInterval(7000ms) → GET /api/v1/           │  │   │
│  │  │  users/me/achievements                        │  │   │
│  │  │  - Check for new achievements                 │  │   │
│  │  │  - Show toast if unlocked                     │  │   │
│  │  │  - Mark as seen                               │  │   │
│  │  └───────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────┘   │
│                           │                                 │
│                           │ HTTP Polling                    │
│                           │ Every 7 seconds                 │
└───────────────────────────┼─────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                     Backend (Go)                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │            Achievement Engine                        │   │
│  │  ┌──────────────┐  ┌──────────────┐                 │   │
│  │  │ Rules Engine │  │ Progress     │                 │   │
│  │  └──────┬───────┘  └──────┬───────┘                 │   │
│  │         │                 │                          │   │
│  │         │ Check           │ Track                    │   │
│  │         ▼                 ▼                          │   │
│  │  ┌──────────────┐  ┌──────────────┐                 │   │
│  │  │ PostgreSQL   │  │ Cache       │                 │   │
│  │  │ achievements │  │ (Optional)  │                 │   │
│  │  │ table        │  │             │                 │   │
│  │  └──────────────┘  └──────────────┘                 │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### TODO: SSE Implementation (Future)

When implementing SSE, replace polling with: (Current - Polling)

```
┌─────────────────────────────────────────────────────────────┐
│                      Frontend (Svelte)                       │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              i18next (UI Terminology)               │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐          │   │
│  │  │  en.json │  │  ru.json │  │  de.json │  ...     │   │
│  │  └──────────┘  └──────────┘  └──────────┘          │   │
│  └─────────────────────────────────────────────────────┘   │
│                           │                                 │
│                           │ SSE Connection                  │
│                           │ /api/v1/achievements/stream      │
└───────────────────────────┼─────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                     Backend (Go)                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │            Achievement Engine                        │   │
│  │  ┌──────────────┐  ┌──────────────┐                 │   │
│  │  │ Rules Engine │  │ Progress     │                 │   │
│  │  └──────┬───────┘  └──────┬───────┘                 │   │
│  │         │                 │                          │   │
│  │         │ Check           │ Track                    │   │
│  │         ▼                 ▼                          │   │
│  │  ┌──────────────┐  ┌──────────────┐                 │   │
│  │  │ PostgreSQL   │  │ Redis       │                 │   │
│  │  │ achievements │  │ Cache       │                 │   │
│  │  │ table        │  │             │                 │   │
│  │  └──────────────┘  └──────────────┘                 │   │
│  └─────────────────────────────────────────────────────┘   │
│                           │                                 │
│                           │ SSE Stream                      │
│                           │ Push achievements              │
└───────────────────────────┼─────────────────────────────────┘
                            │
                    ┌───────▼────────┐
                    │ User Browser  │
                    │ Show toast    │
                    │ Play sound    │
                    └───────────────┘
```

### TODO: SSE Implementation (Future)

When implementing SSE, replace polling with:

```
┌─────────────────────────────────────────────────────────────┐
│                      Frontend (Svelte)                       │
│  ┌─────────────────────────────────────────────────────┐   │
│  │         EventSource Connection                       │   │
│  │  ┌───────────────────────────────────────────────┐  │   │
│  │  │  SSE: /api/v1/achievements/stream              │  │   │
│  │  │  - Auto-reconnect on disconnect               │  │   │
│  │  │  - Listen for 'achievement' events            │  │   │
│  │  │  - Show toast immediately                     │  │   │
│  │  └───────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────┘   │
│                           │                                 │
│                           │ SSE Connection (Long-lived)     │
│                           │ Push on achievement unlocked    │
└───────────────────────────┼─────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                     Backend (Go)                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │          SSE Event Bus                               │   │
│  │  ┌───────────────────────────────────────────────┐  │   │
│  │  │  Publish achievement to user channel          │  │   │
│  │  │  eventBus.Publish("achievements:{user_id}",   │  │   │
│  │  │                   achievement)                 │  │   │
│  │  └───────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

**SSE Benefits over Polling:**
- ✅ Real-time: Instant notifications
- ✅ Efficient: No unnecessary requests
- ✅ Battery-friendly: No constant polling
- ✅ Lower server load: Push instead of pull

**Implementation Priority:** Medium - polling works, but SSE improves UX

### Implementation Details

#### 1. Frontend i18next Configuration

**Language File Structure:**
```json
// src/shared/locales/en.json
{
  "galaxy": {
    "note": "Celestial Body",
    "notes": "Celestial Bodies",
    "link": "Orbital Connection",
    "links": "Orbital Connections",
    "type": {
      "star": "Star",
      "planet": "Planet",
      "nebula": "Nebula",
      "blackhole": "Black Hole"
    }
  },
  "achievements": {
    "first_note": {
      "title": "First Discovery",
      "description": "Create your first celestial body"
    },
    "connector": {
      "title": "Constellation Builder",
      "description": "Create 10 orbital connections"
    }
  }
}
```

**Svelte Component Usage:**
```svelte
<script>
  import { t } from 'i18next';
</script>

<h1>{$t('galaxy.note')}</h1>
```

#### 2. Database Schema for Achievements

```sql
CREATE TABLE achievements (
    id UUID PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    icon VARCHAR(100),
    category VARCHAR(50),
    requirement_type VARCHAR(50), -- count, milestone, etc.
    requirement_value INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE user_achievements (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    achievement_id UUID NOT NULL,
    progress INTEGER DEFAULT 0,
    completed_at TIMESTAMP,
    notified_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (achievement_id) references achievements(id),
    UNIQUE(user_id, achievement_id)
);
```

#### 3. Achievement Engine

```go
type AchievementEngine struct {
    db          *sql.DB
    redis       *redis.Client
    eventBus    EventBus
}

func (e *AchievementEngine) CheckAchievements(ctx context.Context, userID uuid.UUID, event Event) {
    // Get applicable achievements for event type
    achievements := e.getAchievementsForEvent(event.Type)

    for _, achievement := range achievements {
        progress := e.calculateProgress(ctx, userID, achievement, event)

        // Update progress
        e.updateProgress(ctx, userID, achievement.ID, progress)

        // Check if completed
        if progress >= achievement.RequirementValue {
            if !e.isCompleted(ctx, userID, achievement.ID) {
                e.completeAchievement(ctx, userID, achievement.ID)
                e.notifyAchievement(ctx, userID, achievement)
            }
        }
    }
}
```

#### 4. SSE Endpoint for Real-time Notifications

```go
func (h *AchievementHandler) StreamAchievements(c *gin.Context) {
    userID := c.GetHeader("X-User-ID")

    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")

    // Subscribe to user's achievement channel
    ch := h.eventBus.Subscribe(fmt.Sprintf("achievements:%s", userID))

    defer h.eventBus.Unsubscribe(ch)

    // Send existing unnotified achievements
    unnotified := h.getUnnotifiedAchievements(userID)
    for _, achievement := range unnotified {
        c.SSEvent("achievement", achievement)
        h.markAsNotified(userID, achievement.ID)
    }

    // Stream new achievements
    for {
        select {
        case achievement := <-ch:
            c.SSEvent("achievement", achievement)
            h.markAsNotified(userID, achievement.ID)
        case <-c.Request.Context().Done():
            return
        }
    }
}
```

#### 5. Frontend SSE Client

```typescript
// src/shared/achievements/stream.ts
export class AchievementStream {
  private eventSource: EventSource | null = null;

  connect(userId: string) {
    this.eventSource = new EventSource(`/api/v1/achievements/stream?userId=${userId}`);

    this.eventSource.addEventListener('achievement', (event) => {
      const achievement = JSON.parse(event.data);
      this.showAchievementToast(achievement);
      this.playAchievementSound();
    });

    this.eventSource.addEventListener('error', (error) => {
      console.error('SSE error:', error);
      this.reconnect();
    });
  }

  private reconnect() {
    setTimeout(() => this.connect(this.userId), 5000);
  }
}
```

### Terminology Strategy

#### Galactic Lexicon Standardization

| Concept | Old Term | New Term (Galactic) |
|---------|----------|---------------------|
| Note | Note | Celestial Body |
| Link | Link | Orbital Connection |
| Type | Type | Celestial Type |
| Search | Search | Stellar Scan |
| Recommendations | Recommendations | Cosmic Guidance |
| Draft | Draft | Proto-Body |

#### Achievement Categories

1. **Discovery** - Creating notes, exploring features
2. **Connection** - Creating links, building constellations
3. **Knowledge** - Learning advanced features
4. **Mastery** - Advanced usage patterns

### SSE vs Polling for Achievements

#### Why SSE over Polling?

**Polling Issues:**
- ❌ Server load: Constant requests even when no updates
- ❌ Latency: Updates only seen on next poll
- ❌ Waste: Bandwidth and CPU wasted on empty responses
- ❌ Battery: Drains mobile battery on constant requests

**SSE Advantages:**
- ✅ Real-time: Updates pushed immediately
- ✅ Efficient: No unnecessary requests
- ✅ Simple: Standard HTTP, no special libraries needed
- ✅ Reconnection: Built-in browser reconnection support
- ✅ Lower load: Server only pushes when there are updates

#### SSE Implementation Considerations

**Connection Management:**
- Auto-reconnection with exponential backoff
- Last-Event-ID header for resuming from last event
- Connection timeout handling
- Multiple tab support (shared connections)

**Scalability:**
- Long-lived connections (need connection pooling)
- Load balancer configuration (sticky sessions)
- Resource limits (max connections per server)

**Fallback:**
- If SSE not supported, fall back to polling
- User can disable real-time notifications

## Consequences

### Positive Consequences
- ✅ **Consistent terminology**: Uniform galactic theme across application
- ✅ **Internationalization ready**: Easy to add new languages via JSON files
- ✅ **Real-time feedback**: Users see achievements immediately
- ✅ **Better engagement**: Gamification increases user engagement
- ✅ **Maintainable**: Centralized terminology and achievement definitions
- ✅ **Performance**: SSE is more efficient than polling
- ✅ **Extensible**: Easy to add new achievements and languages

### Negative Consequences
- ❌ **Complexity**: Multiple systems to manage (i18n + database + SSE)
- ❌ **Connection overhead**: SSE maintains long-lived connections
- ❌ **Scaling challenges**: Need to handle many SSE connections
- ❌ **Terminology sync**: Need to keep frontend JSON and backend in sync
- ❌ **Browser support**: SSE not supported in IE/Edge (but polyfills available)

### Mitigation Strategies
- **Complexity**: Clear separation of concerns, good documentation
- **Connection overhead**: Connection pooling, load balancing, monitoring
- **Scaling**: Horizontal scaling with sticky sessions, consider SSE proxy
- **Terminology sync**: Regular audits, automated checks
- **Browser support**: Polyfills for older browsers, graceful degradation

## When to Reconsider
- If SSE connection count becomes unmanageable
- If terminology sync becomes major maintenance burden
- If achievement logic becomes too complex for current approach
- If need more sophisticated real-time features (WebSocket)

## Alternatives for Future
- **WebSocket**: If need bidirectional communication
- **GraphQL Subscriptions**: If using GraphQL extensively
- **Unified terminology backend**: If need centralized terminology management
- **Push notifications**: If need notifications when app is closed

## References
- [i18next Documentation](https://www.i18next.com/)
- [MDN: Server-Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events)
- [Gamification Best Practices](https://www.nngroup.com/articles/engagement-techniques-gamification/)
