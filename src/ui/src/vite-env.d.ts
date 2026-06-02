interface ViteTypeOptions {
	strictImportMetaEnv: unknown;
}

interface ImportMetaEnv {
	readonly VITE_VERSION_LABEL: string;
	readonly VITE_API_BASE_PATH: string;
}

interface ImportMeta {
	readonly env: ImportMetaEnv;
}
