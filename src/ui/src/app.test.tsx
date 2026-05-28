import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { App } from "@archivist/app.tsx";
import {
	ApiRequestError,
	ApiUnauthorizedError,
	type ArticleDetail,
	type ArticleMetadata,
	type Deps,
} from "@archivist/deps.ts";

function makeTestDeps(overrides: Partial<Deps["api"]> = {}): Deps {
	return {
		api: {
			getSession: vi.fn(async () => true),
			login: vi.fn(async () => true),
			logout: vi.fn(async () => "ok" as const),
			listArticles: vi.fn(async () => ({
				items: articles,
				nextCursor: null,
				previousCursor: null,
			})),
			getArticle: vi.fn(
				async (id: string) => articleDetails[id] ?? readyArticle,
			),
			deleteArticle: vi.fn(async () => undefined),
			...overrides,
		},
	};
}

const readyArticle: ArticleDetail = {
	id: "01HREADY000000000000000000",
	title: "Analysis of Subterranean Anomalies in Sector 4",
	originalUrl: "https://example.com/original",
	canonicalUrl: "https://example.com/canonical",
	status: "ready",
	errorMessage: null,
	createdAt: "2026-05-06T12:00:00Z",
	summaryMarkdown:
		"Recent seismic sensors registered **anomalous** vibrations. <img src=x onerror=alert(1)> [bad](javascript:alert(1)) [good](https://example.com/good)",
	contentMarkdown:
		"## Sensor Data Acquisition\n\n<script>alert(1)</script>\n\nData was collected from nodes S4-01 through S4-15.",
};

const queuedArticle: ArticleDetail = {
	id: "01HQUEUED00000000000000000",
	title: "Queued article",
	originalUrl: "https://example.com/queued",
	canonicalUrl: null,
	status: "queued",
	errorMessage: null,
	createdAt: "2026-05-06T13:00:00Z",
	summaryMarkdown: null,
	contentMarkdown: null,
};

const failedArticle: ArticleDetail = {
	id: "01HFAILED00000000000000000",
	title: "Failed article",
	originalUrl: "https://example.com/failed-original",
	canonicalUrl: "https://example.com/failed-canonical",
	status: "failed",
	errorMessage: "ARC-100: extraction failed.",
	createdAt: "2026-05-06T14:00:00Z",
	summaryMarkdown: null,
	contentMarkdown: null,
};

const articles: ArticleMetadata[] = [
	readyArticle,
	queuedArticle,
	failedArticle,
];

const articleDetails: Record<string, ArticleDetail> = {
	[readyArticle.id]: readyArticle,
	[queuedArticle.id]: queuedArticle,
	[failedArticle.id]: failedArticle,
};

function deferred<T>() {
	let resolve!: (value: T) => void;
	let reject!: (error: unknown) => void;
	const promise = new Promise<T>((promiseResolve, promiseReject) => {
		resolve = promiseResolve;
		reject = promiseReject;
	});

	return { promise, reject, resolve };
}

function mountAt(path: string, deps: Deps) {
	window.history.replaceState(null, "", path);
	const root = document.createElement("div");
	document.body.appendChild(root);

	act(() => {
		render(<App deps={deps} />, root);
	});
}

afterEach(() => {
	render(null, document.body);
	document.body.replaceChildren();
	window.history.replaceState(null, "", "/");
	vi.restoreAllMocks();
});

