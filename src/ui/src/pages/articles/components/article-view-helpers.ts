import type { ArticleDetail, ArticleMetadata } from "@archivist/deps/models.ts";

export function articleTitle(article: ArticleMetadata | ArticleDetail): string {
	return article.title?.trim() || "Untitled article";
}

export function articleSourceUrl(
	article: ArticleMetadata | ArticleDetail,
): string | null {
	return article.canonicalUrl || article.originalUrl;
}

export function isArticleDetail(
	article: ArticleMetadata | ArticleDetail,
): article is ArticleDetail {
	return "canForceDelete" in article;
}

export function focusArticleShell(shellRef: { current: HTMLElement | null }) {
	window.requestAnimationFrame(() => {
		shellRef.current?.focus();
	});
}

export function errorMessage(error: unknown, fallbackMessage: string): string {
	return error instanceof Error ? error.message : fallbackMessage;
}
