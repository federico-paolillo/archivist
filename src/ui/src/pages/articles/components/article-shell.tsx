import { useEffect, useRef, useState } from "preact/hooks";
import { useLocation } from "preact-iso";
import {
	ApiUnauthorizedError,
	type ApiClient,
	type ArticleDetail,
	type ArticleMetadata,
} from "@archivist/deps.ts";
import { UserMenu } from "@archivist/pages/articles/components/user-menu.tsx";
import { MarkdownContent } from "./markdown-content.tsx";

interface ArticleShellProps {
	api: ApiClient;
	onAuthExpired: () => void;
	onLogout: () => void;
	selectedArticleId?: string;
}

type DetailState =
	| { kind: "idle" }
	| { kind: "loading" }
	| { article?: ArticleMetadata; kind: "error"; message: string }
	| { article: ArticleDetail; kind: "loaded" };

export function ArticleShell({
	api,
	onAuthExpired,
	onLogout,
	selectedArticleId,
}: ArticleShellProps) {
	const location = useLocation();
	const [isMenuOpen, setIsMenuOpen] = useState(false);
	const [articles, setArticles] = useState<ArticleMetadata[]>([]);
	const articlesRef = useRef<ArticleMetadata[]>([]);
	const [listError, setListError] = useState<string | null>(null);
	const [isListLoading, setIsListLoading] = useState(true);
	const [detailState, setDetailState] = useState<DetailState>({ kind: "idle" });
	const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
	const [isDeleting, setIsDeleting] = useState(false);
	const detailRequestRef = useRef(0);
	const loadedArticleIdRef = useRef<string | null>(null);

	async function loadArticleDetail(articleId: string) {
		const requestId = detailRequestRef.current + 1;
		detailRequestRef.current = requestId;
		loadedArticleIdRef.current = articleId;
		setDetailState({ kind: "loading" });

		try {
			const article = await api.getArticle(articleId);

			if (detailRequestRef.current === requestId) {
				setDetailState({ kind: "loaded", article });
			}
		} catch (error) {
			if (detailRequestRef.current !== requestId) {
				return;
			}

			if (error instanceof ApiUnauthorizedError) {
				onAuthExpired();
				return;
			}

			const article = articlesRef.current.find((item) => item.id === articleId);
			setDetailState({
				kind: "error",
				article,
				message: errorMessage(error, "Article detail failed."),
			});
		}
	}

	useEffect(() => {
		let isCurrent = true;

		async function loadArticles() {
			setIsListLoading(true);
			setListError(null);

			try {
				const response = await api.listArticles();

				if (isCurrent) {
					articlesRef.current = response.items;
					setArticles(response.items);
				}
			} catch (error) {
				if (!isCurrent) {
					return;
				}

				if (error instanceof ApiUnauthorizedError) {
					onAuthExpired();
					return;
				}

				setListError(errorMessage(error, "Article list failed."));
			} finally {
				if (isCurrent) {
					setIsListLoading(false);
				}
			}
		}

		loadArticles();

		return () => {
			isCurrent = false;
		};
	}, [api, onAuthExpired]);

	useEffect(() => {
		if (!selectedArticleId) {
			loadedArticleIdRef.current = null;
			setDetailState({ kind: "idle" });
			setIsDeleteModalOpen(false);
			return;
		}

		if (loadedArticleIdRef.current !== selectedArticleId) {
			void loadArticleDetail(selectedArticleId);
		}
	}, [selectedArticleId]);

	const selectedArticle =
		detailState.kind === "loaded" ? detailState.article : undefined;
	const selectedListItem = articles.find(
		(item) => item.id === selectedArticleId,
	);
	const actionArticle =
		selectedArticle ??
		(detailState.kind === "error" ? detailState.article : undefined) ??
		selectedListItem;

	function selectArticle(articleId: string) {
		void loadArticleDetail(articleId);
		location.route(`/articles/${articleId}`);
	}

	async function confirmDelete() {
		if (!selectedArticleId) {
			return;
		}

		setIsDeleting(true);

		try {
			await api.deleteArticle(selectedArticleId);
			articlesRef.current = articlesRef.current.filter(
				(article) => article.id !== selectedArticleId,
			);
			setArticles((currentArticles) =>
				currentArticles.filter((article) => article.id !== selectedArticleId),
			);
			setIsDeleteModalOpen(false);
			setDetailState({ kind: "idle" });
			location.route("/articles", true);
		} catch (error) {
			if (error instanceof ApiUnauthorizedError) {
				onAuthExpired();
				return;
			}

			setIsDeleteModalOpen(false);
			setDetailState({
				kind: "error",
				article: actionArticle,
				message: errorMessage(error, "Delete failed."),
			});
		} finally {
			setIsDeleting(false);
		}
	}

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
					onDelete={() => {
						setIsDeleteModalOpen(true);
					}}
				/>
			</section>
			<footer className="article-footer">VERSION: 7F8A2C1_STABLE</footer>
			{isDeleteModalOpen ? (
				<DeleteModal
					isDeleting={isDeleting}
					onCancel={() => {
						setIsDeleteModalOpen(false);
					}}
					onConfirm={confirmDelete}
				/>
			) : null}
		</main>
	);
}

