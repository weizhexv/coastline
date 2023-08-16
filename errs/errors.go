package errs

type ApiErr struct {
	Code    int    `json:"Code"`
	Message string `json:"Message"`
}

var (
	ErrSystem    = system()
	ErrAuth      = auth()
	ErrGateway   = gateway()
	ErrForbidden = forbidden()
)

func (e *ApiErr) Error() string {
	return e.Message
}

func system() ApiErr {
	return apiErr(1, "System Error")
}

func auth() ApiErr {
	return apiErr(2, "Need Login")
}

func gateway() ApiErr {
	return apiErr(3, "Bad Gateway")
}

func forbidden() ApiErr {
	return apiErr(4009, "No Permission")
}

func apiErr(code int, msg string) ApiErr {
	return ApiErr{
		Code:    code,
		Message: msg,
	}
}
