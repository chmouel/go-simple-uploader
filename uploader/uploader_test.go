package uploader

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func httpUploadMultiPart(t *testing.T, tempdir, s, p string) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "hello.txt")
	_, _ = part.Write([]byte(s))
	_ = writer.WriteField("path", p)
	_ = writer.Close()

	r, _ := http.NewRequest(http.MethodPost, "/upload", body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	return r
}

func TestMultipleDirectory(t *testing.T) {
	tempdir, _ := ioutil.TempDir("", "test-uploader")
	expectedSring := "HELLO MOTO"
	targetPath := "a/foo/bar/moto.txt"

	e := echo.New()
	req := httpUploadMultiPart(t, tempdir, expectedSring, targetPath)
	rec := httptest.NewRecorder()

	directory = tempdir

	context := e.NewContext(req, rec)

	if assert.Nil(t, upload(context)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
	}

	dat, err := ioutil.ReadFile(filepath.Join(tempdir, targetPath))
	assert.Nil(t, err)
	assert.Equal(t, string(dat), expectedSring)
}

func TestUploaderSimple(t *testing.T) {
	tempdir, _ := ioutil.TempDir("", "test-uploader")
	expectedSring := "HELLO SIMPLE MOTO"
	targetPath := "moto.txt"

	e := echo.New()
	req := httpUploadMultiPart(t, tempdir, expectedSring, targetPath)
	rec := httptest.NewRecorder()

	directory = tempdir
	context := e.NewContext(req, rec)

	if assert.Nil(t, upload(context)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
	}

	dat, err := ioutil.ReadFile(filepath.Join(tempdir, targetPath))
	assert.Nil(t, err)
	assert.Equal(t, string(dat), expectedSring)
}

func TestUploaderTraversal(t *testing.T) {
	tempdir, _ := ioutil.TempDir("", "test-uploader")
	expectedSring := "HELLO MOTO"
	targetPath := "../../../../../../../../../../etc/passwd"

	e := echo.New()
	req := httpUploadMultiPart(t, tempdir, expectedSring, targetPath)
	rec := httptest.NewRecorder()

	directory = tempdir

	context := e.NewContext(req, rec)
	err := upload(context)
	if assert.Error(t, err) {
		he, ok := err.(*echo.HTTPError)
		if ok {
			assert.Equal(t, http.StatusForbidden, he.Code)
		}
	}

}
