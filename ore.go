package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var endPoint = ValueOrDefault(os.Getenv("CRUDE_API_ENDPOINT"), "https://proxyman.local:8080")
var apiKey = ValueOrDefault(os.Getenv("CRUDE_API_KEY"), "crude_api_key")

func ensureDirExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, os.ModePerm)
	}
}

func getFieldTypes() string {
	response := queryRequest(apiKey, endPoint+"/models/fields")
	return response.Body
}

func loadModelsFromAPI() ModelInfos {
	apiResponse := queryRequest(apiKey, endPoint+"/models/browse")
	var response ModelResponse
	json.Unmarshal([]byte(apiResponse.Body), &response)
	return ModelInfos{Models: response.Models, ModelNames: response.ModelNames}
}

type ModelInfos struct {
	Models     map[string]Model
	ModelNames []string
}

func browse(modelName string, page int) string {
	response := get(endPoint + "/entries/read/" + modelName + "?page=" + fmt.Sprintf("%d", page))
	return response.Body
}

func read(compoundID string) string {
	response := get(endPoint + "/entries/read/" + compoundID)
	return response.Body
}

func delete(compoundID string) string {
	response := httpDelete(endPoint + "/entries/" + compoundID)
	return response.Body
}

func save(compoundIDOrModel string, data JsonObject) string {
	var etag string
	var ok bool
	if etag, ok = data["_etag"].(string); ok {
	} else {
		etag = ""
	}
	var response ResourceResult
	if etag != "" {
		response = postUpdate(endPoint+"/entries/"+compoundIDOrModel, toJson(data), etag)
	} else {
		response = post(endPoint+"/entries/"+compoundIDOrModel, toJson(data))
	}

	responseObject := parseJson(response.Body)
	dataPart := responseObject["data"].(map[string]interface{})
	entryId := dataPart["id"].(string)
	if entryId == "" {
		return ""
	}
	// check if compoundIDOrModel contains a slash somewhere
	var cid string
	if strings.Contains(compoundIDOrModel, "/") {
		cid = compoundIDOrModel
	} else {
		cid = compoundIDOrModel + "/" + entryId
	}

	fmt.Println("Saving of " + cid + " returned " + response.ETag)
	return cid
}

func main() {
	if len(os.Args) > 1 {
		// save, read, browse, delete
		if os.Args[1] == "save" {
			compoundIDOrModel := os.Args[2]
			jsonString := os.Args[3]
			if jsonString == "-" {
				streamCSV(compoundIDOrModel)
				os.Exit(0)
			}
			data := parseJson(jsonString)
			id := save(compoundIDOrModel, data)
			if id == "" {
				os.Exit(1)
			}
			fmt.Println(id)
			os.Exit(0)
		} else if os.Args[1] == "read" {
			compoundID := os.Args[2]
			response := read(compoundID)
			pprint(response)
		} else if os.Args[1] == "browse" {
			modelName := os.Args[2]
			page := 1
			if len(os.Args) > 3 {
				page, _ = strconv.Atoi(os.Args[3])
			}
			response := browse(modelName, page)
			pprint(response)
		} else if os.Args[1] == "models" {
			var listOfModels []Model
			modelInfos := loadModelsFromAPI()
			for _, modelName := range modelInfos.ModelNames {
				listOfModels = append(listOfModels, modelInfos.Models[modelName])
			}
			jsonStringOfModels := toJson(listOfModels)
			pprint(jsonStringOfModels)
		} else if os.Args[1] == "fields" {
			response := getFieldTypes()
			pprint(response)
		} else if os.Args[1] == "delete" {
			compoundID := os.Args[2]
			if compoundID == "-" {
				streamDelete()
				os.Exit(0)
			} else if strings.Contains(compoundID, "/") {
				response := delete(compoundID)
				pprint(response)
				os.Exit(0)
			}
			os.Exit(1)
		}
	}
	fmt.Println("Usage: ")
	fmt.Println("  models:  " + os.Args[0] + " models")
	fmt.Println("  fields:  " + os.Args[0] + " fields")
	fmt.Println("  browse:  " + os.Args[0] + " browse <model> [page]")
	fmt.Println("  read:    " + os.Args[0] + " read <compoundID>")
	fmt.Println("  save:    " + os.Args[0] + " save <model> <json>")
	fmt.Println("  delete:  " + os.Args[0] + " delete <compoundID>")

}

func streamDelete() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "/") {
			pprint(delete(line))
		}
	}
}

func streamCSV(compoundIDOrModel string) {
	fmt.Println("Streaming CSV from stdin to " + compoundIDOrModel)
	r := io.ReadCloser(os.Stdin)
	defer r.Close()
	result := postStream(endPoint+"/imports/"+compoundIDOrModel, r)
	fmt.Println(result)
}

func pprint(jsonString string) {
	fmt.Println(jsonPrettify(jsonString))
}

func createTestData() {
	for i := 0; i < 10000; i++ {
		newId := save("topics", JsonObject{
			"name":   "Da new shite" + fmt.Sprintf("%d", i),
			"status": "open",
		})
		fmt.Println(newId)
	}
}
