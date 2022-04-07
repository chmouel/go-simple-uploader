package uploader

import (
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	/// HOST where to bind the upload !
	host      = "localhost"
	port      = "8080"
	directory = "./pub"
)

func uploaderDelete(c echo.Context) error {
	path := c.FormValue("path")

	// Directory traversal detection
	savePath := filepath.Join(directory, path)
	abspath, _ := filepath.Abs(savePath)
	absoluteUploadDir, _ := filepath.Abs(directory)
	if !strings.HasPrefix(abspath, absoluteUploadDir) {
		return echo.NewHTTPError(http.StatusForbidden, "DENIED: You should not upload outside the upload directory.")
	}

	if _, err := os.Stat(abspath); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Could not find your file")
	}

	err := os.RemoveAll(abspath)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Could not delete your your file: %s", err.Error()))
	}

	return c.HTML(
		http.StatusAccepted,
		fmt.Sprintf("File %s has been deleted 💇", path))
}

func upload(c echo.Context) error {
	// parse and validate file and post parameters
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	untargz := c.FormValue("targz")
	path := c.FormValue("path")
	// Directory traversal detection
	savepath := filepath.Join(directory, path)
	abspath, _ := filepath.Abs(savepath)
	absuploaddir, _ := filepath.Abs(directory)
	if !strings.HasPrefix(abspath, absuploaddir) {
		return echo.NewHTTPError(http.StatusForbidden, "DENIED: You should not upload outside the upload directory.")
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	if untargz != "" {
		if err := os.MkdirAll(savepath, 0o755); err != nil {
			return err
		}
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		err = UntarGz(abspath, src)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		return c.HTML(http.StatusCreated, fmt.Sprintf("File has been uploaded to %s 🚀\n", path))
	}

	if _, err := os.Stat(savepath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(savepath), 0o755); err != nil {
			return err
		}
	}

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
		fmt.Sprintf("File has been uploaded to %s 🚀\n", path))
}

func lastModified(c echo.Context) error {
	path := c.Param("path")
	filePath := filepath.Join(directory, path)
	abspath, _ := filepath.Abs(filePath)
	absoluteUploadDir, _ := filepath.Abs(directory)
	if !strings.HasPrefix(abspath, absoluteUploadDir) {
		return echo.NewHTTPError(http.StatusForbidden, "DENIED: You should not try to get outside the root directory.")
	}

	info, err := os.Stat(abspath)
	if err != nil {
		return echo.NotFoundHandler(c)
	}

	c.Response().Header().Set(echo.HeaderLastModified, info.ModTime().UTC().Format(http.TimeFormat))
	return c.NoContent(http.StatusOK)
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

	e.HEAD("/:path", lastModified)
	e.Static("/", directory)
	e.POST("/upload", upload)
	e.DELETE("/upload", uploaderDelete)

	if os.Getenv("UPLOADER_UPLOAD_CREDENTIALS") != "" {
		creds := strings.Split(os.Getenv("UPLOADER_UPLOAD_CREDENTIALS"), ":")
		c := middleware.DefaultBasicAuthConfig
		c.Validator = (func(username, password string, c echo.Context) (bool, error) {
			if subtle.ConstantTimeCompare([]byte(username), []byte(creds[0])) == 1 &&
				subtle.ConstantTimeCompare([]byte(password), []byte(strings.Join(creds[1:], ":"))) == 1 {
				return true, nil
			}
			return false, nil
		})
		c.Skipper = func(c echo.Context) bool {
			if c.Request().URL.String() != "/upload" {
				return true
			}
			return false
		}
		e.Use(middleware.BasicAuthWithConfig(c))
	}

	return (e.Start(fmt.Sprintf("%s:%s", host, port)))
}
