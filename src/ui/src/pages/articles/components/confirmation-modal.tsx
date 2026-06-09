import { useEffect, useRef } from "preact/hooks";

export interface ConfirmationModalProps {
	confirmClassName: string;
	confirmText: string;
	isDeleting: boolean;
	title: string;
	titleId: string;
	onCancel: () => void;
	onConfirm: () => void;
}

export function ConfirmationModal({
	confirmClassName,
	confirmText,
	isDeleting,
	onCancel,
	onConfirm,
	title,
	titleId,
}: ConfirmationModalProps) {
	const modalRef = useRef<HTMLDivElement>(null);
	const isDeletingRef = useRef(isDeleting);
	const onCancelRef = useRef(onCancel);

	useEffect(() => {
		isDeletingRef.current = isDeleting;
		onCancelRef.current = onCancel;
	});

	useEffect(() => {
		const previouslyFocused =
			document.activeElement instanceof HTMLElement
				? document.activeElement
				: null;
		const modal = modalRef.current;

		if (!modal) {
			return;
		}

		const initialFocus = getModalFocusableElements(modal)[0] ?? modal;
		initialFocus.focus();

		function handleKeyDown(event: KeyboardEvent) {
			if (event.key === "Escape") {
				if (!isDeletingRef.current) {
					event.preventDefault();
					onCancelRef.current();
				}
				return;
			}

			if (event.key !== "Tab" || !modal) {
				return;
			}

			const focusableElements = getModalFocusableElements(modal);

			if (focusableElements.length === 0) {
				event.preventDefault();
				modal.focus();
				return;
			}

			const firstElement = focusableElements[0];
			const lastElement = focusableElements[focusableElements.length - 1];
			if (!firstElement || !lastElement) {
				return;
			}
			const activeElement = document.activeElement;

			if (!modal.contains(activeElement)) {
				event.preventDefault();
				firstElement.focus();
				return;
			}

			if (event.shiftKey && activeElement === firstElement) {
				event.preventDefault();
				lastElement.focus();
				return;
			}

			if (!event.shiftKey && activeElement === lastElement) {
				event.preventDefault();
				firstElement.focus();
			}
		}

		document.addEventListener("keydown", handleKeyDown);

		return () => {
			document.removeEventListener("keydown", handleKeyDown);

			if (previouslyFocused?.isConnected) {
				previouslyFocused.focus();
			}
		};
	}, []);

	return (
		<div className="modal-backdrop" role="presentation">
			<div
				aria-labelledby={titleId}
				aria-modal="true"
				className="delete-modal"
				ref={modalRef}
				role="dialog"
				tabIndex={-1}
			>
				<p id={titleId}>{title}</p>
				<div className="delete-modal-actions">
					<button
						className={confirmClassName}
						type="button"
						disabled={isDeleting}
						onClick={onConfirm}
					>
						{confirmText}
					</button>
					<button
						className="button-outline"
						type="button"
						disabled={isDeleting}
						onClick={onCancel}
					>
						Nevermind
					</button>
				</div>
			</div>
		</div>
	);
}

function getModalFocusableElements(modal: HTMLElement): HTMLElement[] {
	return Array.from(
		modal.querySelectorAll<HTMLElement>(
			'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
		),
	).filter((element) => element.tabIndex >= 0);
}
