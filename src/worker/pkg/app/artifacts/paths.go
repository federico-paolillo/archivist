package artifacts

import (
	"errors"
	"path/filepath"
	"unicode"
)

var errInvalidArticleID = errors.New("article id must be a ULID path segment")

type ArticlePaths struct {
	dataDir string
}

func NewArticlePaths(dataDir string) *ArticlePaths {
	return &ArticlePaths{dataDir: dataDir}
}

func (p *ArticlePaths) ArticleDirectory(articleID string) (string, error) {
	err := validateArticleID(articleID)
	if err != nil {
		return "", err
	}

	return filepath.Join(p.dataDir, "articles", articleID), nil
}

func (p *ArticlePaths) SnapshotHTML(articleID string) (string, error) {
	return p.artifactPath(articleID, "snapshot.html")
}

func (p *ArticlePaths) ContentMarkdown(articleID string) (string, error) {
	return p.artifactPath(articleID, "content.md")
}

func (p *ArticlePaths) SummaryMarkdown(articleID string) (string, error) {
	return p.artifactPath(articleID, "summary.md")
}

func (p *ArticlePaths) SummaryJSON(articleID string) (string, error) {
	return p.artifactPath(articleID, "summary.json")
}

func (p *ArticlePaths) MetadataJSON(articleID string) (string, error) {
	return p.artifactPath(articleID, "metadata.json")
}

func (p *ArticlePaths) artifactPath(articleID string, filename string) (string, error) {
	dir, err := p.ArticleDirectory(articleID)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, filename), nil
}

func validateArticleID(articleID string) error {
	if len(articleID) != 26 {
		return errInvalidArticleID
	}

	for _, c := range articleID {
		if !unicode.IsDigit(c) && !unicode.IsUpper(c) {
			return errInvalidArticleID
		}
	}

	return nil
}
