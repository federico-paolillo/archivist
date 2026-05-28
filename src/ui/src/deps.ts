export type Fetcher = typeof fetch;

export interface ArticleMetadata {
	id: string;
	title: string | null;
	originalUrl: string | null;
	canonicalUrl: string | null;
	status: string;
	errorMessage: string | null;
	createdAt: string;
}

export interface ArticleDetail extends ArticleMetadata {
	summaryMarkdown: string | null;
	contentMarkdown: string | null;
}

export interface ArticleListResponse {
	items: ArticleMetadata[];
	nextCursor: string | null;
	previousCursor: string | null;
}

export interface AuthApiClient {
	getSession: () => Promise<boolean>;
	login: (password: string) => Promise<boolean>;
	logout: () => Promise<"ok" | "unauthorized" | "failed">;
}

export interface ArticleApiClient {
	listArticles: () => Promise<ArticleListResponse>;
	getArticle: (id: string) => Promise<ArticleDetail>;
	deleteArticle: (id: string) => Promise<void>;
}

export type ApiClient = AuthApiClient & ArticleApiClient;

export interface Deps {
	api: ApiClient;
}

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
	const rawValue = configuredValue || "/api";
	const withLeadingSlash = rawValue.startsWith("/") ? rawValue : `/${rawValue}`;
	const sameOriginPath = withLeadingSlash.replace(/^\/+/u, "/");
	const normalized = sameOriginPath.replace(/\/+$/u, "");

	return normalized || (configuredValue ? "" : "/api");
}

export function makeAuthApiClient(
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

export function makeDeps(): Deps {
	return {
		api: makeAuthApiClient(import.meta.env.VITE_API_BASE_PATH),
	};
}
