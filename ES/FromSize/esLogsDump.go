package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// from config.json
var configSearchURL string
var configSize int
var configOutputFileName string
var configOutputDirectory string

const (
	timeFormatLocalLong = "2006-01-02 15:04:05.000"
)

type Configuration struct {
	Configs []Config `json:"Configuration"`
}
type Config struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// from es
var globalTotalHits int
var globalFromPos int

type Log struct {
	Timestamp string `json:"@timestamp"`
	Host      string `json:"host"`
	Module    string `json:"module"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}
type Hit struct {
	Log Log `json:"_source"`
}
type Result struct {
	Hits struct {
		Total int   `json:"total"`
		Hits  []Hit `json:"hits"`
	}
}

func main() {

	err := getConfig()
	if err != nil {
		fmt.Println(configOutputFileName, err)
		return
	}

	f, err := os.Create(configOutputFileName)
	if err != nil {
		fmt.Println(configOutputFileName, err)
		return
	}
	defer f.Close()

	fileCounter := 0
	for globalFromPos := 0; globalFromPos <= globalTotalHits; globalFromPos += configSize {
		// get from db
		resp, err := getLogData(strconv.Itoa(globalFromPos), strconv.Itoa(configSize))
		if err != nil {
			return
		}
		defer resp.Body.Close()

		// parse
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		var result Result
		err = json.Unmarshal(body, &result)
		if err != nil {
			fmt.Println(err)
		}

		//write file
		for _, v := range result.Hits.Hits {
			t, _ := time.Parse(time.RFC3339, v.Log.Timestamp)
			str := fmt.Sprintf("%s %s %s %s %s\n", t.Local().Format(timeFormatLocalLong),
				v.Log.Host, v.Log.Module, v.Log.Level, v.Log.Message)
			f.WriteString(str)
		}

		// next loop
		globalTotalHits = result.Hits.Total
		if globalTotalHits > 10000 {
			globalTotalHits = 10000
		}
		fileCounter++
	}
}

func getLogData(from, size string) (resp *http.Response, err error) {

	req := configSearchURL + "/logs/_search?pretty"
	if from != "" {
		req = req + "&from=" + from
	}
	if size != "" {
		req = req + "&size=" + size
	}

	fmt.Println(req)

	resp, err = http.Get(req)
	if err != nil {
		fmt.Println(err)
	}

	return
}

// This function will open up the Settings.json file and load the values that this program needs out of it.
//  it ignores keys that are not used and writes a message of those keys
func getConfig() (err error) {
	jsonFile, err := os.Open("config.json")
	if err != nil {
		return
	}
	defer jsonFile.Close()

	// read all the settings
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return
	}
	// init the config array
	var configs Configuration

	// unmarshal the byteArray into the configuration object
	err = json.Unmarshal(byteValue, &configs)
	if err != nil {
		return
	}

	// loop through the configuration and find the keys that we need to use and assign them to the global variables to be used
	for i := 0; i < len(configs.Configs); i++ {
		var iter = configs.Configs[i]

		switch strings.ToLower(iter.Key) {
		case "searchurl":
			configSearchURL = iter.Value
		case "outputfilename":
			configOutputFileName = iter.Value
		case "outputdirectory":
			configOutputDirectory = iter.Value
		case "size":
			temp, convErr := strconv.Atoi(iter.Value)
			if convErr != nil {
				fmt.Println("Size invalid, expected an integer, actual value: " + iter.Value)
			} else {
				configSize = temp
			}
		default:
			fmt.Println(iter.Key + " is not a tracked key")
		}
	}
	return nil
}
