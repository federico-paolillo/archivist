import type { ArticleDetail } from "@archivist/deps/models.ts";
import { articleTitle } from "@archivist/pages/articles/components/article-view-helpers.ts";
import { MarkdownContent } from "@archivist/pages/articles/components/markdown-content.tsx";

interface LoadedArticleDetailProps {
	article: ArticleDetail;
}

export function LoadedArticleDetail({ article }: LoadedArticleDetailProps) {
	if (article.status === "ready") {
		return (
			<article className="ready-article">
				<header className="detail-hero">
					<h1>{articleTitle(article)}</h1>
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
