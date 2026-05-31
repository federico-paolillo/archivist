import { ProtectedRoute } from "@archivist/components/protected-route.tsx";
import type { Deps } from "@archivist/deps.ts";
import { ArticleShell } from "@archivist/pages/articles/components/article-shell.tsx";
import { useLocation } from "preact-iso";

interface ArticlesPageProps {
	articleId?: string;
	deps: Deps;
}

export function ArticlesPage({ articleId, deps }: ArticlesPageProps) {
	const location = useLocation();

	async function logout() {
		const result = await deps.api.logout();

		if (result === "ok" || result === "unauthorized") {
			location.route("/login", true);
		}
	}

	return (
		<ProtectedRoute deps={deps}>
			<ArticleShell
				api={deps.api}
				onAuthExpired={() => {
					location.route("/login", true);
				}}
				onLogout={logout}
				selectedArticleId={articleId}
			/>
		</ProtectedRoute>
	);
}