interface ArticleMasterListProps {
	articles: ArticleMetadata[];
	error: string | null;
	isLoading: boolean;
	onSelect: (articleId: string) => void;
	selectedArticleId?: string;
}

function ArticleMasterList({
	articles,
	error,
	isLoading,
	onSelect,
	selectedArticleId,
}: ArticleMasterListProps) {
	return (
		<aside className="article-master" aria-label="Article list">
			{isLoading ? <p className="master-message">Loading archive.</p> : null}
			{error ? (
				<p className="master-message master-message-error">{error}</p>
			) : null}
			{!isLoading && !error && articles.length === 0 ? (
				<p className="master-message">No articles.</p>
			) : null}
			{articles.map((article) => (
				<button
					className={
						article.id === selectedArticleId
							? "article-row article-row-selected"
							: "article-row"
					}
					key={article.id}
					type="button"
					onClick={() => {
						onSelect(article.id);
					}}
				>
					<span className="article-row-id">ID: {article.id}</span>
					<span className="article-row-title">
						{article.title?.trim() || "Untitled article"}
					</span>
					<span className="article-row-url">
						{article.canonicalUrl || article.originalUrl || "No source URL"}
					</span>
					<span className="article-row-status">{article.status}</span>
				</button>
			))}
		</aside>
	);
}

interface ArticleDetailPaneProps {
	article?: ArticleMetadata;
	detailState: DetailState;
	onDelete: () => void;
}

function ArticleDetailPane({
	article,
	detailState,
	onDelete,
}: ArticleDetailPaneProps) {
	if (detailState.kind === "idle") {
		return <section className="article-detail article-detail-blank" />;
	}

	if (detailState.kind === "loading") {
		return (
			<section className="article-detail article-detail-centered">
				<div
					className="loading-spinner"
					role="status"
					aria-label="Loading article detail"
				/>
			</section>
		);
	}

	const loadedArticle =
		detailState.kind === "loaded" ? detailState.article : undefined;
	const displayArticle = loadedArticle ?? article;

	return (
		<section className="article-detail">
			{displayArticle ? (
				<ArticleActions article={displayArticle} onDelete={onDelete} />
			) : null}
			{detailState.kind === "error" ? (
				<div className="detail-message detail-message-error">
					{detailState.message}
				</div>
			) : null}
			{loadedArticle ? <LoadedArticleDetail article={loadedArticle} /> : null}
		</section>
	);
}

interface ArticleActionsProps {
	article: ArticleMetadata;
	onDelete: () => void;
}

function ArticleActions({ article, onDelete }: ArticleActionsProps) {
	const originalUrl = article.canonicalUrl || article.originalUrl;

	return (
		<div className="article-actions">
			<button className="button-outline" type="button" onClick={onDelete}>
				Delete
			</button>
			{originalUrl ? (
				<a
					className="button-outline"
					href={originalUrl}
					target="_blank"
					rel="noopener noreferrer"
				>
					Original
				</a>
			) : null}
		</div>
	);
}

interface LoadedArticleDetailProps {
	article: ArticleDetail;
}

function LoadedArticleDetail({ article }: LoadedArticleDetailProps) {
	if (article.status === "ready") {
		return (
			<article className="ready-article">
				<header className="detail-hero">
					<h1>{article.title?.trim() || "Untitled article"}</h1>
				</header>
				<section className="summary-panel" aria-label="Summary">
					<p className="section-label">Summary</p>
					<MarkdownContent markdown={article.summaryMarkdown || ""} />
				</section>
				<section className="content-panel" aria-label="Content">
					<MarkdownContent markdown={article.contentMarkdown || ""} />
				</section>
			</article>
		);
	}

	if (article.status === "failed") {
		return (
			<div className="detail-message detail-message-error">
				{article.errorMessage || "Article failed."}
			</div>
		);
	}

	return <div className="detail-message">Come back later.</div>;
}

interface DeleteModalProps {
	isDeleting: boolean;
	onCancel: () => void;
	onConfirm: () => void;
}

function DeleteModal({ isDeleting, onCancel, onConfirm }: DeleteModalProps) {
	return (
		<div className="modal-backdrop" role="presentation">
			<div
				aria-labelledby="delete-modal-title"
				aria-modal="true"
				className="delete-modal"
				role="dialog"
			>
				<p id="delete-modal-title">Are you sure?</p>
				<div className="delete-modal-actions">
					<button
						className="button-outline"
						type="button"
						disabled={isDeleting}
						onClick={onConfirm}
					>
						Yes
					</button>
					<button
						className="button-outline"
						type="button"
						disabled={isDeleting}
						onClick={onCancel}
					>
						Nevermind
					</button>
				</div>
			</div>
		</div>
	);
}

function errorMessage(error: unknown, fallbackMessage: string): string {
	return error instanceof Error ? error.message : fallbackMessage;
}
