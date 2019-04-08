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
var configSearchAddr string
var configPageSize int
var configTotalSize int
var configScrollTime string
var configOutputFileName string
var configOutputDirectory string

const (
	timeFormatAMZLong   = "2006-01-02T15:04:05.000Z" // Reply date format with nanosecond precision.
	timeFormatLocalLong = "2006-01-02 15:04:05.000"  // Reply date format with nanosecond precision.
)

type Configuration struct {
	Configs []Config `json:"Configuration"`
}
type Config struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

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
	ScrollId string `json:"_scroll_id"`
	Hits     struct {
		Total int   `json:"total"`
		Hits  []Hit `json:"hits"`
	}
}

func main() {
	var result Result
	writeSize := 0
	hitsTotal := 0

	// config
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

	// create scroll
	resp, err := createScroll(configScrollTime, strconv.Itoa(configPageSize))
	if err != nil {
		return
	}

	for {
		// parse
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = json.Unmarshal(body, &result)
		if err != nil {
			fmt.Println(err)
			return
		}
		hitLen := len(result.Hits.Hits)
		if hitLen == 0 {
			fmt.Printf("dump(%d),configTotal(%d),stop!\n", writeSize, configTotalSize)
			return
		}
		hitsTotal = result.Hits.Total

		//write file
		for _, v := range result.Hits.Hits {
			t, _ := time.Parse(time.RFC3339, v.Log.Timestamp)
			str := fmt.Sprintf("%s %s %s %s %s\n", t.Local().Format(timeFormatLocalLong),
				v.Log.Host, v.Log.Module, v.Log.Level, v.Log.Message)
			f.WriteString(str)
			writeSize++
			if writeSize >= configTotalSize || writeSize >= hitsTotal {
				fmt.Printf("dump(%d) >= configTotal(%d) or hitsTotal(%d), stop!\n", writeSize, configTotalSize, hitsTotal)
				return
			}
		}

		// continue get scroll
		scrollId := result.ScrollId
		resp, err = getScroll(configScrollTime, scrollId)
		if err != nil {
			return
		}
	}
}

/*
curl '192.168.209.129:9200/logs/_search?pretty&scroll=1m' -H 'Content-Type: application/json' -d '{"size": 2}'
<- {
  "_scroll_id" : "DnF1......FvWV9n",
  "hits" : {
    "total" : 5,
    "max_score" : 1.0,
    "hits" : [ ]
  }
}
*/
func createScroll(scrollTime, pageSize string) (resp *http.Response, err error) {
	doc := fmt.Sprintf(`{"size":%s}`, pageSize)
	client := http.Client{}
	url := fmt.Sprintf("%s/logs/_search?scroll=%s", configSearchAddr, scrollTime)
	request, _ := http.NewRequest("GET", url, strings.NewReader(doc))
	request.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(request)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println(url)
	return
}

/*
curl '192.168.209.129:9200/_search/scroll?pretty' -H 'Content-Type: application/json' -d '
{"scroll":"1m","scroll_id": "DnF1......FvWV9n"}'
*/
func getScroll(scrollTime, scrollId string) (resp *http.Response, err error) {

	doc := fmt.Sprintf(`{"scroll":"%s","scroll_id":"%s"}`, scrollTime, scrollId)
	client := http.Client{}
	url := fmt.Sprintf("%s/_search/scroll", configSearchAddr)
	request, _ := http.NewRequest("GET", url, strings.NewReader(doc))
	request.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(request)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println(url)
	return
}

/*
config.json
{
"Configuration": [
        {"Key": "SearchURL","Value":"http://192.168.209.129:9200"},
        {"Key": "TotalSize","Value": "10000"},
        {"Key": "PageSize","Value": "2"},
        {"Key": "ScrollTime","Value": "1m"},
        {"Key": "OutputFileName","Value":"logs-output.out"},
        {"Key": "OutputDirectory","Value":""}
    ]
}
*/
func getConfig() (err error) {
	// init the config array
	var configs Configuration

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

	// unmarshal the byteArray into the configuration object
	err = json.Unmarshal(byteValue, &configs)
	if err != nil {
		return
	}

	// loop through the configuration and find the keys that we need to use and assign them to the global variables to be used
	for i := 0; i < len(configs.Configs); i++ {
		var iter = configs.Configs[i]

		switch strings.ToLower(iter.Key) {
		case "searchaddr":
			configSearchAddr = iter.Value
		case "outputfilename":
			configOutputFileName = iter.Value
		case "outputdirectory":
			configOutputDirectory = iter.Value
		case "scrolltime":
			configScrollTime = iter.Value
		case "pagesize":
			temp, convErr := strconv.Atoi(iter.Value)
			if convErr != nil {
				fmt.Println("PageSize invalid, expected an integer, actual value: " + iter.Value)
			} else {
				configPageSize = temp
			}
		case "totalsize":
			temp, convErr := strconv.Atoi(iter.Value)
			if convErr != nil {
				fmt.Println("TotalSize invalid, expected an integer, actual value: " + iter.Value)
			} else {
				configTotalSize = temp
			}
		default:
			fmt.Println(iter.Key + " is not a tracked key")
		}
	}
	return nil
}