describe("auth routes", () => {
	it("navigates to /articles after successful login", async () => {
		const deps = makeTestDeps({
			login: vi.fn(async () => true),
			getSession: vi.fn(async () => true),
		});
		const user = userEvent.setup();

		mountAt("/login", deps);

		await user.type(screen.getByLabelText("Password"), "correct-password");
		await user.click(screen.getByRole("button", { name: "IDENTIFY" }));

		await waitFor(() => {
			expect(window.location.pathname).toBe("/articles");
		});
		expect(deps.api.login).toHaveBeenCalledWith("correct-password");
	});

	it("navigates invalid login to a blank black route", async () => {
		const deps = makeTestDeps({
			login: vi.fn(async () => false),
		});
		const user = userEvent.setup();

		mountAt("/login", deps);

		await user.type(screen.getByLabelText("Password"), "wrong-password");
		await user.click(screen.getByRole("button", { name: "IDENTIFY" }));

		await waitFor(() => {
			expect(window.location.pathname).toBe("/login/failed");
		});
		expect(document.body.textContent).toBe("");
		expect(document.querySelector(".blank-page")).not.toBeNull();
	});

	it("renders /login/failed as a blank black page", async () => {
		mountAt("/login/failed", makeTestDeps());

		expect(document.body.textContent).toBe("");
		expect(document.querySelector(".blank-page")).not.toBeNull();
	});

	it("redirects article routes to /login on session 401", async () => {
		const deps = makeTestDeps({
			getSession: vi.fn(async () => false),
		});

		mountAt("/articles", deps);

		await waitFor(() => {
			expect(window.location.pathname).toBe("/login");
		});
	});

	it("logs out from the article shell", async () => {
		const deps = makeTestDeps({
			getSession: vi.fn(async () => true),
			logout: vi.fn(async () => "ok" as const),
		});
		const user = userEvent.setup();

		mountAt("/articles/01ABC", deps);

		await waitFor(() => {
			expect(document.querySelector(".article-shell")).not.toBeNull();
		});

		await user.click(screen.getByRole("button", { name: "User menu" }));

		await waitFor(() => {
			expect(screen.getByRole("button", { name: "Logout" })).not.toBeNull();
		});
		await user.click(screen.getByRole("button", { name: "Logout" }));

		await waitFor(() => {
			expect(window.location.pathname).toBe("/login");
		});
		expect(deps.api.logout).toHaveBeenCalledTimes(1);
	});
});

