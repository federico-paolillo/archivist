// Package pipeline orchestrates article-processing pipeline stages.
package pipeline

// PipelineError carries stage and job context while preserving both the ARC
// classification and any lower-level diagnostic cause.
type PipelineError struct {
	Stage     string
	Op        string
	ArticleID string
	JobID     string
	URL       string
	Err       error
	Cause     error
}

func (e *PipelineError) Error() string {
	if e == nil {
		return ""
	}

	msg := "pipeline"
	if e.Stage != "" {
		msg += ": stage=" + e.Stage
	}

	if e.Op != "" {
		msg += ": op=" + e.Op
	}

	if e.ArticleID != "" {
		msg += ": article_id=" + e.ArticleID
	}

	if e.JobID != "" {
		msg += ": job_id=" + e.JobID
	}

	if e.URL != "" {
		msg += ": url=" + e.URL
	}

	if e.Err != nil {
		msg += ": " + e.Err.Error()
	}

	if e.Cause != nil {
		msg += ": cause=" + e.Cause.Error()
	}

	return msg
}

func (e *PipelineError) Unwrap() []error {
	if e == nil {
		return nil
	}

	errs := make([]error, 0, 2)
	if e.Err != nil {
		errs = append(errs, e.Err)
	}

	if e.Cause != nil {
		errs = append(errs, e.Cause)
	}

	return errs
}

func pipelineFailure(stage string, op string, err error, cause error, opts ...func(*PipelineError)) error {
	pipelineErr := &PipelineError{
		Stage: stage,
		Op:    op,
		Err:   err,
		Cause: cause,
	}
	for _, opt := range opts {
		opt(pipelineErr)
	}

	return pipelineErr
}

func withJobContext(articleID, jobID string) func(*PipelineError) {
	return func(err *PipelineError) {
		err.ArticleID = articleID
		err.JobID = jobID
	}
}

func withPipelineURL(url string) func(*PipelineError) {
	return func(err *PipelineError) {
		err.URL = url
	}
}
