package artifacts

import (
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
