#!/bin/bash
# Seed Test Data - Linux/Mac
# Creates a known set of test notes and links for manual/automated testing.
# Requires the test stack to be running at localhost:18083.

set -e

NOTE_COUNT=${NOTE_COUNT:-100}
LINK_COUNT=${LINK_COUNT:-60}
NLP_WAIT_SECONDS=${NLP_WAIT_SECONDS:-600}
REPORT_PATH=${REPORT_PATH:-"$(dirname "$0")/seed-report.json"}
SEED=${SEED:-}
# Percentage of created notes to publish (make public) so the anonymous/public
# graph view has data to render in real-auth mode. Set to 0 to keep all notes
# private (previous behavior).
PUBLIC_PERCENT=${PUBLIC_PERCENT:-20}

# Optional deterministic seed for reproducible visual regression fixtures
if [ -n "$SEED" ]; then
    RANDOM=$SEED
fi

API_URL="http://localhost:18083/api/v1"
POSTGRES_CONTAINER="kg-test-postgres"

TEST_USER='{
    "login": "testuser",
    "email": "testuser@example.com",
    "password": "TestPassword123!"
}'

NOTE_TYPES=(
    star planet comet galaxy asteroid
    satellite debris nebula dust unknown blackhole
)

LINK_TYPES=(reference dependency related custom parent child)

REPORT=$(jq -n \
    --arg startedAt "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    '{
        startedAt: $startedAt,
        noteCount: 0,
        linkCount: 0,
        publicNoteCount: 0,
        embeddingCount: 0,
        keywordNoteCount: 0,
        noteIds: [],
        linkIds: [],
        typeDistribution: {},
        errors: [],
        durations: {}
    }')

add_error() {
    local context="$1"
    local message="$2"
    echo "[$context] ERROR: $message" >&2
    REPORT=$(echo "$REPORT" | jq --arg ctx "$context" --arg msg "$message" '.errors += [{context: $ctx, message: $msg}]')
}

# ---------------------------------------------------------------------------
# 1. Clean existing test data
# ---------------------------------------------------------------------------
echo "Cleaning existing test data..."
CLEAN_START=$(date +%s)
if docker exec "$POSTGRES_CONTAINER" psql -U kb_user -d knowledge_test -t -A -c \
    "TRUNCATE TABLE notes, note_embeddings, note_keywords, links CASCADE;" >/dev/null 2>&1; then
    echo "Existing test data truncated."
else
    add_error "clean" "Failed to truncate test data. Make sure the test stack is running."
    echo "$REPORT" | jq . > "$REPORT_PATH"
    exit 1
fi
CLEAN_END=$(date +%s)
REPORT=$(echo "$REPORT" | jq --arg v "$((CLEAN_END - CLEAN_START))" '.durations.cleanSeconds = ($v | tonumber)')

# ---------------------------------------------------------------------------
# 1.5. Seed the fixed test user (UUID all-zeros) for SKIP_AUTH mode.
#      This is the only place the test user is created; the migration that
#      used to insert it has been replaced by migration 029.
# ---------------------------------------------------------------------------
SEED_START=$(date +%s)
echo "Seeding fixed test user..."

SEED_PASSWORD="${SEED_TEST_USER_PASSWORD:-TestPassword123!}"
if ! docker exec -e APP_ENV=test -e SEED_TEST_USER_PASSWORD="$SEED_PASSWORD" kg-test-backend ./test-seed >/dev/null 2>&1; then
    add_error "seed-test-user" "Failed to seed test user. Make sure the test backend is running."
    echo "$REPORT" | jq . > "$REPORT_PATH"
    exit 1
fi

echo "Test user seeded."

SEED_END=$(date +%s)
REPORT=$(echo "$REPORT" | jq --arg v "$((SEED_END - SEED_START))" '.durations.seedTestUserSeconds = ($v | tonumber)')

# ---------------------------------------------------------------------------
# 2. Register / login test user
# ---------------------------------------------------------------------------
AUTH_START=$(date +%s)
echo "Registering test user..."
if ! curl -s -X POST "$API_URL/auth/register" -H "Content-Type: application/json" -d "$TEST_USER" >/dev/null 2>&1; then
    echo "Registration skipped or user already exists."
fi

echo "Logging in test user..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" -H "Content-Type: application/json" -d '{"login":"testuser","password":"TestPassword123!"}')
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    add_error "login" "Login failed. Response: $LOGIN_RESPONSE"
    echo "$REPORT" | jq . > "$REPORT_PATH"
    exit 1
