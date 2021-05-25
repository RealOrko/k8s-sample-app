package formatters

import (
	"encoding/json"
	"log"
)

// *** JSON Formatter ***

func ToJSON(i interface{}) string {
	res, err := json.Marshal(i)
	if err != nil {
		log.Fatal("Failed to convert to json ... ")
		panic("Exiting ... ")
	}
	return string(res)
}