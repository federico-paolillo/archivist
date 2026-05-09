package artifacts

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	tempSnapshotPattern = ".snapshot.html.*.tmp"
	articleDirPerm      = 0o700
)

var ErrEmptyDataDir = errors.New("artifacts: data dir is empty")

// Store provides traversal-resistant, operation-first access to article artifacts under DATA_DIR.
type Store struct {
	dataDir string
	root    *os.Root
}

func NewStore(dataDir string) (*Store, error) {
	if dataDir == "" {
		return nil, ErrEmptyDataDir
	}

	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, fmt.Errorf("artifacts: resolve data dir: %w", err)
	}

	err = os.MkdirAll(absDataDir, articleDirPerm)
	if err != nil {
		return nil, fmt.Errorf("artifacts: create data dir: %w", err)
	}

	root, err := os.OpenRoot(absDataDir)
	if err != nil {
		return nil, fmt.Errorf("artifacts: open data dir root: %w", err)
	}

	return &Store{
		dataDir: absDataDir,
		root:    root,
	}, nil
}

func (s *Store) Close() error {
	if s == nil || s.root == nil {
		return nil
	}

	err := s.root.Close()
	if err != nil {
		return fmt.Errorf("artifacts: close data dir root: %w", err)
	}

	return nil
}

// OpenSnapshot returns a reader over the persisted snapshot HTML for articleID.
// Caller must Close the returned ReadCloser.
// Returns a fs.ErrNotExist-wrapping error if the snapshot has not been written yet.
func (s *Store) OpenSnapshot(articleID string) (io.ReadCloser, error) {
	return s.openArtifact(articleID, SnapshotHTMLFilename)
}

// WriteSnapshot streams html to the snapshot file for articleID, atomically.
// Replaces any existing snapshot.
func (s *Store) WriteSnapshot(articleID string, html io.Reader) error {
	return s.writeArtifact(articleID, SnapshotHTMLFilename, tempSnapshotPattern, html)
}

func (s *Store) openArtifact(articleID, filename string) (io.ReadCloser, error) {
	err := ValidateArticleID(articleID)
	if err != nil {
		return nil, fmt.Errorf("artifacts: validate article id: %w", err)
	}

	relPath := filepath.Join(ArticlesDirectoryName, articleID, filename)

	file, err := s.root.Open(relPath)
	if err != nil {
		return nil, fmt.Errorf("artifacts: open artifact: %w", err)
	}

	return file, nil
}

func (s *Store) writeArtifact(articleID, filename, tempPattern string, src io.Reader) error {
	relDir, err := articleRelDir(articleID)
	if err != nil {
		return err
	}

	err = s.root.MkdirAll(relDir, articleDirPerm)
	if err != nil {
		return fmt.Errorf("artifacts: create article dir: %w", err)
	}

	absDir := filepath.Join(s.dataDir, relDir)

	file, err := os.CreateTemp(absDir, tempPattern)
	if err != nil {
		return fmt.Errorf("artifacts: create temp file: %w", err)
	}

	tempRelPath := filepath.Join(relDir, filepath.Base(file.Name()))
	finalRelPath := filepath.Join(relDir, filename)

	committed := false
	defer func() {
		if !committed {
			_ = s.root.Remove(tempRelPath)
		}
	}()

	_, err = io.Copy(file, src)
	if err != nil {
		_ = file.Close()

		return fmt.Errorf("artifacts: write temp artifact: %w", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("artifacts: close temp artifact: %w", err)
	}

	err = s.root.Rename(tempRelPath, finalRelPath)
	if err != nil {
		return fmt.Errorf("artifacts: promote artifact: %w", err)
	}

	committed = true

	return nil
}

func articleRelDir(articleID string) (string, error) {
	err := ValidateArticleID(articleID)
	if err != nil {
		return "", fmt.Errorf("artifacts: validate article dir id: %w", err)
	}

	return filepath.Join(ArticlesDirectoryName, articleID), nil
}
