package uploader

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func httpUploadMultiPart(s, p string) *http.Request {
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
	req := httpUploadMultiPart(expectedSring, targetPath)
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
	req := httpUploadMultiPart(expectedSring, targetPath)
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
	req := httpUploadMultiPart(expectedSring, targetPath)
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

func TestUploaderDelete(t *testing.T) {
	tempdir, _ := ioutil.TempDir("", "test-uploader")
	directory = tempdir
	fpath := filepath.Join(tempdir, "foo.txt")

	fp, err := os.Create(fpath)
	assert.Nil(t, err)
	fp.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("path", "foo.txt")
	_ = writer.Close()

	e := echo.New()
	req, _ := http.NewRequest(http.MethodDelete, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	context := e.NewContext(req, rec)
	if assert.Nil(t, uploaderDelete(context)) {
		assert.Equal(t, http.StatusAccepted, rec.Code)
		if _, err = os.Stat(fpath); err != nil {
			assert.True(t, os.IsNotExist(err))
		}
	}
}
