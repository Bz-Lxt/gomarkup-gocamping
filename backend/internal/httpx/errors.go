package httpx

// Application error codes. Keep stable for clients and docs/API.md.
const (
	CodeOK              = 0
	CodeBadRequest      = 40000
	CodeValidation      = 40001
	CodeUnauthorized    = 40100
	CodeForbidden       = 40300
	CodeNotFound        = 40400
	CodeConflict        = 40900
	CodeState           = 40901
	CodePayloadTooLarge = 41300
	CodeInternal        = 50000
	CodeUpstream        = 50200
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	HTTP    int    `json:"-"`
}

func (e *AppError) Error() string { return e.Message }

func New(code, httpStatus int, msg string) *AppError {
	return &AppError{Code: code, Message: msg, HTTP: httpStatus}
}

func BadRequest(msg string) *AppError      { return New(CodeBadRequest, 400, msg) }
func Validation(msg string) *AppError      { return New(CodeValidation, 400, msg) }
func Unauthorized(msg string) *AppError    { return New(CodeUnauthorized, 401, msg) }
func Forbidden(msg string) *AppError       { return New(CodeForbidden, 403, msg) }
func NotFound(msg string) *AppError        { return New(CodeNotFound, 404, msg) }
func Conflict(msg string) *AppError        { return New(CodeConflict, 409, msg) }
func BadState(msg string) *AppError        { return New(CodeState, 409, msg) }
func TooLarge(msg string) *AppError        { return New(CodePayloadTooLarge, 413, msg) }
func Internal(msg string) *AppError        { return New(CodeInternal, 500, msg) }
func Upstream(msg string) *AppError        { return New(CodeUpstream, 502, msg) }

func AsApp(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}
	if ae, ok := err.(*AppError); ok {
		return ae, true
	}
	return nil, false
}
