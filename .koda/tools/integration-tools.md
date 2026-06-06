# Инструменты Integration Агента

## 🎯 Основные задачи

1. API Integration (REST, gRPC)
2. External services mapping
3. Type definitions generation
4. Contract testing
5. API versioning

---

## 🔗 API Integration

### 1. REST Client

#### HTTP Client Wrapper
```typescript
// lib/api/client.ts
import ky from 'ky';

const apiClient = ky.create({
  prefixUrl: import.meta.env.VITE_API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
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
          // Handle token refresh
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
import { api } from './client';
import type { Note, CreateNoteRequest, UpdateNoteRequest } from '$lib/types';

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

// lib/api/graph.ts
export const graphApi = {
  getFullGraph: (limit?: number) =>
    api.get<GraphData>('/api/v1/graph/full', { 
      searchParams: { limit } 
    }),
  
  getNoteGraph: (id: string, depth: number = 2) =>
    api.get<GraphData>(`/api/v1/graph/note/${id}`, {
      searchParams: { depth }
    }),
  
  getDelta: (since: string) =>
    api.get<GraphDelta>('/api/v1/graph/delta', {
      searchParams: { since }
    }),
};
```

### 2. gRPC Integration

#### WebAssembly gRPC Client
```typescript
// grpc/grpc-client.ts
import { GrpcWebFetchTransport } from '@protobuf-ts/grpcweb-transport';
import { NoteServiceClient } from '../gen/notes_pb';

const transport = new GrpcWebFetchTransport({
  baseUrl: import.meta.env.VITE_GRPC_WEB_PROXY,
  meta: {
    'Authorization': `Bearer ${getAuthToken()}`
  }
});

export const noteService = new NoteServiceClient(transport);

export async function getNote(id: string) {
  const response = await noteService.getNote({ id });
  return response.note;
}

export async function createNote(note: CreateNoteRequest) {
  const response = await noteService.createNote(note);
  return response.note;
}
```

### 3. External Services

#### Third-party API Integration
```typescript
// lib/api/external/yandex.ts
interface YandexAuthConfig {
  clientId: string;
  redirectUri: string;
  scope: string[];
}

export class YandexAuthService {
  private config: YandexAuthConfig;
  
  constructor(config: YandexAuthConfig) {
    this.config = config;
  }
  
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
  
  async getUserProfile(accessToken: string): Promise<YandexUser> {
    const response = await fetch('https://login.yandex.ru/info', {
      headers: { 'Authorization': `OAuth ${accessToken}` },
    });
    
    return response.json();
  }
}
```

#### Webhook Integration
```typescript
// lib/services/webhook.ts
interface WebhookConfig {
  url: string;
  secret: string;
  events: string[];
}

export class WebhookService {
  private config: WebhookConfig;
  
  constructor(config: WebhookConfig) {
    this.config = config;
  }
  
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
  
  verifySignature(payload: any, signature: string): boolean {
    const expected = this.generateSignature(payload);
    return crypto.timingSafeEqual(
      Buffer.from(expected),
      Buffer.from(signature)
    );
  }
}
```

---

## 📝 Type Generation

### 1. OpenAPI to TypeScript

#### Code Generation
```bash
# Install openapi-typescript
npm install -D openapi-typescript

# Generate types from OpenAPI spec
npx openapi-typescript http://localhost:8080/openapi.json -o src/types/api.d.ts

# Or from file
npx openapi-typescript ./openapi.yaml -o src/types/api.d.ts
```

#### Generated Types
```typescript
// src/types/api.d.ts
export interface paths {
  '/api/v1/notes': {
    get: operations['getNotes'];
    post: operations['createNote'];
  };
  '/api/v1/notes/{id}': {
    get: operations['getNote'];
    put: operations['updateNote'];
    delete: operations['deleteNote'];
  };
}

export interface components {
  schemas: {
    Note: {
      id: string;
      title: string;
      content: string;
      type: 'star' | 'planet' | 'comet';
      created_at: string;
      updated_at: string;
    };
    CreateNoteRequest: {
      title: string;
      content: string;
      type: string;
    };
  };
}

export interface operations {
  getNotes: {
    responses: {
      200: components['schemas']['Note'][];
    };
  };
  createNote: {
    requestBody: {
      content: {
        'application/json': components['schemas']['CreateNoteRequest'];
      };
    };
    responses: {
      201: components['schemas']['Note'];
    };
  };
}
```

### 2. Protocol Buffers

#### Generate TypeScript from .proto
```bash
# Install protoc-gen-ts
npm install -D @protobuf-ts/plugin

# Generate
protoc --ts_out ./src/gen --proto_path ./protos ./protos/*.proto
```

