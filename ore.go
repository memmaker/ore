package main

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
)

var endPoint = ValueOrDefault(os.Getenv("CRUDE_API_ENDPOINT"), "https://proxyman.local:8080")
var apiKey = ValueOrDefault(os.Getenv("CRUDE_API_KEY"), "crude_api_key")

func main() {
	if len(os.Args) > 1 {
		// save, read, browse, deleteEntry
		if os.Args[1] == "save" {
			modelName := os.Args[2]
			response := streamEntriesFromStdin(modelName)
			fmt.Println(response)
		} else if os.Args[1] == "save-models" {
			fmt.Println(saveModelsFromStdin())
		} else if os.Args[1] == "read" {
			compoundID := os.Args[2]
			response := readEntry(compoundID)
			fmt.Println(response)
		} else if os.Args[1] == "browse" {
			query := os.Args[2]
			response := browse(query)
			fmt.Println(response)
		} else if os.Args[1] == "models" {
			modelName := ""
			if len(os.Args) > 2 {
				modelName = os.Args[2]
			}
			modelInfos := getModels(modelName)
			fmt.Println(modelInfos)
		} else if os.Args[1] == "relation" {
			relationExpression := os.Args[2]
			response := getRelation(relationExpression)
			fmt.Println(response)
		} else if os.Args[1] == "fields" {
			response := getFieldTypes()
			fmt.Println(response)
		} else if os.Args[1] == "delete" {
			compoundID := os.Args[2]
			if compoundID == "-" {
				deleteFromStdin()
				os.Exit(0)
			} else if strings.Contains(compoundID, "/") {
				response := deleteEntry(compoundID)
				fmt.Println(response)
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
	fmt.Println("  delete:  " + os.Args[0] + " deleteEntry <compoundID>")

}

func readEntry(compoundID string) string {
	format := "tsv"
	if strings.Contains(compoundID, ":") {
		parts := strings.Split(compoundID, ":")
		format = parts[0]
		compoundID = parts[1]
	}
	response := get(endPoint + "/entries/read/" + compoundID + "?format=" + format)
	return response.Body
}

func getRelation(expression string) string {
	parts := strings.Split(expression, ".")
	modelName := parts[0]
	fieldName := parts[1]
	response := get(endPoint + "/z/relation/" + modelName + "/" + fieldName)
	return response.Body
}

func saveModelsFromStdin() string {
	data, _ := io.ReadAll(os.Stdin)
	response := post(endPoint+"/z/models", string(data))
	return response.Body
}

func streamEntriesFromStdin(modelName string) string {
	fmt.Println("Streaming data from stdin to " + modelName)
	r := io.ReadCloser(os.Stdin)
	defer r.Close()
	result := postStream(endPoint+"/z/save/"+modelName, r)
	return result.Body
}

func getFieldTypes() string {
	response := queryRequest(apiKey, endPoint+"/models/fields")
	return response.Body
}

func getModels(modelName string) string {
	endpointURL := endPoint + "/z/models"
	if modelName != "" {
		endpointURL += "/" + modelName
	}
	apiResponse := queryRequest(apiKey, endpointURL)
	return apiResponse.Body
}

func browse(query string) string {
	parts := strings.Split(query, ":")
	modelName := parts[0]
	query = strings.Replace(query, modelName+":", "", 1)
	queryArg := "q=" + url.QueryEscape(query)
	formatArg := "format=tsv"
	response := get(endPoint + "/entries/read/" + modelName + "?" + formatArg + "&" + queryArg)
	return response.Body
}

func deleteEntry(compoundID string) string {
	response := httpDelete(endPoint + "/entries/" + compoundID)
	return response.Body
}

func deleteFromStdin() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "/") {
			fmt.Println(deleteEntry(line))
		}
	}
}
