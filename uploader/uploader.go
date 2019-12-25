package uploader

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	HOST      = "localhost"
	PORT      = "8080"
	DIRECTORY = "./pub"
)

func errit(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

func upload(w http.ResponseWriter, r *http.Request) {
	// parse and validate file and post parameters
	file, _, err := r.FormFile("file")
	if err != nil {
		errit(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fmt.Println(r.RemoteAddr)

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		errit(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}

	path := r.FormValue("path")
	if err != nil {
		errit(w, "INVALID_PATH", http.StatusBadRequest)
		return
	}

	// Directory traversal detection
	savepath := filepath.Join(DIRECTORY, path)
	abspath, _ := filepath.Abs(savepath)
	absuploaddir, _ := filepath.Abs(DIRECTORY)
	if !strings.HasPrefix(abspath, absuploaddir) {
		errit(w, "INVALID_PATH", http.StatusBadGateway)
		return
	}

	if _, err := os.Stat(savepath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(savepath), 0755); err != nil {
			errit(w, "CANT_CREATE_DIR", http.StatusInternalServerError)
		}
	}

	newFile, err := os.Create(abspath)
	fmt.Println("Saving file to " + abspath)
	if err != nil {
		fmt.Println(err)
		errit(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
		return
	}
	defer newFile.Close()
	if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		errit(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("👍"))
	// save_path = filepath.Join()

}

func Uploader() error {
	if os.Getenv("UPLOADER_DIRECTORY") != "" {
		DIRECTORY = os.Getenv("UPLOADER_DIRECTORY")
	}

	if os.Getenv("UPLOADER_HOST") != "" {
		HOST = os.Getenv("UPLOADER_HOST")
	}

	if os.Getenv("UPLOADER_PORT") != "" {
		PORT = os.Getenv("UPLOADER_PORT")
	}

	http.HandleFunc("/upload", upload)
	log.Printf("Starting uploader on %s:%s", HOST, PORT)
	return (http.ListenAndServe(fmt.Sprintf("%s:%s", HOST, PORT), nil))
}
