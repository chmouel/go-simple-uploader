package uploader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	/// HOST where to bind the upload
	host      = "localhost"
	port      = "8080"
	directory = "./pub"
)

func errit(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(message))
}

func upload(c echo.Context) error {
	// parse and validate file and post parameters
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	path := c.FormValue("path")
	if err != nil {
		return err
	}

	// Directory traversal detection
	savepath := filepath.Join(directory, path)
	abspath, _ := filepath.Abs(savepath)
	absuploaddir, _ := filepath.Abs(directory)
	if !strings.HasPrefix(abspath, absuploaddir) {
		fmt.Println(absuploaddir, abspath)
		return echo.NewHTTPError(http.StatusForbidden, "DENIED: You should not upload outside the upload directory.")
	}

	if _, err := os.Stat(savepath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(savepath), 0755); err != nil {
			return err
		}
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(savepath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return c.HTML(
		http.StatusCreated,
		fmt.Sprintf("<h1>🚀 File has been uploaded to %s</h1>",
			savepath))
}

func Uploader() error {
	if os.Getenv("UPLOADER_DIRECTORY") != "" {
		directory = os.Getenv("UPLOADER_DIRECTORY")
	}

	if os.Getenv("UPLOADER_HOST") != "" {
		host = os.Getenv("UPLOADER_HOST")
	}

	if os.Getenv("UPLOADER_PORT") != "" {
		port = os.Getenv("UPLOADER_PORT")
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/", "public")
	e.POST("/upload", upload)

	return (e.Start(fmt.Sprintf("%s:%s", host, port)))
}
