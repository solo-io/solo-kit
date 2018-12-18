package util

type ErrorWriter struct {
	error string
}

func (ew *ErrorWriter) Flush() {
	ew.error = ""
}

func (ew *ErrorWriter) Write(p []byte) (n int, err error) {
	ew.error += string(p)
	return len(p), nil
}

func (ew *ErrorWriter) Error() string {
	return ew.error
}
