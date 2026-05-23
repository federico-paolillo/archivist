import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { App } from "@archivist/app.tsx";
import type { Deps } from "@archivist/deps.ts";

function makeTestDeps(overrides: Partial<Deps["api"]> = {}): Deps {
	return {
		api: {
			getSession: vi.fn(async () => true),
			login: vi.fn(async () => true),
			logout: vi.fn(async () => "ok" as const),
			...overrides,
		},
	};
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