#### Generated Code
```typescript
// gen/notes_pb.ts
export interface GetNoteRequest {
  id: string;
}

export interface Note {
  id: string;
  title: string;
  content: string;
  type: string;
  metadata: Record<string, string>;
  createdAt: Date;
  updatedAt: Date;
}

export const NoteService = {
  typeName: "graph.NoteService",
  methods: {
    getNote: {
      name: "GetNote",
      I: GetNoteRequest,
      O: Note,
      kind: MethodKind.Unary,
    },
  },
};
```

---

## 🧪 Contract Testing

### 1. API Contract Tests

#### Backend Contract Tests
```go
// tests/contract/api_contract_test.go
func TestAPISpec(t *testing.T) {
    router := setupRouter()
    
    tests := []struct {
        name           string
        method         string
        path           string
        body           string
        expectedStatus int
        validateResponse func(*testing.T, *httptest.ResponseRecorder)
    }{
        {
            name:           "Create note returns 201",
            method:         "POST",
            path:           "/api/v1/notes",
            body:           `{"title":"Test","content":"Content","type":"star"}`,
            expectedStatus: 201,
            validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
                var note Note
                json.Unmarshal(w.Body.Bytes(), &note)
                assert.NotEmpty(t, note.ID)
                assert.NotEmpty(t, note.CreatedAt)
            },
        },
        {
            name:           "Get non-existent note returns 404",
            method:         "GET",
            path:           "/api/v1/notes/invalid-id",
            expectedStatus: 404,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req, _ := http.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
            if tt.body != "" {
                req.Header.Set("Content-Type", "application/json")
            }
            
            w := httptest.NewRecorder()
            router.ServeHTTP(w, req)
            
            assert.Equal(t, tt.expectedStatus, w.Code)
            if tt.validateResponse != nil {
                tt.validateResponse(t, w)
            }
        })
    }
}
```

#### Frontend Contract Tests
```typescript
// tests/contract/api-contract.spec.ts
import { describe, it, expect } from 'vitest';
import { notesApi } from '$lib/api/notes';
import type { Note } from '$lib/types';

describe('API Contract Tests', () => {
  it('createNote returns valid Note structure', async () => {
    const note = await notesApi.create({
      title: 'Test Note',
      content: 'Test Content',
      type: 'star',
    });
    
    expect(note).toHaveProperty('id');
    expect(note).toHaveProperty('title', 'Test Note');
    expect(note).toHaveProperty('content', 'Test Content');
    expect(note).toHaveProperty('type', 'star');
    expect(note).toHaveProperty('created_at');
    expect(note).toHaveProperty('updated_at');
  });
  
  it('getNote returns note with all fields', async () => {
    const note = await notesApi.getById('123');
    
    expect(note.id).toBeDefined();
    expect(typeof note.title).toBe('string');
    expect(typeof note.content).toBe('string');
    expect(['star', 'planet', 'comet']).toContain(note.type);
  });
  
  it('API returns consistent error format', async () => {
    try {
      await notesApi.getById('invalid-id');
      expect.fail('Should have thrown an error');
    } catch (error: any) {
      expect(error.response).toBeDefined();
      expect(error.response.data).toHaveProperty('code');
      expect(error.response.data).toHaveProperty('message');
    }
  });
});
```

### 2. Pact Contract Testing

```typescript
// tests/pact/notes-provider.spec.ts
import { Pact } from '@pact-foundation/pact';
import { notesApi } from '$lib/api/notes';

const pact = new Pact({
  consumer: 'frontend',
  provider: 'backend',
  port: 1234,
});

describe('Notes API Contract', () => {
  beforeAll(async () => {
    await pact.setup();
  });
  
  afterAll(async () => {
    await pact.finalize();
  });
  
  it('GET /api/v1/notes returns array of notes', async () => {
    await pact.addInteraction({
      state('notes exist'),
      uponReceiving('a request for notes'),
      withRequest({
        method: 'GET',
        path: '/api/v1/notes',
      }),
      willRespondWith({
        status: 200,
        headers: { 'Content-Type': 'application/json' },
        body: {
          nodes: [
            {
              id: '1',
              title: 'Test Note',
              type: 'star',
            },
          ],
          links: [],
        },
      }),
    });
    
    const result = await notesApi.getAll();
    expect(result.nodes.length).toBeGreaterThan(0);
  });
});
```

---

## 🔀 API Versioning

