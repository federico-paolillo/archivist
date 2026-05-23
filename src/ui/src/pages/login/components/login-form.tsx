interface LoginFormProps {
	isSubmitting: boolean;
	onPasswordInput: (password: string) => void;
	onSubmit: (event: SubmitEvent) => void;
	password: string;
}

export function LoginForm({
	isSubmitting,
	onPasswordInput,
	onSubmit,
	password,
}: LoginFormProps) {
	return (
		<form className="login-panel" onSubmit={onSubmit}>
			<h1 className="login-title">ARCHIVIST</h1>
			<label className="field">
				<span className="field-label">PASSWORD</span>
				<input
					type="password"
					aria-label="Password"
					className="password-input"
					value={password}
					onInput={(event) => {
						onPasswordInput(event.currentTarget.value);
					}}
					autoComplete="current-password"
					autoCapitalize="off"
					spellcheck={false}
					required
				/>
			</label>
			<button className="button-primary" type="submit" disabled={isSubmitting}>
				IDENTIFY
			</button>
		</form>
	);
}
