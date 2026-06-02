import type { ArticleDetail, ArticleMetadata } from "@archivist/deps/models.ts";

export type DetailState =
	| { kind: "idle" }
	| { kind: "loading" }
	| {
			article?: ArticleMetadata | ArticleDetail;
			kind: "error";
			message: string;
	  }
	| { article: ArticleDetail; kind: "loaded" };
