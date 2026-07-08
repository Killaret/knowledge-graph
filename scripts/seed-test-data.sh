#!/bin/bash
# Seed Test Data - Linux/Mac
# This script registers a test user and creates test notes and links

set -e

API_URL="http://localhost:18083/api/v1"

echo "Seeding test data..."

# Test user credentials
TEST_USER='{
    "login": "testuser",
    "email": "testuser@example.com",
    "password": "TestPassword123!"
}'

# Register test user
echo "Registering test user..."
if curl -s -X POST "$API_URL/auth/register" -H "Content-Type: application/json" -d "$TEST_USER" > /dev/null; then
    echo "Test user registered successfully"
else
    echo "User might already exist or registration failed"
fi

# Login to get token
echo "Logging in test user..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" -H "Content-Type: application/json" -d '{"login":"testuser","password":"TestPassword123!"}')
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "Login failed"
    exit 1
fi

echo "Login successful"

# Create test notes
echo "Creating test notes..."
NOTE_TYPES=("star" "planet" "comet" "galaxy" "asteroid")
NOTE_IDS=()

for type in "${NOTE_TYPES[@]}"; do
    NOTE="{
        \"title\": \"Test $type Note\",
        \"content\": \"This is a test note of type $type\",
        \"type\": \"$type\"
    }"
    
    NOTE_RESPONSE=$(curl -s -X POST "$API_URL/notes" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d "$NOTE")
    NOTE_ID=$(echo "$NOTE_RESPONSE" | jq -r '.data.id')
    
    if [ -n "$NOTE_ID" ] && [ "$NOTE_ID" != "null" ]; then
        NOTE_IDS+=("$NOTE_ID")
        echo "Created note: Test $type Note (ID: $NOTE_ID)"
    else
        echo "Failed to create note: $NOTE_RESPONSE"
    fi
done

# Create test links
echo "Creating test links..."
if [ ${#NOTE_IDS[@]} -ge 2 ]; then
    LINK="{
        \"source_note_id\": \"${NOTE_IDS[0]}\",
        \"target_note_id\": \"${NOTE_IDS[1]}\",
        \"link_type\": \"related\",
        \"weight\": 1.0
    }"
    
    LINK_RESPONSE=$(curl -s -X POST "$API_URL/links" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d "$LINK")
    echo "Created link between ${NOTE_IDS[0]} and ${NOTE_IDS[1]}"
    
    # Create second link
    if [ ${#NOTE_IDS[@]} -ge 3 ]; then
        LINK2="{
            \"source_note_id\": \"${NOTE_IDS[1]}\",
            \"target_note_id\": \"${NOTE_IDS[2]}\",
            \"link_type\": \"dependency\",
            \"weight\": 0.5
        }"
        
        LINK_RESPONSE2=$(curl -s -X POST "$API_URL/links" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d "$LINK2")
        echo "Created link between ${NOTE_IDS[1]} and ${NOTE_IDS[2]}"
    fi
fi

echo ""
echo "Test data seeded successfully!"
echo "Created ${#NOTE_IDS[@]} notes and 2 links"
