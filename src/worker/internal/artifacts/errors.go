package artifacts

import (
	"errors"
)

var (
	ErrStore                 = errors.New("artifacts: store error")
	ErrEmptyDataDir          = errors.New("artifacts: data dir is empty")
	ErrInvalidArticleID      = errors.New("article id must be a ULID path segment")
	ErrInvalidTempPattern    = errors.New("artifacts: temp pattern must contain a wildcard")
	ErrTempNameCreationLimit = errors.New("artifacts: temp file name creation limit reached")
)

// StoreError carries artifact-store operation metadata while preserving the
// underlying filesystem or validation error for errors.Is / errors.As callers.
type StoreError struct {
	Op        string
	ArticleID string
	Filename  string
	Path      string
	Err       error
}

func (e *StoreError) Error() string {
	if e == nil {
		return ""
	}

	msg := "artifacts"
	if e.Op != "" {
		msg += ": " + e.Op
	}

	if e.ArticleID != "" {
		msg += ": article_id=" + e.ArticleID
	}

	if e.Filename != "" {
		msg += ": filename=" + e.Filename
	}

	if e.Path != "" {
		msg += ": path=" + e.Path
	}

	if e.Err != nil {
		msg += ": " + e.Err.Error()
	}

	return msg
}

func (e *StoreError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func (e *StoreError) Is(target error) bool {
	return target == ErrStore
}

func storeFailure(op string, err error, opts ...func(*StoreError)) error {
	storeErr := &StoreError{
		Op:  op,
		Err: err,
	}
	for _, opt := range opts {
		opt(storeErr)
	}

	return storeErr
}

func withArticleID(articleID string) func(*StoreError) {
	return func(err *StoreError) {
		err.ArticleID = articleID
	}
}

func withFilename(filename string) func(*StoreError) {
	return func(err *StoreError) {
		err.Filename = filename
	}
}

func withPath(path string) func(*StoreError) {
	return func(err *StoreError) {
		err.Path = path
	}
}
