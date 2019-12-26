package uploader

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func httpUploadMultiPart(t *testing.T, tempdir, s, p string) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "hello.txt")
	_, _ = part.Write([]byte(s))
	_ = writer.WriteField("path", p)
	_ = writer.Close()

	r, _ := http.NewRequest("POST", "/upload", body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	return r
}

func TestMultipleDirectory(t *testing.T) {
	tempdir, _ := ioutil.TempDir("", "test-uploader")
	expectedSring := "HELLO MOTO"
	targetPath := "a/foo/bar/moto.txt"

	r := httpUploadMultiPart(t, tempdir, expectedSring, targetPath)

	directory = tempdir

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(uploadHandler)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("Test didn't come back with OK")
	}

	dat, err := ioutil.ReadFile(filepath.Join(tempdir, targetPath))
	if err != nil {
		t.Fatal(err)
	}
	if string(dat) != expectedSring {
		t.Fatal("File didn't upload properly")
	}
}

func TestUploaderSimple(t *testing.T) {
	tempdir, _ := ioutil.TempDir("", "test-uploader")
	expectedSring := "HELLO MOTO"
	targetPath := "moto.txt"

	r := httpUploadMultiPart(t, tempdir, expectedSring, targetPath)

	directory = tempdir

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(uploadHandler)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("Test didn't come back with OK")
	}

	dat, err := ioutil.ReadFile(filepath.Join(tempdir, targetPath))
	if err != nil {
		t.Fatal(err)
	}
	if string(dat) != expectedSring {
		t.Fatal("File didn't upload properly")
	}
}

func TestUploaderTraversal(t *testing.T) {
	tempdir, _ := ioutil.TempDir("", "test-uploader")
	expectedSring := "HELLO MOTO"
	targetPath := "../../../../etc/passwd"

	r := httpUploadMultiPart(t, tempdir, expectedSring, targetPath)

	directory = tempdir

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(uploadHandler)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Test didn't come back with OK: %d", resp.StatusCode)
	}
}
