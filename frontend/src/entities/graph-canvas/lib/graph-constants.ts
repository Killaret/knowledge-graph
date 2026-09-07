/**
 * Shared visual constants for the 2D graph.
 *
 * Keeping BASE_NODE_RADIUS in one place lets the service tools
 * (ghost, black hole, quick capture) size themselves relative to
 * real notes instead of using disconnected magic numbers.
 */

/** Base radius used for drawing every note on the canvas. */
export const BASE_NODE_RADIUS = 16;

/** Comfortable padding between a service tool and the inner edge of the cockpit frame. */
export const SERVICE_TOOL_MARGIN = 48;
