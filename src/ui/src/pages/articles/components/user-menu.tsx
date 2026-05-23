interface UserMenuProps {
	isOpen: boolean;
	onLogout: () => void;
	onToggle: () => void;
}

export function UserMenu({ isOpen, onLogout, onToggle }: UserMenuProps) {
	return (
		<div className="user-menu">
			<button
				aria-expanded={isOpen}
				aria-label="User menu"
				className="icon-button"
				type="button"
				onClick={onToggle}
			>
				<span className="user-glyph" aria-hidden="true" />
			</button>
			{isOpen ? (
				<div className="menu" role="menu">
					<button className="menu-item" type="button" onClick={onLogout}>
						Logout
					</button>
				</div>
			) : null}
		</div>
	);
}
