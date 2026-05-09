package markdown

const (
	ProviderJina Provider = "jina"

	ErrorCodeJinaFailed             ErrorCode = "ARC-010"
	ErrorCodeJinaInsufficientCredit ErrorCode = "ARC-011"
)

type jinaError struct {
	code   ErrorCode
	reason string
}

func jinaFailure(code ErrorCode, reason string) ExtractResult {
	return ExtractResult{
		Status:        ResultStatusFailure,
		Provider:      ProviderJina,
		ErrorCode:     code,
		FailureReason: reason,
	}
}

func localFailure(reason string) ExtractResult {
	return ExtractResult{
		Status:        ResultStatusFailure,
		Provider:      ProviderGoReadability,
		ErrorCode:     ErrorCodeLocalExtractionFailed,
		FailureReason: reason,
	}
}
