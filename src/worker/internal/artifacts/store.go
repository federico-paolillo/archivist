package artifacts

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	appartifacts "codeberg.org/federico-paolillo/archivist/pkg/app/artifacts"
)

const (
	tempSnapshotPrefix  = ".snapshot.html."
	tempSnapshotSuffix  = ".tmp"
	tempNameRandomBytes = 8
	maxTempNameAttempts = 16
	articleDirPerm      = 0o700
	snapshotFilePerm    = 0o600
)

var (
	ErrEmptyDataDir     = errors.New("artifacts: data dir is empty")
	ErrInvalidArticleID = appartifacts.ErrInvalidArticleID
)

// Store provides traversal-resistant access to article artifacts under DATA_DIR.
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

func (s *Store) SnapshotPath(articleID string) (string, error) {
	err := appartifacts.ValidateArticleID(articleID)
	if err != nil {
		return "", fmt.Errorf("artifacts: validate snapshot article id: %w", err)
	}

	return filepath.Join(
		s.dataDir,
		appartifacts.ArticlesDirectoryName,
		articleID,
		appartifacts.SnapshotHTMLFilename,
	), nil
}

func (s *Store) WriteSnapshot(articleID string, html []byte) error {
	articleDir, err := articleDir(articleID)
	if err != nil {
		return err
	}

	err = s.root.MkdirAll(articleDir, articleDirPerm)
	if err != nil {
		return fmt.Errorf("artifacts: create article dir: %w", err)
	}

	file, tempPath, err := s.createTempSnapshot(articleDir)
	if err != nil {
		return err
	}

	finalPath := filepath.Join(articleDir, appartifacts.SnapshotHTMLFilename)

	committed := false
	defer func() {
		if !committed {
			_ = s.root.Remove(tempPath)
		}
	}()

	_, err = file.Write(html)
	if err != nil {
		_ = file.Close()

		return fmt.Errorf("artifacts: write temp snapshot: %w", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("artifacts: close temp snapshot: %w", err)
	}

	err = s.root.Rename(tempPath, finalPath)
	if err != nil {
		return fmt.Errorf("artifacts: promote snapshot: %w", err)
	}

	committed = true

	return nil
}

func (s *Store) createTempSnapshot(articleDir string) (*os.File, string, error) {
	for range maxTempNameAttempts {
		tempPath, err := tempSnapshotPath(articleDir)
		if err != nil {
			return nil, "", err
		}

		file, err := s.root.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, snapshotFilePerm)
		if err == nil {
			return file, tempPath, nil
		}

		if !errors.Is(err, fs.ErrExist) {
			return nil, "", fmt.Errorf("artifacts: create temp snapshot: %w", err)
		}
	}

	return nil, "", errors.New("artifacts: create temp snapshot: exhausted name attempts")
}

func tempSnapshotPath(articleDir string) (string, error) {
	var randomBytes [tempNameRandomBytes]byte

	_, err := rand.Read(randomBytes[:])
	if err != nil {
		return "", fmt.Errorf("artifacts: generate temp snapshot name: %w", err)
	}

	filename := tempSnapshotPrefix + hex.EncodeToString(randomBytes[:]) + tempSnapshotSuffix

	return filepath.Join(articleDir, filename), nil
}

func articleDir(articleID string) (string, error) {
	err := appartifacts.ValidateArticleID(articleID)
	if err != nil {
		return "", fmt.Errorf("artifacts: validate article dir id: %w", err)
	}

	return filepath.Join(appartifacts.ArticlesDirectoryName, articleID), nil
}
