import { ErrorBoundary, LocationProvider, Route, Router } from "preact-iso";
import type { Deps } from "@archivist/deps.ts";
import { ArticlesPage } from "@archivist/pages/articles/articles.tsx";
import { LoginFailedPage } from "@archivist/pages/login-failed/login-failed.tsx";
import { LoginPage, LoginRedirect } from "@archivist/pages/login/login.tsx";

interface AppProps {
	deps: Deps;
}

export function App({ deps }: AppProps) {
	return (
		<LocationProvider>
			<ErrorBoundary>
				<Router>
					<Route path="/login" component={LoginPage} deps={deps} />
					<Route path="/login/failed" component={LoginFailedPage} />
					<Route path="/articles" component={ArticlesPage} deps={deps} />
					<Route
						path="/articles/:articleId"
						component={ArticlesPage}
						deps={deps}
					/>
					<Route default component={LoginRedirect} />
				</Router>
			</ErrorBoundary>
		</LocationProvider>
	);
}
