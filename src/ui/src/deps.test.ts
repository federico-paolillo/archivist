import { describe, expect, it, vi } from "vitest";
import { makeAuthApiClient, normalizeApiBasePath } from "@archivist/deps.ts";

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
		expect(normalizeApiBasePath("/")).toBe("");
	});
});

describe("auth api client", () => {
	it("uses the normalized API base and includes credentials", async () => {
		const fetcher = vi.fn(async () => new Response(null, { status: 204 }));
		const client = makeAuthApiClient("/api///", fetcher);

		await expect(client.login("secret")).resolves.toBe(true);
		await expect(client.getSession()).resolves.toBe(true);
		await expect(client.logout()).resolves.toBe("ok");

		expect(fetcher).toHaveBeenNthCalledWith(1, "/api/login", {
			method: "POST",
			credentials: "include",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify({ password: "secret" }),
		});
		expect(fetcher).toHaveBeenNthCalledWith(2, "/api/auth/session", {
			method: "GET",
			credentials: "include",
		});
		expect(fetcher).toHaveBeenNthCalledWith(3, "/api/logout", {
			method: "POST",
			credentials: "include",
		});
	});
});
