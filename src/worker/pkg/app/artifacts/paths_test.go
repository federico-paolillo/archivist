package artifacts_test

import (
	"path/filepath"
	"testing"

	"codeberg.org/federico-paolillo/archivist/pkg/app/artifacts"
	"github.com/stretchr/testify/require"
)

func TestArticlePathsDeriveDeterministicArtifactPaths(t *testing.T) {
	paths := artifacts.NewArticlePaths("/data")
	articleID := "01ASB2XFCZJY7WHZ2FNRTMQJCT"

	dir, err := paths.ArticleDirectory(articleID)
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/data", "articles", articleID), dir)

	snapshot, err := paths.SnapshotHTML(articleID)
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/data", "articles", articleID, "snapshot.html"), snapshot)

	content, err := paths.ContentMarkdown(articleID)
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/data", "articles", articleID, "content.md"), content)

	summary, err := paths.SummaryMarkdown(articleID)
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/data", "articles", articleID, "summary.md"), summary)
}

func TestArticlePathsRejectTraversalSegments(t *testing.T) {
	paths := artifacts.NewArticlePaths("/data")

	_, err := paths.ArticleDirectory("../escape")

	require.ErrorIs(t, err, artifacts.ErrInvalidArticleID)
}

func TestArticlePathsRejectInvalidULIDSegments(t *testing.T) {
	paths := artifacts.NewArticlePaths("/data")
	invalidIDs := []string{
		"01ASB2XFCZJY7WHZ2FNRTMQJCI",
		"81ASB2XFCZJY7WHZ2FNRTMQJCT",
		"01ASB2XFCZJY7WHZ2FNRTMQJC",
	}

	for _, articleID := range invalidIDs {
		_, err := paths.ArticleDirectory(articleID)

		require.ErrorIs(t, err, artifacts.ErrInvalidArticleID)
	}
}
