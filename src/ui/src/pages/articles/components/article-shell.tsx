import {
	type ApiClient,
	ApiUnauthorizedError,
	type ArticleDetail,
	type ArticleMetadata,
} from "@archivist/deps.ts";
import { UserMenu } from "@archivist/pages/articles/components/user-menu.tsx";
import { useEffect, useRef, useState } from "preact/hooks";
import { useLocation } from "preact-iso";
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
	| {
			article?: ArticleMetadata | ArticleDetail;
			kind: "error";
			message: string;
	  }
	| { article: ArticleDetail; kind: "loaded" };

export function ArticleShell({
	api,
	onAuthExpired,
	onLogout,
	selectedArticleId,
}: ArticleShellProps) {
	const location = useLocation();
	const shellRef = useRef<HTMLElement>(null);
	const [isMenuOpen, setIsMenuOpen] = useState(false);
	const [articles, setArticles] = useState<ArticleMetadata[]>([]);
	const articlesRef = useRef<ArticleMetadata[]>([]);
	const [listError, setListError] = useState<string | null>(null);
	const [isListLoading, setIsListLoading] = useState(true);
	const [detailState, setDetailState] = useState<DetailState>({ kind: "idle" });
	const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
	const [isDeleting, setIsDeleting] = useState(false);
	const [isForceDeleteModalOpen, setIsForceDeleteModalOpen] = useState(false);
	const [isForceDeleting, setIsForceDeleting] = useState(false);
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
			setIsForceDeleteModalOpen(false);
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

		const articleId = selectedArticleId;
		setIsDeleting(true);

		try {
			await api.deleteArticle(articleId);
			articlesRef.current = articlesRef.current.filter(
				(article) => article.id !== articleId,
			);
			setArticles((currentArticles) =>
				currentArticles.filter((article) => article.id !== articleId),
			);
			setIsDeleteModalOpen(false);
			loadedArticleIdRef.current = null;
			setDetailState({ kind: "idle" });
			location.route("/articles", true);
			focusArticleShell(shellRef);
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

	async function confirmForceDelete() {
		if (!selectedArticleId) {
			return;
		}

		const articleId = selectedArticleId;
		setIsForceDeleting(true);

		try {
			await api.forceDeleteArticle(articleId);
			articlesRef.current = articlesRef.current.filter(
				(article) => article.id !== articleId,
			);
			setArticles((currentArticles) =>
				currentArticles.filter((article) => article.id !== articleId),
			);
			setIsForceDeleteModalOpen(false);
			loadedArticleIdRef.current = null;
			setDetailState({ kind: "idle" });
			location.route("/articles", true);
			focusArticleShell(shellRef);
		} catch (error) {
			if (error instanceof ApiUnauthorizedError) {
				onAuthExpired();
				return;
			}

			setIsForceDeleteModalOpen(false);
			setDetailState({
				kind: "error",
				article: actionArticle,
				message: errorMessage(error, "Force delete failed."),
			});
		} finally {
			setIsForceDeleting(false);
		}
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
					onDelete={() => {
						setIsDeleteModalOpen(true);
					}}
					onForceDelete={() => {
						setIsForceDeleteModalOpen(true);
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
			{isForceDeleteModalOpen ? (
				<ForceDeleteModal
					isDeleting={isForceDeleting}
					onCancel={() => {
						setIsForceDeleteModalOpen(false);
					}}
					onConfirm={confirmForceDelete}
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
	article?: ArticleMetadata | ArticleDetail;
	detailState: DetailState;
	onDelete: () => void;
	onForceDelete: () => void;
}

function ArticleDetailPane({
	article,
	detailState,
	onDelete,
	onForceDelete,
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
				<ArticleActions
					article={displayArticle}
					onDelete={onDelete}
					onForceDelete={onForceDelete}
				/>
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
	article: ArticleMetadata | ArticleDetail;
	onDelete: () => void;
	onForceDelete: () => void;
}

function ArticleActions({
	article,
	onDelete,
	onForceDelete,
}: ArticleActionsProps) {
	const originalUrl = article.canonicalUrl || article.originalUrl;
	const canForceDelete = isArticleDetail(article) && article.canForceDelete;

	return (
		<div className="article-actions">
			<button className="button-outline" type="button" onClick={onDelete}>
				Delete
			</button>
			{canForceDelete ? (
				<button
					className="button-outline button-danger"
					type="button"
					onClick={onForceDelete}
				>
					Force Delete
				</button>
			) : null}
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

function isArticleDetail(
	article: ArticleMetadata | ArticleDetail,
): article is ArticleDetail {
	return "canForceDelete" in article;
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
		<ConfirmationModal
			confirmClassName="button-outline"
			confirmText="Yes"
			isDeleting={isDeleting}
			title="Are you sure?"
			titleId="delete-modal-title"
			onCancel={onCancel}
			onConfirm={onConfirm}
		/>
	);
}

function ForceDeleteModal({
	isDeleting,
	onCancel,
	onConfirm,
}: DeleteModalProps) {
	return (
		<ConfirmationModal
			confirmClassName="button-outline button-danger"
			confirmText="Force Delete"
			isDeleting={isDeleting}
			title="Force delete this article?"
			titleId="force-delete-modal-title"
			onCancel={onCancel}
			onConfirm={onConfirm}
		/>
	);
}

interface ConfirmationModalProps extends DeleteModalProps {
	confirmClassName: string;
	confirmText: string;
	title: string;
	titleId: string;
}

function ConfirmationModal({
	confirmClassName,
	confirmText,
	isDeleting,
	onCancel,
	onConfirm,
	title,
	titleId,
}: ConfirmationModalProps) {
	const modalRef = useRef<HTMLDivElement>(null);

	useEffect(() => {
		const previouslyFocused =
			document.activeElement instanceof HTMLElement
				? document.activeElement
				: null;
		const modal = modalRef.current;

		if (!modal) {
			return;
		}

		const initialFocus = getModalFocusableElements(modal)[0] ?? modal;
		initialFocus.focus();

		function handleKeyDown(event: KeyboardEvent) {
			if (event.key !== "Tab" || !modal) {
				return;
			}

			const focusableElements = getModalFocusableElements(modal);

			if (focusableElements.length === 0) {
				event.preventDefault();
				modal.focus();
				return;
			}

			const firstElement = focusableElements[0];
			const lastElement = focusableElements[focusableElements.length - 1];
			const activeElement = document.activeElement;

			if (!modal.contains(activeElement)) {
				event.preventDefault();
				firstElement.focus();
				return;
			}

			if (event.shiftKey && activeElement === firstElement) {
				event.preventDefault();
				lastElement.focus();
				return;
			}

			if (!event.shiftKey && activeElement === lastElement) {
				event.preventDefault();
				firstElement.focus();
			}
		}

		document.addEventListener("keydown", handleKeyDown);

		return () => {
			document.removeEventListener("keydown", handleKeyDown);

			if (previouslyFocused?.isConnected) {
				previouslyFocused.focus();
			}
		};
	}, []);

	return (
		<div className="modal-backdrop" role="presentation">
			<div
				aria-labelledby={titleId}
				aria-modal="true"
				className="delete-modal"
				ref={modalRef}
				role="dialog"
				tabIndex={-1}
			>
				<p id={titleId}>{title}</p>
				<div className="delete-modal-actions">
					<button
						className={confirmClassName}
						type="button"
						disabled={isDeleting}
						onClick={onConfirm}
					>
						{confirmText}
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

function getModalFocusableElements(modal: HTMLElement): HTMLElement[] {
	return Array.from(
		modal.querySelectorAll<HTMLElement>(
			'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
		),
	).filter((element) => element.tabIndex >= 0);
}

function focusArticleShell(shellRef: { current: HTMLElement | null }) {
	window.requestAnimationFrame(() => {
		shellRef.current?.focus();
	});
}

function errorMessage(error: unknown, fallbackMessage: string): string {
	return error instanceof Error ? error.message : fallbackMessage;
}
