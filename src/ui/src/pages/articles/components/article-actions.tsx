import type { ArticleDetail, ArticleMetadata } from "@archivist/deps/models.ts";
import {
	articleSourceUrl,
	isArticleDetail,
} from "@archivist/pages/articles/components/article-view-helpers.ts";

interface ArticleActionsProps {
	article: ArticleMetadata | ArticleDetail;
	onDelete: () => void;
	onForceDelete: () => void;
}

export function ArticleActions({
	article,
	onDelete,
	onForceDelete,
}: ArticleActionsProps) {
	const originalUrl = articleSourceUrl(article);
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
