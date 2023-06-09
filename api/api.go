package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

type File struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Bytes     int    `json:"bytes"`
	CreatedAt int    `json:"created_at"`
	Filename  string `json:"filename"`
	Purpose   string `json:"purpose"`
}

type FileList struct {
	Data   []File `json:"data"`
	Object string `json:"object"`
}

// The function sends a GET request with a bearer token and returns the response body as a byte array.
func get_request(bearer string, url string) []byte {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("[GetRequest] Failed to create get :", err)
	}

	req.Header.Add("Authorization", "Bearer "+bearer)
	req.Header.Set("User-Agent", "curl/8.0.1")
	req.Header.Set("accept", "*/*")
	log.Println("[GetRequest] Ready to do Get on ", url, " with ", bearer[0:7])

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("[GetRequest] Failed to do the request :", err)
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("[GetRequest] Read", len(bodyBytes), "bytes")

	if resp.StatusCode != http.StatusOK {
		log.Println("[GetRequest] Failed with error code ", resp.Status)
		log.Println("[GetRequest] Failed with body ", string(bodyBytes))

	}

	return bodyBytes
}

// The ListFiles function retrieves a list of files from the OpenAI API using a bearer token.
func ListFiles(bearer string) *FileList {
	url := "https://api.openai.com/v1/files"

	bodyBytes := get_request(bearer, url)

	fileList := new(FileList)

	err := json.Unmarshal(bodyBytes, fileList)
	if err != nil {
		fmt.Println("error:", err)
		return nil
	}

	return fileList
}

// The RetrieveFile function retrieves a file from the OpenAI API using a bearer token and file ID.
func RetrieveFile(bearer string, fileId string) *File {
	url := "https://api.openai.com/v1/files/" + fileId

	bodyBytes := get_request(bearer, url)

	file := new(File)

	err := json.Unmarshal(bodyBytes, file)
	if err != nil {
		fmt.Println("error:", err)
		return nil
	}

	return file
}

// The function retrieves the content of a file from the OpenAI API using a bearer token and file ID.
func RetrieveFileContent(bearer string, fileId string) []byte {
	url := "https://api.openai.com/v1/files/" + fileId + "/content"

	bodyBytes := get_request(bearer, url)

	return bodyBytes
}

// The function uploads a file to the OpenAI API for fine-tuning.
func UploadFile(bearer string, filename string, fileContent string) *File {

	url := "https://api.openai.com/v1/files"

	var requestBody bytes.Buffer

	multiPartWriter := multipart.NewWriter(&requestBody)

	err := multiPartWriter.WriteField("purpose", "fine-tune")
	if err != nil {
		log.Fatal("[UploadFile] Failed to write purpose field", err)
	}

	fileWriter, err := multiPartWriter.CreateFormFile("file", filename)
	if err != nil {
		log.Fatal("[UploadFile] Failed to create form file", err)
	}

	_, err = io.Copy(fileWriter, strings.NewReader(fileContent))
	if err != nil {
		log.Fatal("[UploadFile] Failed to new reader", err)
	}

	multiPartWriter.Close()

	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		log.Fatal("[UploadFile] Failed to create request", err)
	}

	req.Header.Add("Authorization", "Bearer "+bearer)
	req.Header.Set("User-Agent", "curl/8.0.1")
	req.Header.Set("accept", "*/*")

	req.Header.Set("Content-Type", multiPartWriter.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("[UploadFile] Request error : ", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatal("[UploadFile] Failted, code status is", resp.Status)
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	file := new(File)

	err = json.Unmarshal(bodyBytes, file)
	if err != nil {
		fmt.Println("[UploadFile] error while Unmarshal :", err)
		return nil
	}

	return file

}

type DeletedFile struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// This function deletes a file from the OpenAI API using a bearer token and file ID.
func DeleteFile(bearer string, fileId string) *DeletedFile {
	url := "https://api.openai.com/v1/files/" + fileId

	var bodyBytes []byte
	success := false
	retry := 0
	for retry < 10 {
		client := &http.Client{}
		retry += 1

		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Add("Authorization", "Bearer "+bearer)
		req.Header.Set("User-Agent", "curl/8.0.1")
		req.Header.Set("accept", "*/*")

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode != http.StatusOK {
			time.Sleep(time.Second * 2)

			continue
		}

		// Lire le corps de la réponse
		defer resp.Body.Close()
		bodyBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		success = true
		break
	}

	if !success {
		log.Fatalf("Fail to retry")
	}

	deletedFile := new(DeletedFile)

	err := json.Unmarshal(bodyBytes, deletedFile)
	if err != nil {
		fmt.Println("error:", err)
		return nil
	}

	return deletedFile
}
