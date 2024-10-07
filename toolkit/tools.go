package toolkit

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Tools is the type to instantiate this module.
// Any variable of this type will have access to
// all the methods with receiver *Tools
type Tools struct {
	MaxFileSize      int
	AllowedFileTypes []string
}

// RandomString() returns a string of random characters
// of the length n, using randomStringSource
func (t Tools) RandomString(n int) string {

	s, r := make([]rune, n), []rune(randomStringSource)

	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))

		s[i] = r[x%y]
	}

	return string(s)
}

type UploadedFile struct {
	NewFileName      string
	OriginalFileName string
	FileSize         int64
}

func (t *Tools) UploadOneFile(r *http.Request, uploadDir string, rename ...bool) (*UploadedFile, error) {
	if files, err := t.UploadFiles(r, uploadDir, rename...); err != nil {
		return nil, err
	} else {
		return files[0], nil
	}
}

func (t *Tools) UploadFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadedFile, error) {

	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	var uploadedFiles []*UploadedFile

	if t.MaxFileSize == 0 {
		t.MaxFileSize = 1024 * 1024 * 1024
	}

	err := r.ParseMultipartForm(int64(t.MaxFileSize))

	if err != nil {
		return nil, err
	}

  if err = t.CreateDirIfNotExists(uploadDir); err != nil{
    return nil, err
  }

	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			uploadedFiles, err = func(uploadedFiles []*UploadedFile) ([]*UploadedFile, error) {
				var uploadedFile UploadedFile
				infile, err := hdr.Open()

				if err != nil {
					return nil, err
				}
				defer infile.Close()

				buff := make([]byte, 512)
				_, err = infile.Read(buff)

				if err != nil {
					return nil, err
				}

				allowed := true

				fileType := http.DetectContentType(buff)

				if len(t.AllowedFileTypes) > 0 {
					allowed = slices.Contains(t.AllowedFileTypes, fileType)
				}

				if !allowed {
					return nil, fmt.Errorf("file type [%s] is not allowed", fileType)
				}

				_, err = infile.Seek(0, 0)

				if err != nil {
					return nil, err
				}

				if renameFile {
					uploadedFile.NewFileName = fmt.Sprintf("%s%s", t.RandomString(25), filepath.Ext(hdr.Filename))
				} else {
					uploadedFile.NewFileName = hdr.Filename
				}

				uploadedFile.OriginalFileName = hdr.Filename
				var outfile *os.File

				defer outfile.Close()

				if outfile, err = os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName)); err != nil {
					return nil, err
				} else {
					fileSize, err := io.Copy(outfile, infile)
					if err != nil {
						return nil, err
					}

					uploadedFile.FileSize = fileSize
				}

				uploadedFiles = append(uploadedFiles, &uploadedFile)

				return uploadedFiles, nil

			}(uploadedFiles)

			if err != nil {
				return uploadedFiles, err
			}
		}
	}

	return uploadedFiles, nil
}

func (t *Tools) CreateDirIfNotExists(path string) error {
	const mode = 0755

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, mode)
		if err != nil {
			return err
		}
	}
	return nil
}


func (t *Tools) Slugify(s string) (string, error) {
  if s == ""{
    return "", errors.New("empty string not permitted")
  }


  var re = regexp.MustCompile(`[^a-z\d]+`)


  lower := strings.ToLower(s)
  clean := re.ReplaceAllLiteralString(lower, "-")
  slug := strings.Trim(clean,"-")


  if len(slug) == 0{
    return "", errors.New("slug created with a length of zero")
  }


  return slug,nil 
}