### Version Management
```typescript
// lib/api/versioning.ts
type APIVersion = 'v1' | 'v2';

const VERSIONS: Record<APIVersion, string> = {
  v1: '/api/v1',
  v2: '/api/v2',
};

export function getVersionedURL(version: APIVersion, endpoint: string): string {
  return `${VERSIONS[version]}${endpoint}`;
}

// Migration strategy
export const APIMigrator = {
  async migrateToV2<T>(v1Call: () => Promise<T>, fallback: boolean = true): Promise<T> {
    try {
      const v2Result = await v1Call(); // In practice, call v2 endpoint
      return v2Result;
    } catch (error) {
      if (fallback) {
        console.warn('V2 migration failed, falling back to V1');
        return v1Call();
      }
      throw error;
    }
  },
};
```

---

## 📊 Monitoring & Logging

### Request Tracking
```typescript
// lib/api/tracking.ts
interface RequestMetrics {
  url: string;
  method: string;
  startTime: number;
  endTime?: number;
  status?: number;
  error?: Error;
}

class RequestTracker {
  private metrics: RequestMetrics[] = [];
  
  track(url: string, method: string): {
    markEnd: (status: number, error?: Error) => void;
  } {
    const metric: RequestMetrics = {
      url,
      method,
      startTime: performance.now(),
    };
    
    this.metrics.push(metric);
    
    return {
      markEnd: (status: number, error?: Error) => {
        metric.endTime = performance.now();
        metric.status = status;
        metric.error = error;
        
        // Send to analytics
        if (this.shouldSend(metric)) {
          this.sendToAnalytics(metric);
        }
      },
    };
  }
  
  private shouldSend(metric: RequestMetrics): boolean {
    return metric.status === 500 || 
           (metric.endTime - metric.startTime) > 1000;
  }
  
  private sendToAnalytics(metric: RequestMetrics): void {
    // Send to monitoring service
    console.log('[API Metric]', metric);
  }
}

export const requestTracker = new RequestTracker();
```

### Error Tracking
```typescript
// lib/api/error-tracking.ts
interface APIError {
  code: string;
  message: string;
  details?: Record<string, any>;
  timestamp: string;
  endpoint: string;
  method: string;
}

export function trackAPIError(error: any, endpoint: string, method: string): void {
  const apiError: APIError = {
    code: error.code || 'UNKNOWN_ERROR',
    message: error.message || 'An error occurred',
    details: error.details,
    timestamp: new Date().toISOString(),
    endpoint,
    method,
  };
  
  // Send to error tracking service
  if (window.Sentry) {
    window.Sentry.captureException(error, {
      contexts: { api: apiError },
    });
  }
  
  // Log locally for debugging
  console.error('[API Error]', apiError);
}
```

---

## 🔧 Команды

### Generate Types
```powershell
# OpenAPI types
npx openapi-typescript http://localhost:8080/openapi.json -o src/types/api.d.ts

# Protocol Buffers
protoc --ts_out ./src/gen --proto_path ./protos ./protos/*.proto

# GraphQL (if using)
npx graphql-codegen --config codegen.ts
```

### Run Contract Tests
```powershell
# Backend contract tests
go test -v ./tests/contract/...

# Frontend contract tests
npm run test:unit -- --run tests/contract/

# Pact tests
npx pact-provider-verifier
```

### API Documentation
```powershell
# Generate OpenAPI spec
go run ./cmd/server -generate-spec > openapi.yaml

# Serve Swagger UI
docker run -p 8081:8080 \
  -e SWAGGER_JSON=/openapi.yaml \
  -v $(pwd):/openapi.yaml \
  swaggerapi/swagger-ui
```

---

## 📚 Best Practices

### 1. Type Safety
```typescript
// ✅ GOOD: Strongly typed
interface User {
  id: string;
  name: string;
}

async function getUser(id: string): Promise<User> {
  const response = await fetch(`/api/users/${id}`);
  return response.json(); // Type-safe!
}

// ❌ BAD: Any type
async function getUserBad(id: string): Promise<any> {
  const response = await fetch(`/api/users/${id}`);
  return response.json(); // No type safety!
}
```

### 2. Error Handling
```typescript
// ✅ GOOD: Typed errors
class APIError extends Error {
  constructor(
    public code: string,
    public status: number,
    public details?: Record<string, any>
  ) {
    super(code);
  }
}

async function safeAPI<T>(
  apiCall: () => Promise<T>
): Promise<{ data: T | null; error: APIError | null }> {
  try {
    const data = await apiCall();
    return { data, error: null };
  } catch (error: any) {
    const apiError = new APIError(
      error.code || 'UNKNOWN',
      error.status || 500,
      error.details
    );
    return { data: null, error: apiError };
  }
}
```

### 3. Retry Logic
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
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      
      return await response.json();
    } catch (error) {
      lastError = error as Error;
      await sleep(Math.pow(2, i) * 1000); // Exponential backoff
    }
  }
  
  throw lastError;
}
```
