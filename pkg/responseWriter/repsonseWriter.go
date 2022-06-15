package responseWriter

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ResponseWriter struct {
	httpWriter http.ResponseWriter
}

func New(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		httpWriter: w,
	}
}

func (rw ResponseWriter) Write(content string) {
	fmt.Fprintf(rw.httpWriter, content+"\n")
}

func (rw ResponseWriter) WriteBody(content interface{}) {
	jout, err := json.Marshal(content)
	if err != nil {
		rw.BadRequest(err.Error())
		return
	}

	rw.Write(string(jout))
}

func (rw ResponseWriter) NotAllowed(content string) {
	http.Error(rw.httpWriter, content, http.StatusMethodNotAllowed)
}

func (rw ResponseWriter) BadRequest(content string) {
	http.Error(rw.httpWriter, content, http.StatusBadRequest)
}
