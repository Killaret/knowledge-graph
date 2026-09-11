/**
 * Check whether a path is an exact match or a sub-route of any of the given
 * public route prefixes. A route "/" only matches the exact "/" path.
 */
export function isPublicRoute(currentPath: string, publicRoutes: string[]): boolean {
  return publicRoutes.some(
    (route) => currentPath === route || currentPath.startsWith(`${route}/`)
  );
}
