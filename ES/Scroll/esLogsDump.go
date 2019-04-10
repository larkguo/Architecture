package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// from config.json
var configSearchAddr string
var configPageSize int
var configTotalSize int
var configStartTime string
var configEndTime string
var configLevel string
var configScrollTime string
var configOutputFile string

const (
	timeFormatAMZLong   = "2006-01-02T15:04:05.000Z" // Reply date format with nanosecond precision.
	timeFormatAMZ       = "2006-01-02T15:04:05Z"     // Reply date format with nanosecond precision.
	timeFormatLocalLong = "2006-01-02 15:04:05.000"  // Reply date format with nanosecond precision.
	timeFormatLocal     = "2006-01-02 15:04:05"      // Reply date format with nanosecond precision.
)

type configuration struct {
	Configs []config `json:"Configuration"`
}
type config struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

type log struct {
	Timestamp string `json:"@timestamp"`
	Host      string `json:"host"`
	Module    string `json:"module"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}
type hit struct {
	Log log `json:"_source"`
}
type result struct {
	ScrollID string `json:"_scroll_id"`
	Hits     struct {
		Total int   `json:"total"`
		Hits  []hit `json:"hits"`
	}
}

func main() {
	var result result
	writeSize := 0
	hitsTotal := 0

	// config
	err := getConfig()
	if err != nil {
		fmt.Println(configOutputFile, err)
		return
	}

	path := filepath.Dir(configOutputFile)
	os.MkdirAll(path, os.ModePerm)
	f, err := os.Create(configOutputFile)
	if err != nil {
		fmt.Println(configOutputFile, err)
		return
	}
	defer f.Close()

	// create scroll
	resp, err := createScroll(configScrollTime, strconv.Itoa(configPageSize), configStartTime, configEndTime, configLevel)
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
			fmt.Printf("dump(%d),configTotal(%d),stop!\r\n", writeSize, configTotalSize)
			return
		}
		hitsTotal = result.Hits.Total

		//write file
		for _, v := range result.Hits.Hits {
			t, _ := time.Parse(time.RFC3339, v.Log.Timestamp)
			str := fmt.Sprintf("%s %s %s %s %s\r\n", t.Local().Format(timeFormatLocalLong),
				v.Log.Host, v.Log.Module, v.Log.Level, v.Log.Message)
			f.WriteString(str)
			writeSize++
			if writeSize >= configTotalSize || writeSize >= hitsTotal {
				fmt.Printf("dump(%d) >= configTotal(%d) or hitsTotal(%d), stop!\r\n", writeSize, configTotalSize, hitsTotal)
				return
			}
		}

		// continue get scroll
		scrollID := result.ScrollID
		resp, err = getScroll(configScrollTime, scrollID)
		if err != nil {
			return
		}
	}
}

/*
curl '192.168.209.129:9200/logs/_search?pretty&scroll=1m' -H 'Content-Type: application/json' -d '
{
    "size": 2,
    "query": {"bool": {"must": [
        {"range": {"@timestamp": { "gte": "2017-12-25T01:00:00.000Z","lte": "2019-12-25T02:10:00.000Z"}}},
        {"match": {"level": "WARN ERROR FATAL"}}
    ]}}
}'
<- {
  "_scroll_id" : "DnF1......FvWV9n",
  "hits" : {
    "total" : 5,
    "max_score" : 1.0,
    "hits" : [ ]
  }
}
*/
func createScroll(scrollTime, pageSize, startTime, endTime, level string) (resp *http.Response, err error) {
	var sizeStr, levelStr, timeStr, doc string
	sizeStr = fmt.Sprintf(`"size":%s`, pageSize)
	if level != "" {
		levelStr = fmt.Sprintf(`{"match": {"level": "%s"}}`, level)
	}
	if startTime != "" {
		if endTime != "" {
			timeStr = fmt.Sprintf(`{"range": {"@timestamp": { "gte": "%s","lte": "%s"}}}`, startTime, endTime)
		} else {
			timeStr = fmt.Sprintf(`{"range": {"@timestamp": { "gte": "%s"}}}`, startTime)
		}
	} else { // startTime == ""
		if endTime != "" {
			timeStr = fmt.Sprintf(`{"range": {"@timestamp": { "lte": "%s"}}}`, endTime)
		}
	}

	if levelStr != "" {
		if timeStr != "" {
			doc = fmt.Sprintf(`{%s,"query": {"bool": {"must": [%s,%s]}}}`, sizeStr, timeStr, levelStr)
		} else {
			doc = fmt.Sprintf(`{%s,"query": {"bool": {"must": [%s]}}}`, sizeStr, levelStr)
		}
	} else { // levelStr ==""
		if timeStr != "" {
			doc = fmt.Sprintf(`{%s,"query": {"bool": {"must": [%s]}}}`, sizeStr, timeStr)
		} else {
			doc = fmt.Sprintf(`{%s}`, sizeStr)
		}
	}

	fmt.Println(doc)

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
func getScroll(scrollTime, scrollID string) (resp *http.Response, err error) {

	doc := fmt.Sprintf(`{"scroll":"%s","scroll_id":"%s"}`, scrollTime, scrollID)
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
        {"Key": "SearchAddr","Value":"http://192.168.209.129:9200"},
        {"Key": "TotalSize","Value": "10000"},
        {"Key": "PageSize","Value": "2"},
				{"Key": "ScrollTime","Value": "1m"},
        {"Key": "StartTime","Value": "2017-12-25 01:00:00.000 "},
        {"Key": "EndTime","Value": "2022-12-25 01:00:00.000"},
        {"Key": "Level","Value": "WARN ERROR FATAL"},
        {"Key": "OutputFile","Value":"G:\Architecture\ES\Scroll\dump.log"}
    ]
}
*/
func getConfig() (err error) {
	// init the config array
	var configs configuration

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
		key := strings.TrimSpace(iter.Key)
		value := strings.TrimSpace(iter.Value)

		switch strings.ToLower(key) {
		case "searchaddr":
			configSearchAddr = value
		case "outputfile":
			configOutputFile = value
		case "scrolltime":
			configScrollTime = value
		case "starttime":
			configStartTime = value
			t, e := time.Parse(timeFormatLocal, configStartTime)
			if e != nil {
				t, e = time.Parse(timeFormatLocalLong, configStartTime)
				if e == nil {
					configStartTime = t.Format(timeFormatAMZLong)
				}
			} else {
				configStartTime = t.Format(timeFormatAMZLong)
			}
		case "endtime":
			configEndTime = value
			t, e := time.Parse(timeFormatLocal, configEndTime)
			if e != nil {
				t, e = time.Parse(timeFormatLocalLong, configEndTime)
				if e == nil {
					configEndTime = t.Format(timeFormatAMZLong)
				}
			} else {
				configEndTime = t.Format(timeFormatAMZLong)
			}
		case "level":
			configLevel = value
		case "pagesize":
			temp, convErr := strconv.Atoi(value)
			if convErr != nil {
				fmt.Println("PageSize invalid, expected an integer, actual value: " + iter.Value)
			} else {
				configPageSize = temp
			}
		case "totalsize":
			temp, convErr := strconv.Atoi(value)
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
