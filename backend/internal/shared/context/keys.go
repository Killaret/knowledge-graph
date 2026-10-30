package contextkeys

type key string

// SkipAuthKey is a context key used to signal that SKIP_AUTH is active.
// The repository can use this to avoid scoping notes to public-only when
// the test user is authenticated via the skip-auth middleware.
const SkipAuthKey key = "skip_auth_enabled"

// AuthenticatedKey is a context key set by the JWT middleware after a token
// is validated. It lets the repository distinguish a verified identity that
// happens to be uuid.Nil (the seeded test user) from a truly anonymous
// request when scoping notes.
const AuthenticatedKey key = "authenticated"
