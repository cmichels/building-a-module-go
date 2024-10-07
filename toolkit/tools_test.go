package toolkit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
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

	req, _ := http.NewRequest("GET", "/", nil)

	var testTools Tools

	testTools.DownloadStaticFile(rr, req, "testdata", "pic.jpg", "puppy.jpg")

	res := rr.Result()

	defer res.Body.Close()

	if res.Header["Content-Length"][0] != "98827" {
		t.Error("incorrect content length", res.Header["Content-Length"][0])
	}

	if res.Header["Content-Disposition"][0] != "attachment; filename=\"puppy.jpg\"" {
		t.Error("wrong content-diposition")
	}

	_, err := io.ReadAll(res.Body)

	if err != nil {
		t.Error(err)
	}
}

var jsonTests = []struct {
	name          string
	json          string
	errorExpected bool
	maxSize       int
	allowUnknown  bool
}{
	{name: "good json", json: `{"foo":"bar"}`,
		errorExpected: false, maxSize: 1024, allowUnknown: false},
	{name: "bad json", json: `sateawef`,
		errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "incorrect type", json: `{"foo":1}`,
		errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "2 json files", json: `{"foo":"bar"}{"foo":"bar"}`,
		errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "empty body", json: ``,
		errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "syntax json error", json: `{"foo":1"}`,
		errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "unknown fields", json: `{"food":"bar"}`,
		errorExpected: true, maxSize: 1024, allowUnknown: false},
	{name: "allow unknown field", json: `{"food":"bar"}`,
		errorExpected: false, maxSize: 1024, allowUnknown: true},
	{name: "missing field name", json: `{foo:"bar"}`,
		errorExpected: true, maxSize: 1024, allowUnknown: true},
	{name: "file too large", json: `{"foo":"bar"}`,
		errorExpected: false, maxSize: 5, allowUnknown: false},
}

func TestTools_ReadJson(t *testing.T) {

	var testTool Tools

	for _, e := range jsonTests {
		testTool.MaxFileSize = e.maxSize
		testTool.AllowUnknownFields = e.allowUnknown

		var decodedJson struct {
			Foo string `json:"foo"`
		}

		req, err := http.NewRequest("POST", "/", bytes.NewReader([]byte(e.json)))

		if err != nil {
			t.Errorf("test: %s caused by %s", e.name, err)
		}

		rr := httptest.NewRecorder()
		err = testTool.ReadJSON(rr, req, &decodedJson)

		if e.errorExpected && err == nil {
			t.Errorf("%s: err expected", e.name)
		}

		if !e.errorExpected && err != nil {
			t.Errorf("%s: err not expected", e.name)
		}

		req.Body.Close()

	}
}

func TestTools_WriteJson(t *testing.T) {

	var testTool Tools

	rr := httptest.NewRecorder()
	payload := JSONResponse{
		Error:   false,
		Message: "foo",
	}

	headers := make(http.Header)
	headers.Add("foo", "bar")

	err := testTool.WriteJSON(rr, http.StatusOK, payload, headers)

	if err != nil {
		t.Errorf("failed to write json %s", err)
	}
}

func TestTools_ErrorJson(t *testing.T) {

	var testTool Tools

	rr := httptest.NewRecorder()
	status := http.StatusNotFound

	err := testTool.ErrorJSON(rr, errors.New("some error"), status)

	if err != nil {
		t.Error(err)
	}

	var payload JSONResponse
	decoder := json.NewDecoder(rr.Body)

	err = decoder.Decode(&payload)

	if err != nil {
		t.Errorf("error endocing %s", err)
	}

	if !payload.Error {
		t.Error("expected true recieved fail")
	}

	if rr.Code != status {
		t.Errorf("expected %d. received %d", status, rr.Code)
	}

}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestTools_PushJsonToRemote(t *testing.T) {

	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString("ok")),
			Header:     make(http.Header),
		}
	})

	var testTool Tools

	var foo struct {
		Bar string `json:"bar"`
	}

	foo.Bar = "bar"

	_, _, err := testTool.PushJSONToRemote("http://example.com/some/path", foo, client)

	if err != nil {
		t.Errorf("failed to call remote service: %s", err)
	}
}
