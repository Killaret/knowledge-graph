# Obsidian Import Specification

## Overview

**Goal:** Enable users to import their Obsidian vaults into Knowledge Graph while preserving all internal links and structure.

**Value Proposition:** Obsidian users (millions of personal knowledge base users) can instantly migrate their existing graph-structured notes without manual re-entry.

---

## API Specification

### POST /import/obsidian

Upload an Obsidian vault (as ZIP archive or server-side directory path).

**Request:**
```http
POST /api/v1/import/obsidian
Content-Type: multipart/form-data

file: <obsidian-vault.zip>
options: {
  "duplicateStrategy": "skip" | "overwrite" | "merge",
  "extractFrontmatter": true | false,
  "defaultLinkType": "reference" | "related"
}
```

**Response (202 Accepted):**
```json
{
  "taskId": "import:obsidian:uuid",
  "status": "pending",
  "message": "Import queued for processing",
  "estimatedTime": "30s"
}
```

### GET /import/status/{taskId}

Track import progress.

**Response:**
```json
{
  "taskId": "import:obsidian:uuid",
  "status": "processing" | "completed" | "failed",
  "progress": {
    "totalFiles": 1500,
    "processedFiles": 750,
    "createdNotes": 748,
    "createdLinks": 1245,
    "errors": 2
  },
  "startedAt": "2024-01-15T10:30:00Z",
  "completedAt": null,
  "errorMessage": null
}
```

---

## Markdown Parsing

### Obsidian Link Syntax

| Syntax | Meaning | Graph Action |
|--------|---------|--------------|
| `[[Target Note]]` | Internal link | Create `reference` link to note titled "Target Note" |
| `[[Target\|Display]]` | Aliased link | Create `reference` link, use "Display" as UI text |
| `[[Target#Heading]]` | Block reference | Link to note (heading anchors not yet supported) |
| `![[Embedded.png]]` | Embedded file | Skip or create attachment placeholder |

### Parsing Rules

1. **Title Extraction:**
   - First priority: First H1 header (`# Title`)
   - Second priority: Filename (without `.md` extension)

2. **Content Extraction:**
   - Full Markdown body after title
   - Preserve formatting (converted to internal format if needed)

3. **Link Discovery:**
   - Regex: `\[\[([^\]|#]+)(?:#([^\]]+))?(?:\|([^\]]+))?\]\]`
   - Extract target, optional heading, optional alias

4. **YAML Frontmatter (Phase 5.1):**
   ```yaml
   ---
   title: Override Title
   tags: [tag1, tag2]
   aliases: [Alias Name]
   created: 2023-01-01
   ---
   ```

---

## Architecture

### Component Diagram

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  User Upload    │────▶│  API Handler     │────▶│  Asynq Queue    │
│  (ZIP file)     │     │  (validate, save)│     │  (import:obsidian)
└─────────────────┘     └──────────────────┘     └────────┬────────┘
                                                           │
                           ┌───────────────────────────────┘
                           ▼
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Batch Insert   │◄────│  Import Worker   │◄────│  ZIP Streamer   │
│  (Repository)   │     │  (orchestrator)  │     │  (extract files)│
└─────────────────┘     └──────────────────┘     └─────────────────┘
        │
        ▼
┌─────────────────┐     ┌──────────────────┐
│  Link Creator   │────▶│  Progress Store  │
│  (Async tasks)  │     │  (Redis)         │
└─────────────────┘     └──────────────────┘
```

### Worker Task Breakdown

```go
const TypeImportObsidian = "import:obsidian"

type ImportPayload struct {
    UploadID         string
    FilePath         string
    DuplicateStrategy string
    BatchSize        int  // default 100
}

