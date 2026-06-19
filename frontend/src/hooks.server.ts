import { sequence } from '@sveltejs/kit/hooks';
import type { Handle } from '@sveltejs/kit';

/**
 * Graph service proxy middleware
 * Proxies requests from /graph-service/api/* to graph service at localhost:9091
 */
const graphServiceProxy: Handle = async ({ event, resolve }) => {
	const url = event.url;
	
	// Check if this is a graph service request
	if (url.pathname.startsWith('/graph-service/api/')) {
		// Remove /graph-service prefix to get the actual path
		const targetPath = url.pathname.replace('/graph-service', '');
		const targetUrl = `http://localhost:9091${targetPath}${url.search}`;
		
		try {
			// Forward the request to graph service
			const response = await fetch(targetUrl, {
				method: event.request.method,
				headers: {
					...Object.fromEntries(event.request.headers),
					host: 'localhost:9091',
				},
				body: event.request.method !== 'GET' && event.request.method !== 'HEAD'
					? await event.request.blob()
					: undefined,
			});
			
			// Return the response from graph service
			return new Response(response.body, {
				status: response.status,
				headers: {
					...Object.fromEntries(response.headers.entries()),
					// Remove hop-by-hop headers
					'transfer-encoding': undefined,
					'connection': undefined,
					'keep-alive': undefined,
				},
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