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

func delete(c echo.Context) error {
	path := c.FormValue("path")

	// Directory traversal detection
	savepath := filepath.Join(directory, path)
	abspath, _ := filepath.Abs(savepath)
	absuploaddir, _ := filepath.Abs(directory)
	if !strings.HasPrefix(abspath, absuploaddir) {
		return echo.NewHTTPError(http.StatusForbidden, "DENIED: You should not upload outside the upload directory.")
	}

	if _, err := os.Stat(abspath); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Could not find your file")
	}

	if stat, err := os.Stat(abspath); err == nil && stat.IsDir() {
		return echo.NewHTTPError(http.StatusBadRequest, "Deleting a directory is not supported yet")
	}

	err := os.Remove(abspath)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Could not delete your file: %s", err)
	}

	return c.HTML(
		http.StatusAccepted,
		fmt.Sprintf("💇 File %s has been deleted", path))
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
		fmt.Sprintf("<h1>🚀 File has been uploaded to %s</h1>", path))
}

// Uploader main uploader function
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
	e.DELETE("/upload", delete)

	return (e.Start(fmt.Sprintf("%s:%s", host, port)))
}
