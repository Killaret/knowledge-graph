import { sequence } from '@sveltejs/kit/hooks';
import type { Handle } from '@sveltejs/kit';

/**
 * Graph service proxy middleware
 * Proxies requests from /graph-service/api/* to graph service
 * Uses environment variable GRAPH_SERVICE_URL or defaults to Docker service name
 */
const graphServiceProxy: Handle = async ({ event, resolve }) => {
	const url = event.url;

	// Get graph service URL from environment or use Docker service name
	const graphServiceUrl = process.env.GRAPH_SERVICE_URL || process.env.VITE_GRAPH_SERVICE_URL || 'http://graph-service:9091';

	// Check if this is a graph service request
	if (url.pathname.startsWith('/graph-service/api/')) {
		// Remove /graph-service prefix to get the actual path
		const targetPath = url.pathname.replace('/graph-service', '');
		const targetUrl = `${graphServiceUrl}${targetPath}${url.search}`;

		try {
			// Forward the request to graph service
			const response = await fetch(targetUrl, {
				method: event.request.method,
				headers: {
					...Object.fromEntries(event.request.headers),
					host: new URL(graphServiceUrl).host,
				},
				body: event.request.method !== 'GET' && event.request.method !== 'HEAD'
					? await event.request.blob()
					: undefined,
			});
			
			// Return the response from graph service
			const headers = new Headers(response.headers);
			// Remove hop-by-hop headers
			headers.delete('transfer-encoding');
			headers.delete('connection');
			headers.delete('keep-alive');
			
			return new Response(response.body, {
				status: response.status,
				headers: headers,
			});
		} catch (error) {
			console.error('[Graph Service Proxy] Error:', error);
			return new Response(JSON.stringify({ error: 'Failed to connect to graph service' }), {
				status: 502,
				headers: { 'content-type': 'application/json' }
			});
		}
	}
	
	// Not a graph service request, continue normally
	return resolve(event);
};

export const handle = sequence(graphServiceProxy);