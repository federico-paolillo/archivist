import {
	type ApiClient,
	ApiUnauthorizedError,
} from "@archivist/deps/api-client.ts";
import type { ArticleDetail, ArticleMetadata } from "@archivist/deps/models.ts";
import { errorMessage } from "@archivist/pages/articles/components/article-view-helpers.ts";
import { useCallback, useEffect, useState } from "preact/hooks";

interface UseArticleDeleteProps {
	actionArticle?: ArticleMetadata | ArticleDetail | undefined;
	api: ApiClient;
	clearDetail: () => void;
	focusShell: () => void;
	navigateToArticles: () => void;
	onAuthExpired: () => void;
	removeArticle: (articleId: string) => void;
	selectedArticleId?: string | undefined;
	showDetailError: (props: {
		article?: ArticleMetadata | ArticleDetail | undefined;
		message: string;
	}) => void;
}

export function useArticleDelete({
	actionArticle,
	api,
	clearDetail,
	focusShell,
	navigateToArticles,
	onAuthExpired,
	removeArticle,
	selectedArticleId,
	showDetailError,
}: UseArticleDeleteProps) {
	const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
	const [isDeleting, setIsDeleting] = useState(false);
	const [isForceDeleteModalOpen, setIsForceDeleteModalOpen] = useState(false);
	const [isForceDeleting, setIsForceDeleting] = useState(false);

	useEffect(() => {
		if (!selectedArticleId) {
			setIsDeleteModalOpen(false);
			setIsForceDeleteModalOpen(false);
		}
	}, [selectedArticleId]);

	const finishSuccessfulDelete = useCallback(
		(articleId: string) => {
			removeArticle(articleId);
			clearDetail();
			navigateToArticles();
			focusShell();
		},
		[clearDetail, focusShell, navigateToArticles, removeArticle],
	);

	const confirmDelete = useCallback(async () => {
		if (!selectedArticleId) {
			return;
		}

		const articleId = selectedArticleId;
		setIsDeleting(true);

		try {
			await api.deleteArticle(articleId);
			setIsDeleteModalOpen(false);
			finishSuccessfulDelete(articleId);
		} catch (error) {
			if (error instanceof ApiUnauthorizedError) {
				onAuthExpired();
				return;
			}

			setIsDeleteModalOpen(false);
			showDetailError({
				article: actionArticle,
				message: errorMessage(error, "Delete failed."),
			});
		} finally {
			setIsDeleting(false);
		}
	}, [
		actionArticle,
		api,
		finishSuccessfulDelete,
		onAuthExpired,
		selectedArticleId,
		showDetailError,
	]);

	const confirmForceDelete = useCallback(async () => {
		if (!selectedArticleId) {
			return;
		}

		const articleId = selectedArticleId;
		setIsForceDeleting(true);

		try {
			await api.forceDeleteArticle(articleId);
			setIsForceDeleteModalOpen(false);
			finishSuccessfulDelete(articleId);
		} catch (error) {
			if (error instanceof ApiUnauthorizedError) {
				onAuthExpired();
				return;
			}

			setIsForceDeleteModalOpen(false);
			showDetailError({
				article: actionArticle,
				message: errorMessage(error, "Force delete failed."),
			});
		} finally {
			setIsForceDeleting(false);
		}
	}, [
		actionArticle,
		api,
		finishSuccessfulDelete,
		onAuthExpired,
		selectedArticleId,
		showDetailError,
	]);

	return {
		closeDeleteModal: () => {
			setIsDeleteModalOpen(false);
		},
		closeForceDeleteModal: () => {
			setIsForceDeleteModalOpen(false);
		},
		confirmDelete,
		confirmForceDelete,
		isDeleteModalOpen,
		isDeleting,
		isForceDeleteModalOpen,
		isForceDeleting,
		openDeleteModal: () => {
			setIsDeleteModalOpen(true);
		},
		openForceDeleteModal: () => {
			setIsForceDeleteModalOpen(true);
		},
	};
}
