import { render as svelteRender } from '@testing-library/svelte';

// Расширенная функция render, которая автоматически подключает MSW (он уже активен глобально)
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function render(component: any, props?: Record<string, unknown>) {
	return svelteRender(component, { props });
}

// Хелпер для создания ответа с заголовками (имитация API)
export function createApiResponse<T>(data: T, options?: {
	status?: number;
	headers?: Record<string, string>;
}) {
	const { status = 200, headers = {} } = options || {};
	return new Response(JSON.stringify(data), {
		status,
		headers: new Headers(headers),
	});
}

// Хелпер для ожидания debounce
export function sleep(ms: number): Promise<void> {
	return new Promise(resolve => setTimeout(resolve, ms));
}