describe("article routes", () => {
	it("renders /articles with a master list and blank detail pane", async () => {
		const deps = makeTestDeps();

		mountAt("/articles", deps);

		expect(await screen.findByText(readyArticle.title ?? "")).not.toBeNull();
		const blankDetail = document.querySelector(".article-detail-blank");
		expect(blankDetail).not.toBeNull();
		expect(blankDetail?.textContent).toBe("");
		expect(deps.api.getArticle).not.toHaveBeenCalled();
	});

	it("navigates immediately when selecting a row and shows detail loading", async () => {
		const detail = deferred<ArticleDetail>();
		const deps = makeTestDeps({
			getArticle: vi.fn(async () => detail.promise),
		});
		const user = userEvent.setup();

		mountAt("/articles", deps);

		await user.click(await screen.findByText(readyArticle.title ?? ""));

		expect(window.location.pathname).toBe(`/articles/${readyArticle.id}`);
		expect(screen.getByLabelText("Loading article detail")).not.toBeNull();

		await act(async () => {
			detail.resolve(readyArticle);
			await detail.promise;
		});

		expect(await screen.findByText("Sensor Data Acquisition")).not.toBeNull();
	});

	it("renders ready detail with safe markdown, Original, and Delete", async () => {
		const deps = makeTestDeps();

		mountAt(`/articles/${readyArticle.id}`, deps);

		expect((await screen.findByRole("heading", { level: 1 })).textContent).toBe(
			readyArticle.title,
		);
		expect(screen.getByText("anomalous")).not.toBeNull();
		expect(screen.getByRole("button", { name: "Delete" })).not.toBeNull();

		const original = screen.getByRole("link", { name: "Original" });
		expect(original.getAttribute("href")).toBe(readyArticle.canonicalUrl);
		expect(original.getAttribute("rel")).toBe("noopener noreferrer");
		expect(original.getAttribute("target")).toBe("_blank");

		const renderedGoodLink = screen.getByRole("link", { name: "good" });
		expect(renderedGoodLink.getAttribute("rel")).toBe("noopener noreferrer");
		expect(document.querySelector("script")).toBeNull();
		expect(document.querySelector("img")).toBeNull();
		expect(document.querySelector('a[href^="javascript:"]')).toBeNull();
	});

	it("renders queued and future non-terminal detail states as come-back-later", async () => {
		const deps = makeTestDeps();

		mountAt(`/articles/${queuedArticle.id}`, deps);

		expect(await screen.findByText("Come back later.")).not.toBeNull();
		expect(screen.getByRole("button", { name: "Delete" })).not.toBeNull();
		expect(
			screen.getByRole("link", { name: "Original" }).getAttribute("href"),
		).toBe(queuedArticle.originalUrl);
	});

	it("renders failed detail with the persisted error message", async () => {
		const deps = makeTestDeps();

		mountAt(`/articles/${failedArticle.id}`, deps);

		expect(
			await screen.findByText(failedArticle.errorMessage ?? ""),
		).not.toBeNull();
		expect(screen.getByRole("button", { name: "Delete" })).not.toBeNull();
		expect(
			screen.getByRole("link", { name: "Original" }).getAttribute("href"),
		).toBe(failedArticle.canonicalUrl);
	});

	it("renders detail fetch failures centered in the detail pane", async () => {
		const deps = makeTestDeps({
			getArticle: vi.fn(async () => {
				throw new ApiRequestError("Article detail unavailable.");
			}),
		});

		mountAt(`/articles/${readyArticle.id}`, deps);

		expect(
			await screen.findByText("Article detail unavailable."),
		).not.toBeNull();
		expect(document.querySelector(".detail-message-error")).not.toBeNull();
	});

	it("redirects to /login when an article API call returns 401", async () => {
		const deps = makeTestDeps({
			listArticles: vi.fn(async () => {
				throw new ApiUnauthorizedError();
			}),
		});

		mountAt("/articles", deps);

		await waitFor(() => {
			expect(window.location.pathname).toBe("/login");
		});
	});

	it("does not delete when the confirmation modal is cancelled", async () => {
		const deps = makeTestDeps({
			deleteArticle: vi.fn(async () => undefined),
		});
		const user = userEvent.setup();

		mountAt(`/articles/${readyArticle.id}`, deps);

		await user.click(await screen.findByRole("button", { name: "Delete" }));
		expect(screen.getByText("Are you sure?")).not.toBeNull();
		await user.click(screen.getByRole("button", { name: "Nevermind" }));

		expect(deps.api.deleteArticle).not.toHaveBeenCalled();
		expect(window.location.pathname).toBe(`/articles/${readyArticle.id}`);
		expect(screen.queryByText("Are you sure?")).toBeNull();
	});

	it("deletes on confirmation, removes the item, and clears detail", async () => {
		const deps = makeTestDeps({
			deleteArticle: vi.fn(async () => undefined),
		});
		const user = userEvent.setup();

		mountAt(`/articles/${readyArticle.id}`, deps);

		await user.click(await screen.findByRole("button", { name: "Delete" }));
		await user.click(screen.getByRole("button", { name: "Yes" }));

		await waitFor(() => {
			expect(window.location.pathname).toBe("/articles");
		});
		expect(deps.api.deleteArticle).toHaveBeenCalledWith(readyArticle.id);
		expect(screen.queryByText(readyArticle.title ?? "")).toBeNull();
		expect(document.querySelector(".article-detail-blank")).not.toBeNull();
	});

	it("leaves the article selected and renders delete failures", async () => {
		const deps = makeTestDeps({
			deleteArticle: vi.fn(async () => {
				throw new ApiRequestError("Delete rejected.");
			}),
		});
		const user = userEvent.setup();

		mountAt(`/articles/${readyArticle.id}`, deps);

		await user.click(await screen.findByRole("button", { name: "Delete" }));
		await user.click(screen.getByRole("button", { name: "Yes" }));

		expect(await screen.findByText("Delete rejected.")).not.toBeNull();
		expect(window.location.pathname).toBe(`/articles/${readyArticle.id}`);
		expect(screen.getByText(readyArticle.title ?? "")).not.toBeNull();
	});
});
