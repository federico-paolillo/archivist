export type Fetcher = typeof fetch;

export interface AuthApiClient {
	getSession: () => Promise<boolean>;
	login: (password: string) => Promise<boolean>;
	logout: () => Promise<"ok" | "unauthorized" | "failed">;
}

export interface Deps {
	api: AuthApiClient;
}

export function normalizeApiBasePath(value: string | undefined): string {
	const configuredValue = value?.trim();
	const rawValue = configuredValue || "/api";
	const withLeadingSlash = rawValue.startsWith("/") ? rawValue : `/${rawValue}`;
	const sameOriginPath = withLeadingSlash.replace(/^\/+/u, "/");
	const normalized = sameOriginPath.replace(/\/+$/u, "");

	return normalized || (configuredValue ? "" : "/api");
}

export function makeAuthApiClient(
	apiBasePath: string,
	fetcher: Fetcher = fetch,
): AuthApiClient {
	const basePath = normalizeApiBasePath(apiBasePath);
	const apiUrl = (path: string) => `${basePath}${path}`;

	return {
		async getSession() {
			const response = await fetcher(apiUrl("/auth/session"), {
				method: "GET",
				credentials: "include",
			});

			if (response.status === 204) {
				return true;
			}

			if (response.status === 401) {
				return false;
			}

			throw new Error("Session check failed.");
		},

		async login(password: string) {
			const response = await fetcher(apiUrl("/login"), {
				method: "POST",
				credentials: "include",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify({ password }),
			});

			return response.status === 204;
		},

		async logout() {
			const response = await fetcher(apiUrl("/logout"), {
				method: "POST",
				credentials: "include",
			});

			if (response.status === 204) {
				return "ok";
			}

			if (response.status === 401) {
				return "unauthorized";
			}

			return "failed";
		},
	};
}

export function makeDeps(): Deps {
	return {
		api: makeAuthApiClient(import.meta.env.VITE_API_BASE_PATH),
	};
}
