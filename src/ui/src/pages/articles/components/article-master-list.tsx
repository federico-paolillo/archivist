import type { ArticleMetadata } from "@archivist/deps/models.ts";
import {
	articleSourceUrl,
	articleTitle,
} from "@archivist/pages/articles/components/article-view-helpers.ts";

interface ArticleMasterListProps {
	articles: ArticleMetadata[];
	error: string | null;
	isLoading: boolean;
	onSelect: (articleId: string) => void;
	selectedArticleId?: string;
}

export function ArticleMasterList({
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
			{articles.map((article) => {
				const isSelected = article.id === selectedArticleId;

				return (
					<button
						aria-pressed={isSelected}
						className={
							isSelected ? "article-row article-row-selected" : "article-row"
						}
						key={article.id}
						type="button"
						onClick={() => {
							onSelect(article.id);
						}}
					>
						<span className="article-row-id">ID: {article.id}</span>
						<span className="article-row-title">{articleTitle(article)}</span>
						<span className="article-row-url">
							{articleSourceUrl(article) || "No source URL"}
						</span>
						<span className="article-row-status">{article.status}</span>
					</button>
				);
			})}
		</aside>
	);
}
