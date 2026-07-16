/**
 * Adds random jitter to a base delay value to prevent thundering herd
 * @param baseMs - base delay in milliseconds
 * @param jitterPercent - jitter percentage (default 0.15 = ±15%)
 * @returns delay with jitter applied
 */
export function addJitter(baseMs: number, jitterPercent = 0.15): number {
	const jitter = baseMs * jitterPercent;
	const randomJitter = Math.random() * jitter * 2 - jitter; // ±jitterPercent
	return Math.max(0, Math.round(baseMs + randomJitter));
}
