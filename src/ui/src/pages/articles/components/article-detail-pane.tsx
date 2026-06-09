import type { ArticleDetail, ArticleMetadata } from "@archivist/deps/models.ts";
import { ArticleActions } from "@archivist/pages/articles/components/article-actions.tsx";
import type { DetailState } from "@archivist/pages/articles/components/article-detail-state.ts";
import { LoadedArticleDetail } from "@archivist/pages/articles/components/loaded-article-detail.tsx";

interface ArticleDetailPaneProps {
	article?: ArticleMetadata | ArticleDetail | undefined;
	detailState: DetailState;
	onDelete: () => void;
	onForceDelete: () => void;
}

export function ArticleDetailPane({
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
