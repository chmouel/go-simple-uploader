package uploader

import (
	"crypto/subtle"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
	abspath, err := filepath.Abs(savePath)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid path")
	}
	absoluteUploadDir, err := filepath.Abs(directory)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine upload directory")
	}

	// Ensure we don't match prefixes like /data vs /database
	if !strings.HasPrefix(abspath, absoluteUploadDir+string(os.PathSeparator)) && abspath != absoluteUploadDir {
		return echo.NewHTTPError(http.StatusForbidden, "DENIED: You should not upload outside the upload directory.")
	}

	if _, err := os.Stat(abspath); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Could not find your file")
	}

	err = os.RemoveAll(abspath)
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
	abspath, err := filepath.Abs(savepath)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid path")
	}
	absuploaddir, err := filepath.Abs(directory)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine upload directory")
	}

	// Ensure we don't match prefixes like /data vs /database
	if !strings.HasPrefix(abspath, absuploaddir+string(os.PathSeparator)) && abspath != absuploaddir {
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
	abspath, err := filepath.Abs(filePath)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid path")
	}
	absoluteUploadDir, err := filepath.Abs(directory)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine upload directory")
	}

	// Ensure we don't match prefixes like /data vs /database
	if !strings.HasPrefix(abspath, absoluteUploadDir+string(os.PathSeparator)) && abspath != absoluteUploadDir {
		return echo.NewHTTPError(http.StatusForbidden, "DENIED: You should not try to get outside the root directory.")
	}

	info, err := os.Stat(abspath)
	if err != nil {
		return echo.NotFoundHandler(c)
	}

	c.Response().Header().Set(echo.HeaderLastModified, info.ModTime().UTC().Format(http.TimeFormat))
	return c.NoContent(http.StatusOK)
}

func deleteOldFilesOfDir(c echo.Context) error {
	path := c.FormValue("path")
	days, _ := strconv.Atoi(c.FormValue("days"))
	recursive_flag := c.FormValue("recursive")

	if len(recursive_flag) == 0 {
		recursive_flag = "false"
	}

	recursive, err := strconv.ParseBool(recursive_flag)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "DENIED: check if your formvalue recursive should be any of this ('true', 'True', 'false','False','TRUE','FALSE','f','t','F','T', '') ")
	}

	filePath := filepath.Join(directory, path)
	abspath, err := filepath.Abs(filePath)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid path")
	}
	absoluteUploadDir, err := filepath.Abs(directory)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine upload directory")
	}

	// Ensure we don't match prefixes like /data vs /database
	if !strings.HasPrefix(abspath, absoluteUploadDir+string(os.PathSeparator)) && abspath != absoluteUploadDir {
		return echo.NewHTTPError(http.StatusForbidden, "DENIED: You should not try to get outside the root directory.")
	}

	_, err = os.Stat(abspath)
	if err != nil {
		return echo.NotFoundHandler(c)
	}

	files, err := findFilesOlderThanXDays(abspath, days, recursive)
	if err != nil {
		return echo.NotFoundHandler(c)
	}

	if len(files) == 0 {
		return c.HTML(
			http.StatusAccepted,
			fmt.Sprintf("There are NO Old Files more than %d days to be deleted 💇", days))
	}

	for _, file := range files {
		NewfilePath := filepath.Join(abspath, file.Name())
		Newabspath, _ := filepath.Abs(NewfilePath)

		if recursive && file.IsDir() {
			err = os.RemoveAll(Newabspath)
		} else {
			err = os.Remove(Newabspath)
		}

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Could not delete your your file: %s", err.Error()))
		}
	}

	if recursive {
		return c.HTML(
			http.StatusAccepted,
			fmt.Sprintf("Old Files/child directories more than %d days has been deleted 💇", days))
	}
	return c.HTML(
		http.StatusAccepted,
		fmt.Sprintf("Old Files more than %d days has been deleted 💇", days))
}

func isOlderThanXDays(t time.Time, days int) bool {
	return time.Since(t) > (time.Duration(days) * 24 * time.Hour)
}

func findFilesOlderThanXDays(dir string, days int, recursive bool) (files []os.FileInfo, err error) {
	tmpfiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range tmpfiles {
		if file.Mode().IsRegular() || (recursive && file.IsDir()) {
			if isOlderThanXDays(file.ModTime(), days) {
				files = append(files, file)
			}
		}
	}
	return files, nil
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

	e.Static("/", directory)
	e.HEAD("/:path", lastModified)
	e.POST("/upload", upload)
	e.DELETE("/upload", uploaderDelete)
	e.DELETE("/delete", deleteOldFilesOfDir)

	if os.Getenv("UPLOADER_UPLOAD_CREDENTIALS") != "" {
		creds := strings.Split(os.Getenv("UPLOADER_UPLOAD_CREDENTIALS"), ":")
		c := middleware.DefaultBasicAuthConfig
		c.Skipper = func(c echo.Context) bool {
			if (c.Request().Method == "HEAD" || c.Request().Method == "GET") && c.Path() != "/upload" && c.Path() != "/delete" {
				return true
			}
			return false
		}
		c.Validator = (func(username, password string, c echo.Context) (bool, error) {
			if subtle.ConstantTimeCompare([]byte(username), []byte(creds[0])) == 1 &&
				subtle.ConstantTimeCompare([]byte(password), []byte(strings.Join(creds[1:], ":"))) == 1 {
				return true, nil
			}
			return false, nil
		})
		e.Use(middleware.BasicAuthWithConfig(c))
	}

	return (e.Start(fmt.Sprintf("%s:%s", host, port)))
}
