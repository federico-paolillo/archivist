import type { Deps } from "@archivist/deps.ts";
import { LoginForm } from "@archivist/pages/login/components/login-form.tsx";
import { useEffect, useState } from "preact/hooks";
import { useLocation } from "preact-iso";

interface LoginPageProps {
	deps: Deps;
}

export function LoginPage({ deps }: LoginPageProps) {
	const location = useLocation();
	const [password, setPassword] = useState("");
	const [isSubmitting, setIsSubmitting] = useState(false);

	async function submitLogin(event: SubmitEvent) {
		event.preventDefault();
		setIsSubmitting(true);

		try {
			const isAuthenticated = await deps.api.login(password);
			setPassword("");
			location.route(isAuthenticated ? "/articles" : "/login/failed", true);
		} catch {
			setPassword("");
			location.route("/login/failed", true);
		} finally {
			setIsSubmitting(false);
		}
	}

	return (
		<main className="login-page">
			<LoginForm
				isSubmitting={isSubmitting}
				onPasswordInput={setPassword}
				onSubmit={submitLogin}
				password={password}
			/>
		</main>
	);
}

export function LoginRedirect() {
	const location = useLocation();

	useEffect(() => {
		location.route("/login", true);
	}, [location]);

	return <div className="blank-page" aria-hidden="true" />;
}
