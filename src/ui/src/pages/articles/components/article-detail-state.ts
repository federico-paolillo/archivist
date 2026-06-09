import type { ArticleDetail, ArticleMetadata } from "@archivist/deps/models.ts";

export type DetailState =
	| { kind: "idle" }
	| { kind: "loading" }
	| {
			article?: ArticleMetadata | ArticleDetail | undefined;
			kind: "error";
			message: string;
	  }
	| { article: ArticleDetail; kind: "loaded" };
