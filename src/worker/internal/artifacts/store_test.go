package artifacts

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const testArticleID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"

func TestSnapshotRoundTrip(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	store, err := NewStore(dataDir)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})

	err = store.WriteSnapshot(testArticleID, strings.NewReader("<html>ok</html>"))
	require.NoError(t, err)

	rc, err := store.OpenSnapshot(testArticleID)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, rc.Close())
	})

	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, "<html>ok</html>", string(got))
}

func TestWriteSnapshotAtomicallyPromotesFinalFile(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	store, err := NewStore(dataDir)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})

	err = store.WriteSnapshot(testArticleID, strings.NewReader("<html>ok</html>"))

	require.NoError(t, err)

	snapshotPath := filepath.Join(dataDir, "articles", testArticleID, "snapshot.html")
	snapshot, err := os.ReadFile(snapshotPath)
	require.NoError(t, err)
	require.Equal(t, "<html>ok</html>", string(snapshot))

	tempFiles, err := filepath.Glob(filepath.Join(dataDir, "articles", testArticleID, ".snapshot.html.*.tmp"))
	require.NoError(t, err)
	require.Empty(t, tempFiles)
}

func TestWriteSnapshotCleansTempFileWhenPromotionFails(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	store, err := NewStore(dataDir)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})

	articleDir := filepath.Join(dataDir, "articles", testArticleID)
	require.NoError(t, os.MkdirAll(articleDir, 0o700))
	require.NoError(t, os.Mkdir(filepath.Join(articleDir, "snapshot.html"), 0o700))

	err = store.WriteSnapshot(testArticleID, strings.NewReader("<html>ok</html>"))

	require.Error(t, err)
	tempFiles, globErr := filepath.Glob(filepath.Join(articleDir, ".snapshot.html.*.tmp"))
	require.NoError(t, globErr)
	require.Empty(t, tempFiles)

	info, statErr := os.Stat(filepath.Join(articleDir, "snapshot.html"))
	require.NoError(t, statErr)
	require.True(t, info.IsDir())
}

func TestWriteSnapshotCleansTempFileWhenSrcFails(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	store, err := NewStore(dataDir)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})

	err = store.WriteSnapshot(testArticleID, &failingReader{})

	require.Error(t, err)

	articleDir := filepath.Join(dataDir, "articles", testArticleID)
	tempFiles, globErr := filepath.Glob(filepath.Join(articleDir, ".snapshot.html.*.tmp"))
	require.NoError(t, globErr)
	require.Empty(t, tempFiles)
}

func TestOpenSnapshotReturnsNotExistWhenAbsent(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	store, err := NewStore(dataDir)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})

	_, err = store.OpenSnapshot(testArticleID)

	require.Error(t, err)
	require.True(t, errors.Is(err, fs.ErrNotExist))
}

func TestArtifactAccessRejectsTraversal(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	store, err := NewStore(dataDir)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})

	traversalIDs := []string{
		"../01ARZ3NDEKTSV4RRFFQ69G5FA",
		"01ARZ3NDEKTSV4RRFFQ69G5FA/",
		"01ARZ3NDEKTSV4RRFFQ69G5Fa",
		"01ARZ3NDEKTSV4RRFFQ69G5FAI",
		"81ARZ3NDEKTSV4RRFFQ69G5FA",
	}

	for _, articleID := range traversalIDs {
		_, openSnapErr := store.OpenSnapshot(articleID)
		require.ErrorIs(t, openSnapErr, ErrInvalidArticleID)

		writeSnapErr := store.WriteSnapshot(articleID, strings.NewReader("<html>no</html>"))
		require.ErrorIs(t, writeSnapErr, ErrInvalidArticleID)

		_, openMDErr := store.OpenMarkdown(articleID)
		require.ErrorIs(t, openMDErr, ErrInvalidArticleID)

		writeMDErr := store.WriteMarkdown(articleID, strings.NewReader("# no"))
		require.ErrorIs(t, writeMDErr, ErrInvalidArticleID)
	}
}

func TestNewStoreRejectsEmptyDataDir(t *testing.T) {
	t.Parallel()

	store, err := NewStore("")

	require.Nil(t, store)
	require.True(t, errors.Is(err, ErrEmptyDataDir))
}

func TestMarkdownRoundTrip(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	store, err := NewStore(dataDir)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})

	err = store.WriteMarkdown(testArticleID, strings.NewReader("# Hello\n\nworld"))
	require.NoError(t, err)

	rc, err := store.OpenMarkdown(testArticleID)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, rc.Close())
	})

	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, "# Hello\n\nworld", string(got))
}

func TestMarkdownPathIsDeterministic(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	store, err := NewStore(dataDir)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})

	err = store.WriteMarkdown(testArticleID, strings.NewReader("# Hello"))
	require.NoError(t, err)

	expectedPath := filepath.Join(dataDir, "articles", testArticleID, "content.md")
	content, err := os.ReadFile(expectedPath)
	require.NoError(t, err)
	require.Equal(t, "# Hello", string(content))
}

func TestWriteMarkdownAtomicallyPromotesFinalFile(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	store, err := NewStore(dataDir)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})

	err = store.WriteMarkdown(testArticleID, strings.NewReader("# Article"))

	require.NoError(t, err)

	contentPath := filepath.Join(dataDir, "articles", testArticleID, "content.md")
	content, err := os.ReadFile(contentPath)
	require.NoError(t, err)
	require.Equal(t, "# Article", string(content))

	tempFiles, err := filepath.Glob(filepath.Join(dataDir, "articles", testArticleID, ".content.md.*.tmp"))
	require.NoError(t, err)
	require.Empty(t, tempFiles)
}

func TestWriteMarkdownCleansTempFileWhenPromotionFails(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	store, err := NewStore(dataDir)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})

	articleDir := filepath.Join(dataDir, "articles", testArticleID)
	require.NoError(t, os.MkdirAll(articleDir, 0o700))
	require.NoError(t, os.Mkdir(filepath.Join(articleDir, "content.md"), 0o700))

	err = store.WriteMarkdown(testArticleID, strings.NewReader("# Article"))

	require.Error(t, err)
	tempFiles, globErr := filepath.Glob(filepath.Join(articleDir, ".content.md.*.tmp"))
	require.NoError(t, globErr)
	require.Empty(t, tempFiles)

	info, statErr := os.Stat(filepath.Join(articleDir, "content.md"))
	require.NoError(t, statErr)
	require.True(t, info.IsDir())
}

func TestWriteMarkdownCleansTempFileWhenSrcFails(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	store, err := NewStore(dataDir)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})

	err = store.WriteMarkdown(testArticleID, &failingReader{})

	require.Error(t, err)

	articleDir := filepath.Join(dataDir, "articles", testArticleID)
	tempFiles, globErr := filepath.Glob(filepath.Join(articleDir, ".content.md.*.tmp"))
	require.NoError(t, globErr)
	require.Empty(t, tempFiles)
}

func TestOpenMarkdownReturnsNotExistWhenAbsent(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	store, err := NewStore(dataDir)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})

	_, err = store.OpenMarkdown(testArticleID)

	require.Error(t, err)
	require.True(t, errors.Is(err, fs.ErrNotExist))
}

// failingReader always returns an error on Read.
type failingReader struct{}

func (f *failingReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read error")
}
