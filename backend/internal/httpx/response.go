package httpx

import (
	"encoding/json"
	"net/http"

	"gocamping/internal/logger"
)

type Envelope struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	TraceID string      `json:"trace_id"`
}

func JSON(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	env := Envelope{
		Code:    CodeOK,
		Message: "ok",
		Data:    data,
		TraceID: TraceID(r),
	}
	if data == nil {
		env.Data = map[string]interface{}{}
	}
	write(w, status, env)
}

func Fail(w http.ResponseWriter, r *http.Request, err error) {
	ae, ok := AsApp(err)
	if !ok {
		logger.Error("unhandled error", "err", err, "trace_id", TraceID(r))
		ae = Internal("内部错误")
	}
	env := Envelope{
		Code:    ae.Code,
		Message: ae.Message,
		Data:    map[string]interface{}{},
		TraceID: TraceID(r),
	}
	write(w, ae.HTTP, env)
}

func write(w http.ResponseWriter, status int, env Envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(env)
}
