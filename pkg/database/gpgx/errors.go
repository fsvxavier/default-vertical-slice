package gpgx

type NotConnectedError struct{}

func (nce NotConnectedError) Error() string {
	return "not connected"
}

type DbError struct {
	Message string
}

func (e DbError) Error() string {
	return e.Message
}
