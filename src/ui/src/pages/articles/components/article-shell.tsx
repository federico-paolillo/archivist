import { useState } from "preact/hooks";
import { UserMenu } from "@archivist/pages/articles/components/user-menu.tsx";

interface ArticleShellProps {
	onLogout: () => void;
}

export function ArticleShell({ onLogout }: ArticleShellProps) {
	const [isMenuOpen, setIsMenuOpen] = useState(false);

	return (
		<main className="article-shell">
			<header className="top-bar">
				<a className="brand-link" href="/articles">
					Archivist
				</a>
				<UserMenu
					isOpen={isMenuOpen}
					onLogout={onLogout}
					onToggle={() => {
						setIsMenuOpen((isOpen) => !isOpen);
					}}
				/>
			</header>
			<section className="article-placeholder" aria-label="Articles">
				<div className="placeholder-column" />
				<div className="placeholder-detail" />
			</section>
		</main>
	);
}
