import {
	type ApiClient,
	ApiUnauthorizedError,
} from "@archivist/deps/api-client.ts";
import type { ArticleDetail, ArticleMetadata } from "@archivist/deps/models.ts";
import type { DetailState } from "@archivist/pages/articles/components/article-detail-state.ts";
import { errorMessage } from "@archivist/pages/articles/components/article-view-helpers.ts";
import { useCallback, useEffect, useRef, useState } from "preact/hooks";

interface UseArticleDetailProps {
	api: ApiClient;
	findArticle: (articleId: string) => ArticleMetadata | undefined;
	onAuthExpired: () => void;
	selectedArticleId?: string;
}

interface ShowDetailErrorProps {
	article?: ArticleMetadata | ArticleDetail;
	message: string;
}

export function useArticleDetail({
	api,
	findArticle,
	onAuthExpired,
	selectedArticleId,
}: UseArticleDetailProps) {
	const [detailState, setDetailState] = useState<DetailState>({
		kind: "idle",
	});
	const detailRequestRef = useRef(0);
	const loadedArticleIdRef = useRef<string | null>(null);

	const clearDetail = useCallback(() => {
		detailRequestRef.current += 1;
		loadedArticleIdRef.current = null;
		setDetailState({ kind: "idle" });
	}, []);

	const showDetailError = useCallback(
		({ article, message }: ShowDetailErrorProps) => {
			setDetailState({
				kind: "error",
				article,
				message,
			});
		},
		[],
	);

	const loadArticleDetail = useCallback(
		async (articleId: string) => {
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

				showDetailError({
					article: findArticle(articleId),
					message: errorMessage(error, "Article detail failed."),
				});
			}
		},
		[api, findArticle, onAuthExpired, showDetailError],
	);

	useEffect(() => {
		if (!selectedArticleId) {
			clearDetail();
			return;
		}

		if (loadedArticleIdRef.current !== selectedArticleId) {
			void loadArticleDetail(selectedArticleId);
		}
	}, [clearDetail, loadArticleDetail, selectedArticleId]);

	const selectedArticle =
		detailState.kind === "loaded" ? detailState.article : undefined;

	return {
		clearDetail,
		detailState,
		loadArticleDetail,
		selectedArticle,
		showDetailError,
	};
}
