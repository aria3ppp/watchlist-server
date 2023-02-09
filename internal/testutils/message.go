package testutils

type ErrorMessageTest struct {
	Message string `json:"message"`
}

func ErrorMessage(message string) ErrorMessageTest {
	return ErrorMessageTest{Message: message}
}
