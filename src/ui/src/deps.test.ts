import { makeAuthApiClient, normalizeApiBasePath } from "@archivist/deps.ts";
import { describe, expect, it, vi } from "vitest";

describe("normalizeApiBasePath", () => {
	it("defaults to /api", () => {
		expect(normalizeApiBasePath(undefined)).toBe("/api");
		expect(normalizeApiBasePath("")).toBe("/api");
	});

	it("normalizes leading and trailing slashes", () => {
		expect(normalizeApiBasePath("api/")).toBe("/api");
		expect(normalizeApiBasePath("/api///")).toBe("/api");
		expect(normalizeApiBasePath("//api")).toBe("/api");
		expect(normalizeApiBasePath("///api///")).toBe("/api");
		expect(normalizeApiBasePath("/edge/api/")).toBe("/edge/api");
		expect(normalizeApiBasePath("/edge//api///")).toBe("/edge/api");
		expect(normalizeApiBasePath("/")).toBe("");
	});
});

describe("auth api client", () => {
	it("uses the normalized internal API path and includes credentials", async () => {
		const fetcher = vi.fn(
			async (_input: RequestInfo | URL, init?: RequestInit) => {
				if (init?.method === "GET" && String(_input).includes("/articles")) {
					return new Response("{}", {
						status: 200,
						headers: { "Content-Type": "application/json" },
					});
				}

				return new Response(null, { status: 204 });
			},
		);
		const client = makeAuthApiClient("/edge//api///", fetcher);

		await expect(client.login("secret")).resolves.toBe(true);
		await expect(client.getSession()).resolves.toBe(true);
		await expect(client.logout()).resolves.toBe("ok");
		await expect(client.listArticles()).resolves.toEqual({});
		await expect(client.getArticle("01H/unsafe")).resolves.toEqual({});
		await expect(client.deleteArticle("01H/unsafe")).resolves.toBeUndefined();
		await expect(
			client.forceDeleteArticle("01H/unsafe"),
		).resolves.toBeUndefined();

		expect(fetcher).toHaveBeenNthCalledWith(1, "/edge/api/login", {
			method: "POST",
			credentials: "include",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify({ password: "secret" }),
		});
		expect(fetcher).toHaveBeenNthCalledWith(2, "/edge/api/auth/session", {
			method: "GET",
			credentials: "include",
		});
		expect(fetcher).toHaveBeenNthCalledWith(3, "/edge/api/logout", {
			method: "POST",
			credentials: "include",
		});
		expect(fetcher).toHaveBeenNthCalledWith(4, "/edge/api/articles", {
			method: "GET",
			credentials: "include",
		});
		expect(fetcher).toHaveBeenNthCalledWith(
			5,
			"/edge/api/articles/01H%2Funsafe",
			{
				method: "GET",
				credentials: "include",
			},
		);
		expect(fetcher).toHaveBeenNthCalledWith(
			6,
			"/edge/api/articles/01H%2Funsafe",
			{
				method: "DELETE",
				credentials: "include",
			},
		);
		expect(fetcher).toHaveBeenNthCalledWith(
			7,
			"/edge/api/articles/01H%2Funsafe/force",
			{
				method: "DELETE",
				credentials: "include",
			},
		);
	});

	it("uses public API error messages for failed article requests", async () => {
		const fetcher = vi.fn(
			async () =>
				new Response(JSON.stringify({ error: "Public article error." }), {
					status: 500,
					headers: { "Content-Type": "application/json" },
				}),
		);
		const client = makeAuthApiClient("/api", fetcher);

		await expect(client.getArticle("01H")).rejects.toThrow(
			"Public article error.",
		);
	});
});
