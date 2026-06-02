import type { ApiClient } from "@archivist/deps/api-client.ts";
import { ArticleDetailPane } from "@archivist/pages/articles/components/article-detail-pane.tsx";
import { ArticleMasterList } from "@archivist/pages/articles/components/article-master-list.tsx";
import { focusArticleShell } from "@archivist/pages/articles/components/article-view-helpers.ts";
import {
	DeleteModal,
	ForceDeleteModal,
} from "@archivist/pages/articles/components/delete-modals.tsx";
import { UserMenu } from "@archivist/pages/articles/components/user-menu.tsx";
import { useArticleDelete } from "@archivist/pages/articles/hooks/use-article-delete.ts";
import { useArticleDetail } from "@archivist/pages/articles/hooks/use-article-detail.ts";
import { useArticleList } from "@archivist/pages/articles/hooks/use-article-list.ts";
import { useCallback, useRef, useState } from "preact/hooks";
import { useLocation } from "preact-iso";

interface ArticleShellProps {
	api: ApiClient;
	onAuthExpired: () => void;
	onLogout: () => void;
	selectedArticleId?: string;
}

export function ArticleShell({
	api,
	onAuthExpired,
	onLogout,
	selectedArticleId,
}: ArticleShellProps) {
	const location = useLocation();
	const shellRef = useRef<HTMLElement>(null);
	const [isMenuOpen, setIsMenuOpen] = useState(false);
	const { articles, findArticle, isListLoading, listError, removeArticle } =
		useArticleList({ api, onAuthExpired });
	const {
		clearDetail,
		detailState,
		loadArticleDetail,
		selectedArticle,
		showDetailError,
	} = useArticleDetail({
		api,
		findArticle,
		onAuthExpired,
		selectedArticleId,
	});

	const selectedListItem = selectedArticleId
		? findArticle(selectedArticleId)
		: undefined;
	const actionArticle =
		selectedArticle ??
		(detailState.kind === "error" ? detailState.article : undefined) ??
		selectedListItem;

	const navigateToArticles = useCallback(() => {
		location.route("/articles", true);
	}, [location]);
	const focusShell = useCallback(() => {
		focusArticleShell(shellRef);
	}, []);
	const {
		closeDeleteModal,
		closeForceDeleteModal,
		confirmDelete,
		confirmForceDelete,
		isDeleteModalOpen,
		isDeleting,
		isForceDeleteModalOpen,
		isForceDeleting,
		openDeleteModal,
		openForceDeleteModal,
	} = useArticleDelete({
		actionArticle,
		api,
		clearDetail,
		focusShell,
		navigateToArticles,
		onAuthExpired,
		removeArticle,
		selectedArticleId,
		showDetailError,
	});

	function selectArticle(articleId: string) {
		void loadArticleDetail(articleId);
		location.route(`/articles/${articleId}`);
	}

	return (
		<main className="article-shell" ref={shellRef} tabIndex={-1}>
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
			<section className="article-workspace" aria-label="Articles">
				<ArticleMasterList
					articles={articles}
					error={listError}
					isLoading={isListLoading}
					onSelect={selectArticle}
					selectedArticleId={selectedArticleId}
				/>
				<ArticleDetailPane
					article={actionArticle}
					detailState={detailState}
					onDelete={openDeleteModal}
					onForceDelete={openForceDeleteModal}
				/>
			</section>
			<footer className="article-footer">
				VERSION: {import.meta.env.VITE_VERSION_LABEL}
			</footer>
			{isDeleteModalOpen ? (
				<DeleteModal
					isDeleting={isDeleting}
					onCancel={closeDeleteModal}
					onConfirm={confirmDelete}
				/>
			) : null}
			{isForceDeleteModalOpen ? (
				<ForceDeleteModal
					isDeleting={isForceDeleting}
					onCancel={closeForceDeleteModal}
					onConfirm={confirmForceDelete}
				/>
			) : null}
		</main>
	);
}
