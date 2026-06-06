# knowledge-graph-integration

**Version:** 1.0  
**Purpose:** API integration, type generation, contract testing  
**Status:** Active  
**Priority:** 🟢 High

---

## Overview

`knowledge-graph-integration` specializes in API integration between frontend and backend, type generation, and contract testing.

**Key Areas:**
- REST API client generation
- gRPC/WebSocket integration
- OpenAPI → TypeScript
- Protocol Buffers → TypeScript
- Contract testing (Pact)
- API versioning
- External services (OAuth, webhooks)

---

## API Integration

### 1. REST Client

#### HTTP Client Wrapper
```typescript
// lib/api/client.ts
import ky from 'ky';

const apiClient = ky.create({
  prefixUrl: import.meta.env.VITE_API_BASE_URL,
  headers: { 'Content-Type': 'application/json' },
  hooks: {
    beforeRequest: [
      (request) => {
        const token = getAuthToken();
        if (token) {
          request.headers.set('Authorization', `Bearer ${token}`);
        }
      },
    ],
    afterResponse: [
      async (request, options, response) => {
        if (response.status === 401) {
          await refreshToken();
          return fetch(request);
        }
        return response;
      },
    ],
  },
});

export const api = {
  get: <T>(url: string, options?: RequestInit): Promise<T> =>
    apiClient.get(url, options).json(),
  
  post: <T>(url: string, data?: any, options?: RequestInit): Promise<T> =>
    apiClient.post(url, { json: data, ...options }).json(),
  
  put: <T>(url: string, data?: any, options?: RequestInit): Promise<T> =>
    apiClient.put(url, { json: data, ...options }).json(),
  
  delete: <T>(url: string, options?: RequestInit): Promise<T> =>
    apiClient.delete(url, options).json(),
};
```

#### API Endpoints
```typescript
// lib/api/notes.ts
export const notesApi = {
  getAll: (params?: { limit?: number; offset?: number }) =>
    api.get<Note[]>('/api/v1/notes', { searchParams: params }),
  
  getById: (id: string) =>
    api.get<Note>(`/api/v1/notes/${id}`),
  
  create: (data: CreateNoteRequest) =>
    api.post<Note>('/api/v1/notes', data),
  
  update: (id: string, data: UpdateNoteRequest) =>
    api.put<Note>(`/api/v1/notes/${id}`, data),
  
  delete: (id: string) =>
    api.delete<void>(`/api/v1/notes/${id}`),
};
```

### 2. Type Generation

#### OpenAPI → TypeScript
```bash
# Generate types from OpenAPI spec
npx openapi-typescript http://localhost:8080/openapi.json -o src/types/api.d.ts
```

#### Protocol Buffers → TypeScript
```bash
# Generate from .proto files
protoc --ts_out ./src/gen --proto_path ./protos ./protos/*.proto
```

### 3. Contract Testing

#### Backend Contract Tests
```go
func TestAPISpec(t *testing.T) {
    router := setupRouter()
    
    tests := []struct {
        name           string
        method         string
        path           string
        body           string
        expectedStatus int
    }{
        {
            name:           "Create note returns 201",
            method:         "POST",
            path:           "/api/v1/notes",
            body:           `{"title":"Test","content":"Content","type":"star"}`,
            expectedStatus: 201,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req, _ := http.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
            w := httptest.NewRecorder()
            router.ServeHTTP(w, req)
            assert.Equal(t, tt.expectedStatus, w.Code)
        })
    }
}
```

#### Frontend Contract Tests
```typescript
describe('API Contract Tests', () => {
  it('createNote returns valid Note structure', async () => {
    const note = await notesApi.create({
      title: 'Test Note',
      content: 'Test Content',
      type: 'star',
    });
    
    expect(note).toHaveProperty('id');
    expect(note).toHaveProperty('title', 'Test Note');
    expect(note).toHaveProperty('created_at');
  });
});
```

---

## External Services

### OAuth Integration
```typescript
export class YandexAuthService {
  getAuthUrl(): string {
    const params = new URLSearchParams({
      client_id: this.config.clientId,
      redirect_uri: this.config.redirectUri,
      response_type: 'code',
      scope: this.config.scope.join(' '),
    });
    return `https://oauth.yandex.ru/authorize?${params}`;
  }
  
  async exchangeCodeForToken(code: string): Promise<YandexTokens> {
    const response = await fetch('https://oauth.yandex.ru/token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        grant_type: 'authorization_code',
        code,
        client_id: this.config.clientId,
        client_secret: this.config.clientSecret,
      }),
    });
    return response.json();
  }
}
```

### Webhook Integration
```typescript
export class WebhookService {
  async send(event: string, payload: any): Promise<void> {
    const signature = this.generateSignature(payload);
    
    await fetch(this.config.url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Webhook-Signature': signature,
        'X-Webhook-Event': event,
      },
      body: JSON.stringify(payload),
    });
  }
  
  private generateSignature(payload: any): string {
    const hmac = await crypto.subtle.digest(
      'SHA-256',
      new TextEncoder().encode(JSON.stringify(payload) + this.config.secret)
    );
    return Array.from(new Uint8Array(hmac))
      .map(b => b.toString(16).padStart(2, '0'))
      .join('');
  }
}
```

---

## API Versioning

```typescript
type APIVersion = 'v1' | 'v2';

const VERSIONS: Record<APIVersion, string> = {
  v1: '/api/v1',
  v2: '/api/v2',
};

export function getVersionedURL(version: APIVersion, endpoint: string): string {
  return `${VERSIONS[version]}${endpoint}`;
}
```

---

## Monitoring

### Request Tracking
```typescript
class RequestTracker {
  track(url: string, method: string): {
    markEnd: (status: number, error?: Error) => void;
  } {
    const startTime = performance.now();
    
    return {
      markEnd: (status: number, error?: Error) => {
        const duration = performance.now() - startTime;
        
        if (status === 500 || duration > 1000) {
          console.log('[API Metric]', { url, method, status, duration, error });
        }
      },
    };
  }
}
```

---

## Commands

### Generate Types
```bash
# OpenAPI types
npx openapi-typescript http://localhost:8080/openapi.json -o src/types/api.d.ts

# Protocol Buffers
protoc --ts_out ./src/gen --proto_path ./protos ./protos/*.proto
```

### Run Contract Tests
```bash
# Backend
go test -v ./tests/contract/...

# Frontend
npm run test:unit -- --run tests/contract/
```

---

## Best Practices

### Type Safety
```typescript
// ✅ GOOD: Strongly typed
interface User {
  id: string;
  name: string;
}

async function getUser(id: string): Promise<User> {
  const response = await fetch(`/api/users/${id}`);
  return response.json();
}
```

### Retry Logic
```typescript
async function retryableFetch<T>(
  url: string,
  options: RequestInit,
  maxRetries: number = 3
): Promise<T> {
  let lastError: Error | null = null;
  
  for (let i = 0; i < maxRetries; i++) {
    try {
      const response = await fetch(url, options);
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      return await response.json();
    } catch (error) {
      lastError = error as Error;
      await sleep(Math.pow(2, i) * 1000);
    }
  }
  
  throw lastError;
}
```

---

**Tools:** `integration-tools.md`  
**Contract Coverage:** 100% of public APIs  
**Type Coverage:** 100% TypeScript strict mode