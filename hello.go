package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	ht "github.com/v-grabko1999/go-html2json"
	"golang.org/x/net/html/atom"
)

var baseUrl = "https://gist.githubusercontent.com"

type Element2 struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

type Element struct {
	Name     string `json:"name"`
	Elements []Element2
}

type Table struct {
	Name     string `json:"name"`
	Elements []Element
}

type History struct {
	Description string
	CreateAt    string
}

type Receiver struct {
	ReceivedBy string
	Histories  []History
}

func main() {

	spaceClient := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}

	req, err := http.NewRequest(http.MethodGet, baseUrl+"/nubors/eecf5b8dc838d4e6cc9de9f7b5db236f/raw/d34e1823906d3ab36ccc2e687fcafedf3eacfac9/jne-awb.html", nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "spacecount-tutorial")
	req.Header.Set("Content-Type", "application/json")

	res, getErr := spaceClient.Do(req)

	if getErr != nil {
		log.Fatal(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	d, err := ht.New(strings.NewReader(string(body)))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		description, err := parseHistory(d)
		if err == nil {
			w.Write([]byte(description))
		} else {
			w.Write([]byte("Error"))
		}
	})

	http.ListenAndServe(":7000", nil)
}

func parseHistory(d *ht.Dom) (string, error) {
	history := []interface{}{}

	response := map[string]interface{}{
		"status": map[string]interface{}{
			"code":    "060102",
			"message": "Delivery tracking detail invalid",
		},
	}

	var jsonRes, err1 = json.Marshal(response)
	if err1 != nil {
		return "", err1
	}

	table_contents, err2 := d.ByClass("table_style")
	if err2 != nil {
		return string(jsonRes), err2
	}

	tbody, err3 := table_contents[2].ByTag(atom.Tbody)
	if err3 != nil {
		return string(jsonRes), err3
	}
	tbody[0].ToNode()

	var jsonData, err4 = json.Marshal(tbody[0].ToNode())
	if err4 != nil {
		return string(jsonRes), err4
	}

	var data Table

	var err5 = json.Unmarshal(jsonData, &data)
	if err5 != nil {
		return "", err5
	}

	Description := ""

	for i := 0; i < len(data.Elements); i++ {
		history = append(history, History{Description: data.Elements[i].Elements[1].Text, CreateAt: data.Elements[i].Elements[0].Text})

		Description = data.Elements[i].Elements[1].Text
		match, err := regexp.MatchString(`DELIVERED TO`, Description)
		if match {
			out := strings.TrimLeft(strings.TrimRight(Description, "]"), "DELIVERED TO [")
			s := strings.Split(out, "|")
			Description = s[0]
			fmt.Println(err)
		}
	}

	data1 := map[string]interface{}{
		"status": map[string]interface{}{
			"code":    "060101",
			"message": "Delivery tracking detail fetched successfully",
		},
		"data": map[string]interface{}{
			"receivedBy": Description,
			"histories":  history,
		},
	}

	var jsonData2, err6 = json.Marshal(data1)
	if err6 != nil {
		return "", err6
	}

	return string(jsonData2), nil
}
