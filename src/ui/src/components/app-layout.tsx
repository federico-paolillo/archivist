import type { ComponentChildren, Ref } from "preact";

interface AppLayoutProps {
	children: ComponentChildren;
	className?: string;
	headerEnd?: ComponentChildren;
	mainClassName?: string;
	rootRef?: Ref<HTMLDivElement> | undefined;
	rootTabIndex?: number | undefined;
}

function classNames(...names: Array<string | undefined>) {
	return names.filter(Boolean).join(" ");
}

export function AppLayout({
	children,
	className,
	headerEnd,
	mainClassName,
	rootRef,
	rootTabIndex,
}: AppLayoutProps) {
	return (
		<div
			className={classNames("app-layout", className)}
			{...(rootRef ? { ref: rootRef } : {})}
			{...(rootTabIndex !== undefined ? { tabIndex: rootTabIndex } : {})}
		>
			<header className="top-bar">
				<a className="brand-link" href="/articles">
					Archivist
				</a>
				{headerEnd}
			</header>
			<main className={classNames("app-main", mainClassName)}>{children}</main>
			<footer className="app-footer article-footer">
				VERSION: {import.meta.env.VITE_VERSION_LABEL}
			</footer>
		</div>
	);
}
