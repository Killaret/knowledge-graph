package contextkeys

// SkipAuthKey is a context key used to signal that SKIP_AUTH is active.
// The repository can use this to avoid scoping notes to public-only when
// the test user is authenticated via the skip-auth middleware.
const SkipAuthKey = "skip_auth_enabled"
