import type { ComponentChildren } from "preact";
import { useEffect, useState } from "preact/hooks";
import { useLocation } from "preact-iso";
import type { Deps } from "@archivist/deps.ts";

interface ProtectedRouteProps {
	children: ComponentChildren;
	deps: Deps;
}

export function ProtectedRoute({ children, deps }: ProtectedRouteProps) {
	const location = useLocation();
	const [sessionState, setSessionState] = useState<
		"checking" | "authenticated"
	>("checking");

	useEffect(() => {
		let isCurrent = true;

		async function checkSession() {
			try {
				const isAuthenticated = await deps.api.getSession();
				if (!isCurrent) {
					return;
				}

				if (isAuthenticated) {
					setSessionState("authenticated");
					return;
				}

				location.route("/login", true);
			} catch {
				if (isCurrent) {
					location.route("/login", true);
				}
			}
		}

		checkSession();

		return () => {
			isCurrent = false;
		};
	}, [deps, location]);

	if (sessionState === "checking") {
		return <main className="blank-page" aria-label="Checking session" />;
	}

	return <>{children}</>;
}
