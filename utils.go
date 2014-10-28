package couchdb

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Get mime type from file name.
func mimeType(name string) string {
	ext := filepath.Ext(name)
	return mime.TypeByExtension(ext)
}

// Convert HTTP response from CouchDB into Error.
func newError(res *http.Response) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	error := &Error{}
	err = json.Unmarshal(body, &error)
	if err != nil {
		return err
	}
	error.Method = res.Request.Method
	error.Url = res.Request.URL.String()
	error.StatusCode = res.StatusCode
	return error
}

// Quote string values because CouchDB needs those double quotes in query params.
func quote(values url.Values) url.Values {
	for key, value := range values {
		if key == "startkey" || key == "endkey" && value != nil {
			arr := strings.Split(value[0], ",")
			for index, element := range arr {
				_, err := strconv.ParseFloat(element, 64)
				if err != nil {
					arr[index] = strconv.Quote(element)
				}
			}
			joined := strings.Join(arr, ",")
			enclosed := "[" + joined + "]"
			values.Set(key, enclosed)
		} else if value[0] != "true" && value[0] != "false" {
			values.Set(key, strconv.Quote(value[0]))
		}
	}
	return values
}

// Create new CouchDB response for any document method.
func newDocumentResponse(body io.ReadCloser) (*DocumentResponse, error) {
	response := &DocumentResponse{}
	return response, json.NewDecoder(body).Decode(&response)
}

// Create new CouchDB response for any database method.
func newDatabaseResponse(body io.ReadCloser) (*DatabaseResponse, error) {
	response := &DatabaseResponse{}
	return response, json.NewDecoder(body).Decode(&response)
}

// Write JSON to multipart/related.
func writeJSON(document *Document, writer *multipart.Writer, file *os.File) error {
	partHeaders := textproto.MIMEHeader{}
	partHeaders.Set("Content-Type", "application/json")
	part, err := writer.CreatePart(partHeaders)
	if err != nil {
		return err
	}

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	path := file.Name()

	// make empty map
	document.Attachments = make(map[string]Attachment)
	attachment := Attachment{
		Follows:     true,
		ContentType: mimeType(path),
		Length:      stat.Size(),
	}
	// add attachment to map
	filename := filepath.Base(path)
	document.Attachments[filename] = attachment

	bytes, err := json.Marshal(document)
	if err != nil {
		return err
	}

	_, err = part.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

// Write actual file content to multipart/related.
func writeMultipart(writer *multipart.Writer, file *os.File) error {
	part, err := writer.CreatePart(textproto.MIMEHeader{})
	if err != nil {
		return err
	}

	// copy file content into multipart message
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	return nil
}