fi

echo "Login successful."
AUTH_END=$(date +%s)
REPORT=$(echo "$REPORT" | jq --arg v "$((AUTH_END - AUTH_START))" '.durations.authSeconds = ($v | tonumber)')

# ---------------------------------------------------------------------------
# 3. Create notes
# ---------------------------------------------------------------------------
NOTES_START=$(date +%s)
echo "Creating $NOTE_COUNT test notes..."

NOTE_IDS=()
TYPE_DIST=$(jq -n '{}')

for i in $(seq 0 $((NOTE_COUNT - 1))); do
    type=${NOTE_TYPES[$((i % ${#NOTE_TYPES[@]}))]}
    idx=$(printf "%03d" $((i + 1)))
    title="Seed $type $idx"
    content="This is a test $type note number $i. It contains enough text for NLP processing to extract keywords and compute embeddings. Knowledge graph helps explore the cosmos of ideas."

    NOTE=$(jq -n \
        --arg title "$title" \
        --arg content "$content" \
        --arg type "$type" \
        '{title: $title, content: $content, type: $type}')

    NOTE_RESPONSE=$(curl -s -X POST "$API_URL/notes" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "$NOTE")

    NOTE_ID=$(echo "$NOTE_RESPONSE" | jq -r '.data.id // empty')

    if [ -n "$NOTE_ID" ] && [ "$NOTE_ID" != "null" ]; then
        NOTE_IDS+=("$NOTE_ID")
        TYPE_DIST=$(echo "$TYPE_DIST" | jq --arg t "$type" '(.[$t] // 0) += 1')
        if [ $(( (i + 1) % 10 )) -eq 0 ]; then
            echo "  Created $((i + 1)) notes..."
        fi
    else
        add_error "create-note" "Failed to create note $i ($type): $NOTE_RESPONSE"
    fi
done

NOTES_END=$(date +%s)
REPORT=$(echo "$REPORT" | jq \
    --argjson noteIds "$(printf '%s\n' "${NOTE_IDS[@]}" | jq -R . | jq -s .)" \
    --argjson typeDist "$TYPE_DIST" \
    --arg v "$((NOTES_END - NOTES_START))" \
    '.noteCount = ($noteIds | length) | .noteIds = $noteIds | .typeDistribution = $typeDist | .durations.createNotesSeconds = ($v | tonumber)')

echo "Created ${#NOTE_IDS[@]} notes."

# ---------------------------------------------------------------------------
# 3.5. Publish a subset of notes so the anonymous/public graph view has data
#      (previously all seeded notes stayed private, leaving the public graph
#      empty in real-auth mode).
# ---------------------------------------------------------------------------
PUBLISH_START=$(date +%s)
PUBLIC_COUNT=$(( (${#NOTE_IDS[@]} * PUBLIC_PERCENT + 99) / 100 ))
PUBLISHED_COUNT=0

if [ "$PUBLIC_COUNT" -gt 0 ]; then
    echo "Publishing $PUBLIC_COUNT of ${#NOTE_IDS[@]} notes ($PUBLIC_PERCENT%) for the public graph..."
    for ((i = 0; i < PUBLIC_COUNT && i < ${#NOTE_IDS[@]}; i++)); do
        NOTE_ID=${NOTE_IDS[$i]}
        if curl -s -f -X POST "$API_URL/notes/$NOTE_ID/publish" \
            -H "Authorization: Bearer $TOKEN" >/dev/null 2>&1; then
            PUBLISHED_COUNT=$((PUBLISHED_COUNT + 1))
        else
            add_error "publish-note" "Failed to publish note $NOTE_ID"
        fi
    done
    echo "Published $PUBLISHED_COUNT notes."
else
    echo "PUBLIC_PERCENT is 0 — all notes remain private."
fi

PUBLISH_END=$(date +%s)
REPORT=$(echo "$REPORT" | jq \
    --arg pc "$PUBLISHED_COUNT" \
    --arg v "$((PUBLISH_END - PUBLISH_START))" \
    '.publicNoteCount = ($pc | tonumber) | .durations.publishSeconds = ($v | tonumber)')

# ---------------------------------------------------------------------------
# 4. Wait for NLP processing
# ---------------------------------------------------------------------------
NLP_START=$(date +%s)
echo "Waiting for NLP service to process notes..."

ELAPSED=0
EMBEDDING_COUNT=0
KEYWORD_NOTE_COUNT=0

while [ $ELAPSED -lt $NLP_WAIT_SECONDS ]; do
    EMBEDDING_COUNT=$(docker exec "$POSTGRES_CONTAINER" psql -U kb_user -d knowledge_test -t -A \
        -c "SELECT COUNT(*) FROM note_embeddings;" 2>/dev/null | tr -d ' \n' || echo 0)
    KEYWORD_NOTE_COUNT=$(docker exec "$POSTGRES_CONTAINER" psql -U kb_user -d knowledge_test -t -A \
        -c "SELECT COUNT(DISTINCT note_id) FROM note_keywords;" 2>/dev/null | tr -d ' \n' || echo 0)

    echo "  NLP progress: $EMBEDDING_COUNT / ${#NOTE_IDS[@]} embeddings, $KEYWORD_NOTE_COUNT / ${#NOTE_IDS[@]} keyword notes..."

    if [ "$EMBEDDING_COUNT" -ge "${#NOTE_IDS[@]}" ] && [ "$KEYWORD_NOTE_COUNT" -ge "${#NOTE_IDS[@]}" ]; then
        break
    fi

    sleep 5
    ELAPSED=$((ELAPSED + 5))
done

NLP_END=$(date +%s)
REPORT=$(echo "$REPORT" | jq \
    --arg ec "$EMBEDDING_COUNT" \
    --arg kc "$KEYWORD_NOTE_COUNT" \
    --arg v "$((NLP_END - NLP_START))" \
    '.embeddingCount = ($ec | tonumber) | .keywordNoteCount = ($kc | tonumber) | .durations.nlpWaitSeconds = ($v | tonumber)')

if [ "$EMBEDDING_COUNT" -lt "${#NOTE_IDS[@]}" ] || [ "$KEYWORD_NOTE_COUNT" -lt "${#NOTE_IDS[@]}" ]; then
    add_error "nlp-wait" "NLP processing did not complete. Embeddings: $EMBEDDING_COUNT/${#NOTE_IDS[@]}, Keywords: $KEYWORD_NOTE_COUNT/${#NOTE_IDS[@]}"
else
    echo "NLP processing complete."
fi

# ---------------------------------------------------------------------------
# 5. Create links
# ---------------------------------------------------------------------------
LINKS_START=$(date +%s)

# Prefer links between public notes so the anonymous/public graph view has
# connected components to render. Fall back to the full pool only when there
# are fewer than two public notes.
if [ "$PUBLIC_COUNT" -ge 2 ]; then
    PUBLIC_NOTE_IDS=("${NOTE_IDS[@]:0:$PUBLIC_COUNT}")
else
    PUBLIC_NOTE_IDS=("${NOTE_IDS[@]}")
fi

echo "Creating $LINK_COUNT test links..."

LINK_IDS=()
CREATED_PAIRS=$(jq -n '{}')
ATTEMPTS=0
MAX_ATTEMPTS=$((LINK_COUNT * 5))

while [ ${#LINK_IDS[@]} -lt $LINK_COUNT ] && [ $ATTEMPTS -lt $MAX_ATTEMPTS ]; do
    ATTEMPTS=$((ATTEMPTS + 1))

    if [ ${#PUBLIC_NOTE_IDS[@]} -lt 2 ]; then
        add_error "create-link" "Not enough notes to create links."
        break
    fi

    SOURCE_INDEX=$((RANDOM % ${#PUBLIC_NOTE_IDS[@]}))
    TARGET_INDEX=$((RANDOM % ${#PUBLIC_NOTE_IDS[@]}))

    if [ "$SOURCE_INDEX" -eq "$TARGET_INDEX" ]; then
        continue
    fi

    SOURCE_ID=${PUBLIC_NOTE_IDS[$SOURCE_INDEX]}
    TARGET_ID=${PUBLIC_NOTE_IDS[$TARGET_INDEX]}

    if echo "$CREATED_PAIRS" | jq -e --arg k "$SOURCE_ID-$TARGET_ID" 'has($k)' >/dev/null 2>&1; then
        continue
    fi

    LINK_TYPE=${LINK_TYPES[$((RANDOM % ${#LINK_TYPES[@]}))]}
    WEIGHT=$(awk -v min=1 -v max=10 'BEGIN{srand(); printf "%.2f", (int(min+rand()*(max-min+1)))/10}')

    LINK=$(jq -n \
        --arg sid "$SOURCE_ID" \
        --arg tid "$TARGET_ID" \
        --arg lt "$LINK_TYPE" \
        --arg w "$WEIGHT" \
        '{source_note_id: $sid, target_note_id: $tid, link_type: $lt, weight: ($w | tonumber)}')

    LINK_RESPONSE=$(curl -s -X POST "$API_URL/links" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "$LINK")

    LINK_ID=$(echo "$LINK_RESPONSE" | jq -r '.data.id // empty')

    if [ -n "$LINK_ID" ] && [ "$LINK_ID" != "null" ]; then
        LINK_IDS+=("$LINK_ID")
        CREATED_PAIRS=$(echo "$CREATED_PAIRS" | jq --arg k "$SOURCE_ID-$TARGET_ID" '.[$k] = true')
        if [ $(( ${#LINK_IDS[@]} % 10 )) -eq 0 ]; then
            echo "  Created ${#LINK_IDS[@]} links..."
        fi
    else
        add_error "create-link" "Failed to create link $SOURCE_ID -> $TARGET_ID ($LINK_TYPE): $LINK_RESPONSE"
    fi
done

LINKS_END=$(date +%s)
REPORT=$(echo "$REPORT" | jq \
    --argjson linkIds "$(printf '%s\n' "${LINK_IDS[@]}" | jq -R . | jq -s .)" \
    --arg v "$((LINKS_END - LINKS_START))" \
    '.linkCount = ($linkIds | length) | .linkIds = $linkIds | .durations.createLinksSeconds = ($v | tonumber)')

echo "Created ${#LINK_IDS[@]} links."

# ---------------------------------------------------------------------------
# 6. Verify graph service
# ---------------------------------------------------------------------------
GRAPH_START=$(date +%s)
echo "Verifying graph service..."
GRAPH_RESPONSE=$(curl -s -X GET "http://localhost:19091/api/v1/graph/full" -H "Content-Type: application/json" || true)
if [ -n "$GRAPH_RESPONSE" ]; then
    TOTAL_NODES=$(echo "$GRAPH_RESPONSE" | jq -r '.meta.total_nodes // 0')
    TOTAL_LINKS=$(echo "$GRAPH_RESPONSE" | jq -r '.meta.total_links // 0')
    REPORT=$(echo "$REPORT" | jq --arg tn "$TOTAL_NODES" --arg tl "$TOTAL_LINKS" '.graphNodes = ($tn | tonumber) | .graphLinks = ($tl | tonumber)')
    echo "Graph service: $TOTAL_NODES nodes, $TOTAL_LINKS links."
else
    add_error "graph-service" "Graph service verification failed."
fi
GRAPH_END=$(date +%s)
REPORT=$(echo "$REPORT" | jq --arg v "$((GRAPH_END - GRAPH_START))" '.durations.graphCheckSeconds = ($v | tonumber)')

# ---------------------------------------------------------------------------
# 7. Save report
# ---------------------------------------------------------------------------
FINISHED_AT=$(date -u +%Y-%m-%dT%H:%M:%SZ)
TOTAL_SECONDS=$(echo "$REPORT" | jq '[.durations[] | tonumber] | add')
REPORT=$(echo "$REPORT" | jq --arg finishedAt "$FINISHED_AT" --arg total "$TOTAL_SECONDS" '.finishedAt = $finishedAt | .durations.totalSeconds = ($total | tonumber)')

echo "$REPORT" | jq . > "$REPORT_PATH"

echo ""
echo "Seed report saved to: $REPORT_PATH"
echo "Notes: $(echo "$REPORT" | jq -r '.noteCount') (public: $(echo "$REPORT" | jq -r '.publicNoteCount // 0')) | Links: $(echo "$REPORT" | jq -r '.linkCount') | Embeddings: $(echo "$REPORT" | jq -r '.embeddingCount') | Keyword notes: $(echo "$REPORT" | jq -r '.keywordNoteCount') | Graph nodes: $(echo "$REPORT" | jq -r '.graphNodes // 0') | Graph links: $(echo "$REPORT" | jq -r '.graphLinks // 0')"
echo "Total duration: $(echo "$REPORT" | jq -r '.durations.totalSeconds // 0') seconds"

if [ "$(echo "$REPORT" | jq '.errors | length')" -gt 0 ]; then
    echo "There were errors. See $REPORT_PATH for details." >&2
    exit 1
fi

echo "Test data seeded successfully!"