// Task Options:
// - MaxRetry(2)
// - Timeout(5m)
// - UniqueKey("import:{upload_id}")
```

**Processing Steps:**
1. Stream ZIP file entries
2. Parse each `.md` file (title, content, links)
3. Batch create notes (respecting duplicate strategy)
4. Resolve link targets (by title)
5. Batch create links
6. Update progress in Redis
7. Mark completion

### Duplicate Resolution Strategies

| Strategy | Behavior |
|----------|----------|
| `skip` | Skip existing notes, preserve current graph |
| `overwrite` | Replace note content, recreate all links |
| `merge` | Combine content (append), union of links |

---

## Database Schema

### Import Job Tracking (Optional Enhancement)

```sql
CREATE TABLE import_jobs (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    source_type VARCHAR(20) NOT NULL, -- 'obsidian'
    status VARCHAR(20) NOT NULL, -- pending, processing, completed, failed
    total_files INT DEFAULT 0,
    processed_files INT DEFAULT 0,
    created_notes INT DEFAULT 0,
    created_links INT DEFAULT 0,
    error_count INT DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

## Error Handling

### Recoverable Errors (Continue Processing)

- Invalid file encoding (try fallback encodings)
- Malformed Markdown (skip file, log warning)
- Duplicate link (skip, log debug)

### Fatal Errors (Stop Import)

- ZIP corruption
- Database connection failure
- Worker timeout

### Error Response Format

```json
{
  "taskId": "import:obsidian:uuid",
  "status": "failed",
  "error": {
    "code": "ZIP_CORRUPTED",
    "message": "Archive cannot be extracted",
    "file": "vault.zip",
    "recoverable": false
  },
  "partialResults": {
    "processedFiles": 150,
    "createdNotes": 148
  }
}
```

---

## Security Considerations

1. **File Size Limits:** Max 100MB ZIP files
2. **Path Traversal:** Sanitize all file paths, reject `../` patterns
3. **File Types:** Only process `.md` files, ignore others
4. **Rate Limiting:** Max 1 import per minute per user
5. **Storage:** Delete uploaded ZIP after processing (or after 24h)

---

## Frontend UX

### Upload Flow

1. **Drag & Drop Zone:** Accept `.zip` files
2. **Options Dialog:** Duplicate strategy selector, frontmatter toggle
3. **Progress Display:** Real-time progress bar with counts
4. **Results View:** Summary of imported notes and links
5. **Error List:** Clickable errors showing which files failed

### Progress Polling

```typescript
// Poll every 2 seconds while processing
const pollImport = async (taskId: string) => {
  const response = await fetch(`/api/v1/import/status/${taskId}`);
  const data = await response.json();
  
  updateProgressBar(data.progress);
  
  if (data.status === 'processing') {
    setTimeout(() => pollImport(taskId), 2000);
  }
};
```

---

## Implementation Phases

### Phase 5.0: MVP (Core Import)

- [ ] ZIP upload endpoint
- [ ] Basic Markdown parsing (title, content, `[[links]]`)
- [ ] Batch note creation
- [ ] Link resolution and creation
- [ ] Progress tracking
- [ ] Simple upload UI

### Phase 5.1: Enhanced Parsing

- [ ] YAML frontmatter support (tags, aliases)
- [ ] Heading anchor links (`[[Note#Heading]]`)
- [ ] Embedded image handling
- [ ] Folder structure preservation (optional)

### Phase 5.2: Incremental Sync

- [ ] Detect changes since last import
- [ ] Update only modified notes
- [ ] Bi-directional sync (advanced, future)

---

## Testing Strategy

### Test Vaults Required

1. **Basic Vault:** 10 notes, simple links
2. **Large Vault:** 2000+ notes, performance test
3. **Complex Vault:** Nested folders, images, frontmatter
4. **Edge Cases:** Special characters, empty files, circular links

### Validation Checklist

- [ ] All notes created with correct titles
- [ ] All `[[links]]` converted to graph connections
- [ ] Link aliases handled correctly
- [ ] Duplicate strategies work as specified
- [ ] Progress tracking accurate
- [ ] Memory usage acceptable for 1000+ files

---

## References

- [Obsidian Help: Internal Links](https://help.obsidian.md/Linking+notes+and+files/Internal+links)
- [Goldmark Markdown Parser](https://github.com/yuin/goldmark)
- [ARCHITECTURE_ROADMAP.md Phase 5](./ARCHITECTURE_ROADMAP.md#phase-5-obsidian-import-killer-feature)
