import { ConfirmationModal } from "@archivist/pages/articles/components/confirmation-modal.tsx";

interface DeleteModalProps {
	isDeleting: boolean;
	onCancel: () => void;
	onConfirm: () => void;
}

export function DeleteModal({
	isDeleting,
	onCancel,
	onConfirm,
}: DeleteModalProps) {
	return (
		<ConfirmationModal
			confirmClassName="button-outline"
			confirmText="Yes"
			isDeleting={isDeleting}
			title="Are you sure?"
			titleId="delete-modal-title"
			onCancel={onCancel}
			onConfirm={onConfirm}
		/>
	);
}

export function ForceDeleteModal({
	isDeleting,
	onCancel,
	onConfirm,
}: DeleteModalProps) {
	return (
		<ConfirmationModal
			confirmClassName="button-outline button-danger"
			confirmText="Force Delete"
			isDeleting={isDeleting}
			title="Force delete this article?"
			titleId="force-delete-modal-title"
			onCancel={onCancel}
			onConfirm={onConfirm}
		/>
	);
}
