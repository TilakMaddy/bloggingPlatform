package internals

import (
	"fmt"
	"github.com/google/uuid"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type UtilError struct {
	message    string
	statusCode int // should be 0 to indicate no error
}

func (e *UtilError) IsError() bool {
	return e.statusCode != 0
}

func convertReqFormToBlog(r *http.Request) (Blog, UtilError) {

	// to populate r.MultipartForm
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return Blog{}, UtilError{"parse failed", http.StatusInternalServerError}
	}

	form := r.MultipartForm

	//goland:noinspection ALL
	defer form.RemoveAll()

	// populate the below fields from multipart-form data
	var (
		title    string
		authorId int64
		content  string
		images   []string
	)

	var err error

	title = form.Value["title"][0]
	if len(strings.TrimSpace(title)) == 0 {
		return Blog{}, UtilError{"title is empty", http.StatusBadRequest}
	}

	if authorId, err = strconv.ParseInt(form.Value["author_id"][0], 10, 64); err != nil || authorId <= 0 {
		return Blog{}, UtilError{"authorId is not a valid number", http.StatusBadRequest}
	}

	content = form.Value["content"][0]
	if len(strings.TrimSpace(content)) == 0 {
		return Blog{}, UtilError{"content is empty", http.StatusBadRequest}
	}

	imagesRaw := form.File["images"]
	if images, err = downloadImages(imagesRaw, strconv.FormatInt(authorId, 10)); err != nil {
		return Blog{}, UtilError{"images could not be processed", http.StatusInternalServerError}
	}

	// ready to be published by delegating to publishBlog(...)
	blog := Blog{
		Title:    title,
		Content:  content,
		Images:   images,
		AuthorID: authorId,
	}

	return blog, UtilError{} // by default statusCode = 0, so it will be no error
}

// Make a permanent location for the images that have been uploaded
func downloadImages(imagesRaw []*multipart.FileHeader, subFolderName string) ([]string, error) {

	var downloadedImages []string

	// todo: when you fail downloading an image, delete all the previous ones
	for _, imageHeader := range imagesRaw {
		file, err := imageHeader.Open()
		if err != nil {
			return nil, err
		}

		//goland:noinspection ALL
		defer file.Close()

		suffix, err := extractSuffix(imageHeader.Filename)
		if err != nil {
			return nil, err
		}

		randFileName := randomFileName(suffix)
		err = createFileAndCopy(randFileName, file, subFolderName)
		if err != nil {
			return nil, err
		}

		downloadedImages = append(downloadedImages, randFileName)
	}

	return downloadedImages, nil

}

// To preserve the extension of the image when generating a new name for it
// Ex: amish.jpg should become <randomUUID>.jpg similarly for .png, .avif, etc
func extractSuffix(filename string) (string, error) {
	filenameChunks := strings.Split(filename, ".")
	if len(filenameChunks) <= 1 {
		return "", fmt.Errorf("file doesn't have a extension")
	}
	return filenameChunks[len(filenameChunks)-1], nil
}

// To (mostly) prevent file name collisions
func randomFileName(suffix string) string {
	return uuid.New().String() + "." + suffix
}

// Decides the location of upload in the local file system
func createFileAndCopy(destFile string, orgFile multipart.File, subFolderName string) error {

	dirPath := filepath.Join(os.Getenv("UPLOAD_DIR"), subFolderName)

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.Mkdir(dirPath, 0777)
		if err != nil {
			return err
		}
	}

	file, err := os.Create(filepath.Join(dirPath, destFile))
	if err != nil {
		return err
	}
	_, err = io.Copy(file, orgFile)
	return err
}

// Convert array to string representation that can be safely inserted to MySQL
func stringifyToMySQLJSONArray(images []string) string {
	var stringBuilder strings.Builder
	stringBuilder.WriteString("[")
	for i, image := range images {
		if i != 0 {
			stringBuilder.WriteString(",")
		}
		stringBuilder.WriteString("\"" + image + "\"")
	}
	stringBuilder.WriteString("]")
	return stringBuilder.String()
}
