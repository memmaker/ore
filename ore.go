package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

var endPoint = ValueOrDefault(os.Getenv("CRUDE_API_ENDPOINT"), "http://localhost:8080")
var apiKey = ValueOrDefault(os.Getenv("CRUDE_API_KEY"), "crude_api_key")
var workdir = path.Join(getHome(), ".crude")
var modelFile = path.Join(workdir, "model_cache.json")
var etagDir = path.Join(workdir, "etag_cache")

func ensureDirExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, os.ModePerm)
	}
}

func getHome() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return homeDir
}

var modelNames []string
var models map[string]Model

func getFieldTypes() string {
	response := queryRequest(apiKey, endPoint+"/models/fields")
	return response.Body
}

func loadModelsFromAPI() string {
	response := queryRequest(apiKey, endPoint+"/models/browse")
	err := saveToDisk(modelFile, response.Body)
	if err != nil {
		log.Fatal(err)
	}
	return response.Body
}

func loadModelsFromString(jsonString string) {
	var response ModelResponse
	json.Unmarshal([]byte(jsonString), &response)
	modelNames = response.ModelNames
	models = response.Models
}

func ensureModels() {
	diskModelString := loadFileIfRecent(modelFile)
	if diskModelString == "" {
		fmt.Println("Re-loading models...")
		diskModelString = loadModelsFromAPI()
	}
	loadModelsFromString(diskModelString)
	if len(modelNames) == 0 {
		log.Fatal("No models found")
	}
}

func browse(modelName string, page int) string {
	response := get(endPoint + "/entries/" + modelName + "?page=" + fmt.Sprintf("%d", page))
	return response.Body
}

func read(compoundID string) string {
	response := get(endPoint + "/entries/" + compoundID)
	writeETag(compoundID, response.ETag)
	return response.Body
}

func writeETag(id string, tag string) {
	fileName := path.Join(etagDir, id)
	// get the directory name of filename
	dirName := path.Dir(fileName)
	ensureDirExists(dirName)
	err := saveToDisk(fileName, tag)
	if err != nil {
		log.Fatal("Writing etag: " + err.Error())
	}
	fmt.Println("Writing etag " + tag + " to " + fileName)
}

func loadETag(id string) string {
	fileName := path.Join(etagDir, id)
	etag := readFile(fileName)
	fmt.Println("Loading etag for " + id + " etag: " + etag)
	return etag
}

func delete(compoundID string) string {
	response := httpDelete(endPoint + "/entries/" + compoundID)
	return response.Body
}

func save(identifier string, data JsonObject) string {
	isUpdate := strings.Contains(identifier, "/")
	var response ResourceResult
	if isUpdate {
		etag := loadETag(identifier)
		response = postUpdate(endPoint+"/entries/"+identifier, toJson(data), etag)
	} else {
		response = post(endPoint+"/entries/"+identifier, toJson(data))
	}

	responseObject := parseJson(response.Body)
	dataPart := responseObject["data"].(map[string]interface{})
	entryId := dataPart["id"].(string)
	if entryId == "" {
		return ""
	}
	// check if identifier contains a slash somewhere

	if isUpdate {
		writeETag(identifier, response.ETag)
		return identifier
	}
	cid := identifier + "/" + entryId
	writeETag(cid, response.ETag)
	fmt.Println("Saving of " + cid + " returned " + response.ETag)
	return cid
}

func main() {
	ensureDirExists(workdir)
	ensureDirExists(etagDir)
	ensureModels()

	//createTestData()
	//fmt.Println(modelNames)

	// check for command line arguments
	if len(os.Args) > 1 {
		// save, read, browse, delete
		if os.Args[1] == "save" {
			identifier := os.Args[2]
			jsonString := os.Args[3]

			data := parseJson(jsonString)
			id := save(identifier, data)
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
			for _, modelName := range modelNames {
				listOfModels = append(listOfModels, models[modelName])
			}
			jsonStringOfModels := toJson(listOfModels)
			pprint(jsonStringOfModels)
		} else if os.Args[1] == "fields" {
			response := getFieldTypes()
			pprint(response)
		} else if os.Args[1] == "delete" {
			compoundID := os.Args[2]
			response := delete(compoundID)
			fmt.Println(response)
			pprint(response)
		}
	}
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
