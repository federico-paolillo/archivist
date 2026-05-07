package artifacts

import (
	"errors"
	"path/filepath"
	"strings"
)

const (
	ArticlesDirectoryName = "articles"
	SnapshotHTMLFilename  = "snapshot.html"
	ContentMDFilename     = "content.md"
	SummaryMDFilename     = "summary.md"
	SummaryJSONFilename   = "summary.json"
	MetadataJSONFilename  = "metadata.json"
)

const (
	ulidLength   = 26
	ulidAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
)

var ErrInvalidArticleID = errors.New("article id must be a ULID path segment")

type ArticlePaths struct {
	dataDir string
}

func NewArticlePaths(dataDir string) *ArticlePaths {
	return &ArticlePaths{dataDir: dataDir}
}

func (p *ArticlePaths) ArticleDirectory(articleID string) (string, error) {
	err := ValidateArticleID(articleID)
	if err != nil {
		return "", err
	}

	return filepath.Join(p.dataDir, ArticlesDirectoryName, articleID), nil
}

func (p *ArticlePaths) SnapshotHTML(articleID string) (string, error) {
	return p.artifactPath(articleID, SnapshotHTMLFilename)
}

func (p *ArticlePaths) ContentMarkdown(articleID string) (string, error) {
	return p.artifactPath(articleID, ContentMDFilename)
}

func (p *ArticlePaths) SummaryMarkdown(articleID string) (string, error) {
	return p.artifactPath(articleID, SummaryMDFilename)
}

func (p *ArticlePaths) SummaryJSON(articleID string) (string, error) {
	return p.artifactPath(articleID, SummaryJSONFilename)
}

func (p *ArticlePaths) MetadataJSON(articleID string) (string, error) {
	return p.artifactPath(articleID, MetadataJSONFilename)
}

func (p *ArticlePaths) artifactPath(articleID string, filename string) (string, error) {
	dir, err := p.ArticleDirectory(articleID)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, filename), nil
}

func ValidateArticleID(articleID string) error {
	if len(articleID) != ulidLength {
		return ErrInvalidArticleID
	}

	if articleID[0] > '7' {
		return ErrInvalidArticleID
	}

	for _, char := range articleID {
		if !strings.ContainsRune(ulidAlphabet, char) {
			return ErrInvalidArticleID
		}
	}

	return nil
}
