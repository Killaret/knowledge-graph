import { goto } from "$app/navigation";
import { isAuthenticated, skipAuthMode } from "$shared/stores/auth.svelte";

/**
 * Redirect already-authenticated users away from public-only pages (login, register, etc.).
 * Respects skip-auth test mode.
 */
export function useAnonymousGuard(redirectTo: string = "/") {
  if (isAuthenticated() && !skipAuthMode()) {
    goto(redirectTo);
  }
}

/**
 * Redirect unauthenticated users to login, then back to the current page.
 */
export function useRequireAuth(redirectTo: string = "/auth/login") {
  if (!isAuthenticated()) {
    goto(redirectTo);
  }
}
