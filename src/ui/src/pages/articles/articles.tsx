import { useLocation } from "preact-iso";
import { ProtectedRoute } from "@archivist/components/protected-route.tsx";
import type { Deps } from "@archivist/deps.ts";
import { ArticleShell } from "@archivist/pages/articles/components/article-shell.tsx";

interface ArticlesPageProps {
	deps: Deps;
}

export function ArticlesPage({ deps }: ArticlesPageProps) {
	const location = useLocation();

	async function logout() {
		const result = await deps.api.logout();

		if (result === "ok" || result === "unauthorized") {
			location.route("/login", true);
		}
	}

	return (
		<ProtectedRoute deps={deps}>
			<ArticleShell onLogout={logout} />
		</ProtectedRoute>
	);
}
