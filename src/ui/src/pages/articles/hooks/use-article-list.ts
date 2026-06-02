import {
	type ApiClient,
	ApiUnauthorizedError,
} from "@archivist/deps/api-client.ts";
import type { ArticleMetadata } from "@archivist/deps/models.ts";
import { errorMessage } from "@archivist/pages/articles/components/article-view-helpers.ts";
import { useCallback, useEffect, useRef, useState } from "preact/hooks";

interface UseArticleListProps {
	api: ApiClient;
	onAuthExpired: () => void;
}

export function useArticleList({ api, onAuthExpired }: UseArticleListProps) {
	const [articles, setArticles] = useState<ArticleMetadata[]>([]);
	const articlesRef = useRef<ArticleMetadata[]>([]);
	const [listError, setListError] = useState<string | null>(null);
	const [isListLoading, setIsListLoading] = useState(true);

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

	const findArticle = useCallback((articleId: string) => {
		return articlesRef.current.find((item) => item.id === articleId);
	}, []);

	const removeArticle = useCallback((articleId: string) => {
		articlesRef.current = articlesRef.current.filter(
			(article) => article.id !== articleId,
		);
		setArticles((currentArticles) =>
			currentArticles.filter((article) => article.id !== articleId),
		);
	}, []);

	return {
		articles,
		findArticle,
		isListLoading,
		listError,
		removeArticle,
	};
}
