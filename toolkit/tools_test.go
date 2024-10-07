package toolkit

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestTools_RandomString(t *testing.T) {
	var testTools Tools

	s := testTools.RandomString(10)
	if len(s) != 10 {
		t.Error("incorrect length")
	}
}

var uploadTests = []struct {
	name          string
	allowedTypes  []string
	renameFile    bool
	errorExpected bool
}{
	{name: "allowed no rename", allowedTypes: []string{"image/jpeg", "image/png"}, renameFile: false, errorExpected: false},
	{name: "allowed rename", allowedTypes: []string{"image/jpeg", "image/png"}, renameFile: true, errorExpected: false},
	{name: "not allowed", allowedTypes: []string{"image/jpeg"}, renameFile: false, errorExpected: true},
}

func TestTools_UploadFiles(t *testing.T) {
	for _, e := range uploadTests {
		pr, pw := io.Pipe()

		writer := multipart.NewWriter(pw)

		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer writer.Close()
			defer wg.Done()

			part, err := writer.CreateFormFile("file", "./testdata/img.png")

			if err != nil {
				t.Error(err)
			}

			f, err := os.Open("./testdata/img.png")

			if err != nil {
				t.Error(err)
			}

			defer f.Close()

			img, _, err := image.Decode(f)

			if err != nil {
				t.Error("error decoding", err)
			}

			err = png.Encode(part, img)

			if err != nil {
				t.Error(err)
			}
		}()

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploadedFiles, err := testTools.UploadFiles(request, "./testdata/uploads", e.renameFile)

		if err != nil && !e.errorExpected {
			t.Error(err)
		}

		if !e.errorExpected {
			if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)); os.IsNotExist(err) {
				t.Errorf("%s: expected file to exist: %s", e.name, err.Error())
			}

			_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName))
		}

		if !e.errorExpected && err != nil {
			t.Errorf("%s: error expected but none received", e.name)
		}

		wg.Wait()
	}

}

func TestTools_UploadOneFile(t *testing.T) {
	for _, e := range uploadTests {
		pr, pw := io.Pipe()

		writer := multipart.NewWriter(pw)

		go func() {
			defer writer.Close()

			part, err := writer.CreateFormFile("file", "./testdata/img.png")

			if err != nil {
				t.Error(err)
			}

			f, err := os.Open("./testdata/img.png")

			if err != nil {
				t.Error(err)
			}

			defer f.Close()

			img, _, err := image.Decode(f)

			if err != nil {
				t.Error("error decoding", err)
			}

			err = png.Encode(part, img)

			if err != nil {
				t.Error(err)
			}
		}()

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools

		uploadedFile, err := testTools.UploadOneFile(request, "./testdata/uploads", e.renameFile)

		if err != nil {
			t.Error(err)
		}

		if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFile.NewFileName)); os.IsNotExist(err) {
			t.Errorf("%s: expected file to exist: %s", e.name, err.Error())
		}

		_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFile.NewFileName))
	}
}

func TestTools_CreateDirIfNotExists(t *testing.T) {
	var testTools Tools

	err := testTools.CreateDirIfNotExists("./testdata/myDir")

	if err != nil {
		t.Error(err)
	}

	err = testTools.CreateDirIfNotExists("./testdata/myDir")

	if err != nil {
		t.Error(err)
	}

	os.Remove("./testdata/myDir")
}

var slugTests = []struct {
	name          string
	s             string
	expected      string
	errorExpected bool
}{
	{name: "valid string", s: "now is the time 123", expected: "now-is-the-time-123", errorExpected: false},
	{name: "empty string", s: "", expected: "", errorExpected: true},
	{name: "Now is the time for !!@2{){$)h}}", s: "Now is the time for !!@2{){$)h}}", expected: "now-is-the-time-for-2-h", errorExpected: false},
	{name: "japanese", s: "こんにちは", expected: "", errorExpected: true},
	{name: "japanese", s: "hello worldこんにちは", expected: "hello-world", errorExpected: false},
}

func TestTools_Slugify(t *testing.T) {

	var testTools Tools

	for _, e := range slugTests {
		slug, err := testTools.Slugify(e.s)

		if err != nil && !e.errorExpected {
			t.Errorf("test: %s -- unexpected error: %s", e.name, err)
		}

		if !e.errorExpected && slug != e.expected {
			t.Errorf("expected %s got %s", e.expected, slug)
		}
	}

}


func TestTools_Download(t *testing.T) {
  rr := httptest.NewRecorder()

  req,_ := http.NewRequest("GET", "/", nil)

  var testTools Tools

  testTools.DownloadStaticFile(rr, req, "testdata", "pic.jpg", "puppy.jpg")


  res := rr.Result()

  defer res.Body.Close()

  if res.Header["Content-Length"][0] != "98827"{
    t.Error("incorrect content length", res.Header["Content-Length"][0])
  }


  if res.Header["Content-Disposition"][0] != "attachment; filename=\"puppy.jpg\"" {
    t.Error("wrong content-diposition")
  }


  _, err := ioutil.ReadAll(res.Body)


  if err != nil {
    t.Error(err)
  }


    

}
