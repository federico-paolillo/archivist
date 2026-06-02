import { type ApiClient, makeApiClient } from "@archivist/deps/api-client.ts";

export interface Deps {
	api: ApiClient;
}

export function makeDeps(): Deps {
	return {
		api: makeApiClient(import.meta.env.VITE_API_BASE_PATH),
	};
}
