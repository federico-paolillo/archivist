import MarkdownIt, { type Options } from "markdown-it";
import type Renderer from "markdown-it/lib/renderer.mjs";
import type Token from "markdown-it/lib/token.mjs";

interface MarkdownContentProps {
	markdown: string;
}

const markdownRenderer = MarkdownIt({
	html: false,
	linkify: false,
	typographer: false,
});

const defaultLinkOpen =
	markdownRenderer.renderer.rules.link_open ??
	((tokens, index, options, _env, self) =>
		self.renderToken(tokens, index, options));

markdownRenderer.renderer.rules.link_open = (
	tokens: Token[],
	index: number,
	options: Options,
	env: unknown,
	self: Renderer,
) => {
	const token = tokens[index];
	if (!token) {
		return "";
	}

	token.attrSet("target", "_blank");
	token.attrSet("rel", "noopener noreferrer");

	return defaultLinkOpen(tokens, index, options, env, self);
};

markdownRenderer.renderer.rules.image = (tokens: Token[], index: number) => {
	const token = tokens[index];
	const altText = token?.content.trim() || "image";

	return `<span class="markdown-image-placeholder">${escapeHtml(`![${altText}]`)}</span>`;
};

export function MarkdownContent({ markdown }: MarkdownContentProps) {
	return (
		<div
			className="markdown-content"
			// markdown-it runs with raw HTML disabled and its default link validator.
			dangerouslySetInnerHTML={{ __html: markdownRenderer.render(markdown) }}
		/>
	);
}

function escapeHtml(value: string) {
	return value
		.replaceAll("&", "&amp;")
		.replaceAll("<", "&lt;")
		.replaceAll(">", "&gt;")
		.replaceAll('"', "&quot;")
		.replaceAll("'", "&#39;");
}
