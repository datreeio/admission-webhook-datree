package responseWriter

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWrite(t *testing.T) {
	responseRecorder := httptest.NewRecorder()
	responseWriter := New(responseRecorder)

	responseContent := "This is a test"
	responseWriter.Write(responseContent)

	assert.Equal(t, strings.TrimSpace(responseRecorder.Body.String()), responseContent)
}

func TestNotAllowed(t *testing.T) {
	responseRecorder := httptest.NewRecorder()
	responseWriter := New(responseRecorder)

	responseBodyErr := "Not allowed"
	responseWriter.NotAllowed(responseBodyErr)

	assert.Equal(t, responseRecorder.Code, http.StatusMethodNotAllowed)
	assert.Equal(t, strings.TrimSpace(responseRecorder.Body.String()), responseBodyErr)
}

func TestBadRequest(t *testing.T) {
	responseRecorder := httptest.NewRecorder()
	responseWriter := New(responseRecorder)

	responseBodyErr := "Bad request"
	responseWriter.BadRequest(responseBodyErr)

	assert.Equal(t, responseRecorder.Code, http.StatusBadRequest)
	assert.Equal(t, strings.TrimSpace(responseRecorder.Body.String()), responseBodyErr)
}

func TestWriteBody(t *testing.T) {
	type TestObject struct {
		UID string
	}

	responseRecorder := httptest.NewRecorder()
	responseWriter := New(responseRecorder)

	testObject := &TestObject{UID: "sdfsdwef-123124-sdfsdfs-213123"}
	responseWriter.WriteBody(testObject)

	testObjectJson, _ := json.Marshal(testObject)
	assert.Equal(t, responseRecorder.Code, http.StatusOK)
	assert.Equal(t, strings.TrimSpace(responseRecorder.Body.String()), string(testObjectJson))
}
