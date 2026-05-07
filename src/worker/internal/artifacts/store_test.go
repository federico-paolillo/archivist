package artifacts

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const testArticleID = "01ARZ3NDEKTSV4RRFFQ69G5FAV"

func TestSnapshotPathIsDeterministic(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	store, err := NewStore(dataDir)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})

	path, err := store.SnapshotPath(testArticleID)

	require.NoError(t, err)
	require.Equal(t, filepath.Join(dataDir, "articles", testArticleID, "snapshot.html"), path)
}

func TestWriteSnapshotAtomicallyPromotesFinalFile(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	store, err := NewStore(dataDir)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, store.Close())
	})

	err = store.WriteSnapshot(testArticleID, []byte("<html>ok</html>"))

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

	err = store.WriteSnapshot(testArticleID, []byte("<html>ok</html>"))

	require.Error(t, err)
	tempFiles, globErr := filepath.Glob(filepath.Join(articleDir, ".snapshot.html.*.tmp"))
	require.NoError(t, globErr)
	require.Empty(t, tempFiles)

	info, statErr := os.Stat(filepath.Join(articleDir, "snapshot.html"))
	require.NoError(t, statErr)
	require.True(t, info.IsDir())
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
		_, pathErr := store.SnapshotPath(articleID)
		require.ErrorIs(t, pathErr, ErrInvalidArticleID)

		writeErr := store.WriteSnapshot(articleID, []byte("<html>no</html>"))
		require.ErrorIs(t, writeErr, ErrInvalidArticleID)
	}
}

func TestNewStoreRejectsEmptyDataDir(t *testing.T) {
	t.Parallel()

	store, err := NewStore("")

	require.Nil(t, store)
	require.True(t, errors.Is(err, ErrEmptyDataDir))
}
