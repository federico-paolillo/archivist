import { defineConfig } from "vitest/config";
import preact from "@preact/preset-vite";
import tailwindcss from "@tailwindcss/vite";
import { fileURLToPath } from "node:url";

const preactCompat = fileURLToPath(
	new URL("./node_modules/preact/compat", import.meta.url),
);

const preactCompatClient = fileURLToPath(
	new URL("./node_modules/preact/compat/client.js", import.meta.url),
);

export default defineConfig({
	plugins: [preact(), tailwindcss()],
	publicDir: 'public/',
	test: {
		setupFiles: ["./src/test-setup.ts"],
		environment: "jsdom",
		environmentOptions: {
			jsdom: {
				url: "http://localhost/",
			},
		},
		include: ["src/**/*.test.{ts,tsx}"],
		server: {
			deps: {
				inline: [
					"@testing-library/react",
					"react",
					"react-dom",
					"react-dom/client",
				],
			},
		},
	},
	resolve: {
		alias: [
			{ find: "react-dom/client", replacement: preactCompatClient },
			{ find: "react-dom", replacement: preactCompat },
			{ find: "react", replacement: preactCompat },
		],
		tsconfigPaths: true,
	},
});
