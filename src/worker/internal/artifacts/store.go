package artifacts

import (
	"io"
	"os"
	"path/filepath"
)

const (
	tempSnapshotPattern = ".snapshot.html.*.tmp"
	tempContentPattern  = ".content.md.*.tmp"
	articleDirPerm      = 0o700
)

// Store provides traversal-resistant, operation-first access to article artifacts under DATA_DIR.
type Store struct {
	dataDir string
	root    *os.Root
}

func NewStore(dataDir string) (*Store, error) {
	if dataDir == "" {
		return nil, storeFailure("validate data dir", ErrEmptyDataDir)
	}

	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, storeFailure("resolve data dir", err, withPath(dataDir))
	}

	err = os.MkdirAll(absDataDir, articleDirPerm)
	if err != nil {
		return nil, storeFailure("create data dir", err, withPath(absDataDir))
	}

	root, err := os.OpenRoot(absDataDir)
	if err != nil {
		return nil, storeFailure("open data dir root", err, withPath(absDataDir))
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
		return storeFailure("close data dir root", err, withPath(s.dataDir))
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

// OpenMarkdown returns a reader over the persisted Markdown content for articleID.
// Caller must Close the returned ReadCloser.
// Returns a fs.ErrNotExist-wrapping error if the content has not been written yet.
func (s *Store) OpenMarkdown(articleID string) (io.ReadCloser, error) {
	return s.openArtifact(articleID, ContentMDFilename)
}

// WriteMarkdown streams markdown to the content.md file for articleID, atomically.
// Replaces any existing content.md.
func (s *Store) WriteMarkdown(articleID string, markdown io.Reader) error {
	return s.writeArtifact(articleID, ContentMDFilename, tempContentPattern, markdown)
}

func (s *Store) openArtifact(articleID, filename string) (io.ReadCloser, error) {
	err := ValidateArticleID(articleID)
	if err != nil {
		return nil, storeFailure("validate article id", err, withArticleID(articleID), withFilename(filename))
	}

	relPath := filepath.Join(ArticlesDirectoryName, articleID, filename)

	file, err := s.root.Open(relPath)
	if err != nil {
		return nil, storeFailure(
			"open artifact",
			err,
			withArticleID(articleID),
			withFilename(filename),
			withPath(relPath),
		)
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
		return storeFailure("create article dir", err, withArticleID(articleID), withPath(relDir))
	}

	absDir := filepath.Join(s.dataDir, relDir)

	file, err := os.CreateTemp(absDir, tempPattern)
	if err != nil {
		return storeFailure("create temp file", err, withArticleID(articleID), withFilename(filename), withPath(absDir))
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

		return storeFailure("write temp artifact", err, withArticleID(articleID), withFilename(filename), withPath(tempRelPath))
	}

	err = file.Close()
	if err != nil {
		return storeFailure("close temp artifact", err, withArticleID(articleID), withFilename(filename), withPath(tempRelPath))
	}

	err = s.root.Rename(tempRelPath, finalRelPath)
	if err != nil {
		return storeFailure("promote artifact", err, withArticleID(articleID), withFilename(filename), withPath(finalRelPath))
	}

	committed = true

	return nil
}

func articleRelDir(articleID string) (string, error) {
	err := ValidateArticleID(articleID)
	if err != nil {
		return "", storeFailure("validate article dir id", err, withArticleID(articleID))
	}

	return filepath.Join(ArticlesDirectoryName, articleID), nil
}
