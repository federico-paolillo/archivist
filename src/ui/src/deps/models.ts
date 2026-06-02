export interface ArticleMetadata {
	id: string;
	title: string | null;
	originalUrl: string | null;
	canonicalUrl: string | null;
	status: string;
	errorMessage: string | null;
	createdAt: string;
}

export interface ArticleDetail extends ArticleMetadata {
	canForceDelete: boolean;
	summaryMarkdown: string | null;
	contentMarkdown: string | null;
}

export interface ArticleListResponse {
	items: ArticleMetadata[];
	nextCursor: string | null;
	previousCursor: string | null;
}
