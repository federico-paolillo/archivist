import type {
	ArticleDetail,
	ArticleListResponse,
} from "@archivist/deps/models.ts";

export type Fetcher = typeof fetch;

export interface AuthApiClient {
	getSession: () => Promise<boolean>;
	login: (password: string) => Promise<boolean>;
	logout: () => Promise<"ok" | "unauthorized" | "failed">;
}

export interface ArticleApiClient {
	listArticles: () => Promise<ArticleListResponse>;
	getArticle: (id: string) => Promise<ArticleDetail>;
	deleteArticle: (id: string) => Promise<void>;
	forceDeleteArticle: (id: string) => Promise<void>;
}

export type ApiClient = AuthApiClient & ArticleApiClient;

export class ApiUnauthorizedError extends Error {
	constructor() {
		super("Authentication required.");
		this.name = "ApiUnauthorizedError";
	}
}

export class ApiRequestError extends Error {
	constructor(message: string) {
		super(message);
		this.name = "ApiRequestError";
	}
}

export function normalizeApiBasePath(value: string | undefined): string {
	const configuredValue = value?.trim();
	if (!configuredValue) {
		return "/api";
	}

	const withLeadingSlash = configuredValue.startsWith("/")
		? configuredValue
		: `/${configuredValue}`;
	return withLeadingSlash.replace(/\/+/gu, "/").replace(/\/+$/u, "");
}

export function makeApiClient(
	apiBasePath: string,
	fetcher: Fetcher = fetch,
): ApiClient {
	const basePath = normalizeApiBasePath(apiBasePath);
	const apiUrl = (path: string) => `${basePath}${path}`;
	const articleUrl = (id: string) =>
		apiUrl(`/articles/${encodeURIComponent(id)}`);

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

		async listArticles() {
			const response = await fetcher(apiUrl("/articles"), {
				method: "GET",
				credentials: "include",
			});

			if (response.ok) {
				return (await response.json()) as ArticleListResponse;
			}

			throw await apiError(response, "Article list failed.");
		},

		async getArticle(id: string) {
			const response = await fetcher(articleUrl(id), {
				method: "GET",
				credentials: "include",
			});

			if (response.ok) {
				return (await response.json()) as ArticleDetail;
			}

			throw await apiError(response, "Article detail failed.");
		},

		async deleteArticle(id: string) {
			const response = await fetcher(articleUrl(id), {
				method: "DELETE",
				credentials: "include",
			});

			if (response.status === 204) {
				return;
			}

			throw await apiError(response, "Delete failed.");
		},

		async forceDeleteArticle(id: string) {
			const response = await fetcher(`${articleUrl(id)}/force`, {
				method: "DELETE",
				credentials: "include",
			});

			if (response.status === 204) {
				return;
			}

			throw await apiError(response, "Force delete failed.");
		},
	};
}

async function apiError(
	response: Response,
	fallbackMessage: string,
): Promise<Error> {
	if (response.status === 401) {
		return new ApiUnauthorizedError();
	}

	const message = await readErrorMessage(response, fallbackMessage);
	return new ApiRequestError(message);
}

async function readErrorMessage(
	response: Response,
	fallbackMessage: string,
): Promise<string> {
	try {
		const body = (await response.json()) as { error?: unknown };

		if (typeof body.error === "string" && body.error.trim()) {
			return body.error;
		}
	} catch {
		return fallbackMessage;
	}

	return fallbackMessage;
}
