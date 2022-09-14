package main

import (
	"os"
	"time"
)

func saveToDisk(filename string, stringData string) error {
	data := []byte(stringData)
	err := os.WriteFile(filename, data, 0644)
	return err
}

func loadFileIfRecent(filename string) string {
	// check when the file was last modified
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return ""
	}
	// if the file was modified less than 10 minutes ago, load it
	now := time.Now()
	fileModTimePlusTen := fileInfo.ModTime().Add(10 * time.Minute)
	if now.After(fileModTimePlusTen) {
		return ""
	}
	return readFile(filename)
}

func readFile(filename string) string {
	data, err := os.ReadFile(filename)
	if err != nil {
		return ""
	}
	return string(data)
}

func ValueOrDefault[T comparable](value T, defaultValue T) T {
	if value == *new(T) {
		return defaultValue
	}
	return value
}
