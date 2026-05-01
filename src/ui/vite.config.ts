import { defineConfig } from "vitest/config";
import preact from "@preact/preset-vite";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
	plugins: [preact(), tailwindcss()],
	test: {
		environment: "jsdom",
		include: ["src/**/*.test.{ts,tsx}"],
	},
	resolve: {
		tsconfigPaths: true
	}
});
