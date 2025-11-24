package agents

type StatusCallback func(status string)

type FileValidationError struct {
	Path    string
	Message string
}

func (e *FileValidationError) Error() string {
	return e.Message
}
