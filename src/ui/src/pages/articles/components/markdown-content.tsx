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
	token.attrSet("target", "_blank");
	token.attrSet("rel", "noopener noreferrer");

	return defaultLinkOpen(tokens, index, options, env, self);
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
