package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type ElectionData struct {
	Partyid            int    `json:"PartyID"`
	Stateid            int    `json:"StateID"`
	Candidatename      string `json:"CandidateName"`
	Gender             string `json:"Gender"`
	Age                int    `json:"Age"`
	Politicalpartyname string `json:"PoliticalPartyName"`
	Districtname       string `json:"DistrictName"`
	Localbodyname      string `json:"LocalBodyName"`
	Wardno             string `json:"WardNo"`
	Postname           string `json:"PostName"`
	Serialno           int    `json:"SerialNo"`
	Totalvotesrecieved int    `json:"TotalVotesRecieved"`
	Estatus            string `json:"EStatus"`
	Rank               int    `json:"Rank"`
}

type AggregateData struct {
	PartyName string `json:"PartyName"`
	PostName  string `json:"PostName"`
	WinCount  int    `json:"WinCount"`
}

func ReadAndParseData() ([]ElectionData, error) {
	jsonFile, err := os.Open("./local-level-election/raw/alldata.json")

	if err != nil {
		return nil, err
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var electionData []ElectionData

	err = json.Unmarshal(byteValue, &electionData)

	if err != nil {
		return nil, err
	}

	return electionData, nil
}

func convertJSONToCSV(electionData []ElectionData, destination string) error {
	outputFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	header := []string{"SerialNo", "PartyID", "StateID", "CandidateName", "Gender", "Age", "PoliticalPartyName", "DistrictName", "LocalBodyName", "WardNo", "PostName", "TotalVotesReceived", "EStatus", "Rank"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, e := range electionData {
		var csvRow []string
		csvRow = append(csvRow, strconv.Itoa(e.Serialno), strconv.Itoa(e.Partyid), strconv.Itoa(e.Stateid), e.Candidatename, e.Gender, strconv.Itoa(e.Age), e.Politicalpartyname, e.Districtname, e.Localbodyname, e.Wardno, e.Postname, strconv.Itoa(e.Totalvotesrecieved), e.Estatus, strconv.Itoa(e.Rank))
		if err := writer.Write(csvRow); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	electionData, err := ReadAndParseData()

	if err != nil {
		fmt.Println(err)
	}

	var mainMap = make(map[string][]ElectionData)
	var wardMap = make(map[string][]ElectionData)
	var winMap = make(map[string][]ElectionData)
	var winMapDistrict = make(map[string][]ElectionData)
	var winMapProvience = make(map[string][]ElectionData)
	var winMapCountry = make(map[string][]ElectionData)

	stateMap := map[string]string{
		"1": "प्रदेश नं. १",
		"2": "मधेश प्रदेश",
		"3": "वाग्मती प्रदेश",
		"4": "गण्डकी प्रदेश",
		"5": "लुम्बिनी प्रदेश",
		"6": "कर्णाली प्रदेश",
		"7": "सुदूरपश्चिम प्रदेश",
	}

	for _, data := range electionData {
		if strings.Contains(data.Politicalpartyname, "नेपाल कम्युनिष्ट पार्टी") {
			data.Politicalpartyname = strings.ReplaceAll(data.Politicalpartyname, "नेपाल कम्युनिष्ट पार्टी", "नेकपा")
		}
		if data.Politicalpartyname == "नेकपा (एकीकृत मार्क्सवादी-लेनिनवादी)" {
			data.Politicalpartyname = "नेकपा (एमाले)"
		}
		if data.Estatus == "E" {
			post := data.Postname
			if post == "दलित महिला सदस्य" || post == "महिला सदस्य" || post == "सदस्य" {
				post = "सदस्य"
			}
			winMapCountry[post] = append(winMapCountry[post], data)
			winMapProvience[fmt.Sprintf("%d__%s", data.Stateid, post)] = append(winMapProvience[fmt.Sprintf("%d__%s", data.Stateid, post)], data)
			winMapDistrict[fmt.Sprintf("%d__%s__%s", data.Stateid, data.Districtname, post)] = append(winMapDistrict[fmt.Sprintf("%d__%s__%s", data.Stateid, data.Districtname, post)], data)
			winMap[fmt.Sprintf("%d__%s__%s__%s", data.Stateid, data.Districtname, data.Localbodyname, post)] = append(winMap[fmt.Sprintf("%d__%s__%s__%s", data.Stateid, data.Districtname, data.Localbodyname, post)], data)
		}
		if data.Wardno != "" {
			wardMap[fmt.Sprintf("%d__%s__%s__%s__%s", data.Stateid, data.Districtname, data.Localbodyname, data.Postname, data.Wardno)] = append(wardMap[fmt.Sprintf("%d__%s__%s__%s__%s", data.Stateid, data.Districtname, data.Localbodyname, data.Postname, data.Wardno)], data)
		}
		mainMap[fmt.Sprintf("%d__%s__%s__%s", data.Stateid, data.Districtname, data.Localbodyname, data.Postname)] = append(mainMap[fmt.Sprintf("%d__%s__%s__%s", data.Stateid, data.Districtname, data.Localbodyname, data.Postname)], data)
	}

	var allCharts []*charts.Pie
	var localLevelChartMap = make(map[string][]*charts.Pie)
	var districtLevelChartMap = make(map[string][]*charts.Pie)
	var stateLevelChartMap = make(map[string][]*charts.Pie)
	var countryLevelChartMap = make(map[string][]*charts.Pie)

	for key, value := range winMapCountry {
		tC := make(map[string]int)
		for _, data := range value {
			tC[data.Politicalpartyname]++
		}
		var pie *charts.Pie
		pie = PieChartAgg(tC, key, true)
		allCharts = append(allCharts, pie)
		countryLevelChartMap["all"] = append(countryLevelChartMap["all"], pie)
	}

	for key, value := range winMapProvience {
		all := strings.Split(key, "__")
		tC := make(map[string]int)
		for _, data := range value {
			tC[data.Politicalpartyname]++
		}
		var pie *charts.Pie
		pie = PieChartAgg(tC, fmt.Sprintf("%s(%s)", stateMap[all[0]], all[1]), true)
		allCharts = append(allCharts, pie)
		stateLevelChartMap[fmt.Sprintf("%s", stateMap[all[0]])] = append(stateLevelChartMap[fmt.Sprintf("%s", stateMap[all[0]])], pie)
	}

	for key, value := range winMapDistrict {
		all := strings.Split(key, "__")
		tC := make(map[string]int)
		for _, data := range value {
			tC[data.Politicalpartyname]++
		}
		var pie *charts.Pie
		pie = PieChartAgg(tC, fmt.Sprintf("%s-  %s", all[1], all[2]), true)
		allCharts = append(allCharts, pie)
		districtLevelChartMap[fmt.Sprintf("%s__%s", stateMap[all[0]], all[1])] = append(districtLevelChartMap[fmt.Sprintf("%s__%s", stateMap[all[0]], all[1])], pie)
	}

	for key, value := range mainMap {
		all := strings.Split(key, "__")
		fileName := fmt.Sprintf("local-level-election/result/%s/%s/%s/", stateMap[all[0]], all[1], all[2])
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			err := os.MkdirAll(fileName, os.ModePerm)
			if err != nil {
				log.Println(err)
			}
		}
		var pie *charts.Pie
		pie = PieChart(value, fmt.Sprintf("%s(%s)-%s", all[2], all[1], all[3]), false)
		allCharts = append(allCharts, pie)
		if all[3] == "प्रमुख" || all[3] == "अध्यक्ष" {
			districtLevelChartMap[fmt.Sprintf("%s__%s", stateMap[all[0]], all[1])] = append(districtLevelChartMap[fmt.Sprintf("%s__%s", stateMap[all[0]], all[1])], pie)
		}
		if all[3] == "प्रमुख" || all[3] == "अध्यक्ष" || all[3] == "उपाध्यक्ष" || all[3] == "उपप्रमुख" {
			localLevelChartMap[fmt.Sprintf("%s__%s__%s", stateMap[all[0]], all[1], all[2])] = append(localLevelChartMap[fmt.Sprintf("%s__%s__%s", stateMap[all[0]], all[1], all[2])], pie)
		}
		csvFileName := fmt.Sprintf("%s%s.csv", fileName, all[3])
		os.Create(csvFileName)
		if err := convertJSONToCSV(value, csvFileName); err != nil {
			log.Fatal(err)
		}
	}

	for key, value := range winMap {
		all := strings.Split(key, "__")
		if all[3] == "वडा अध्यक्ष" || all[3] == "सदस्य" {
			tC := make(map[string]int)
			for _, data := range value {
				tC[data.Politicalpartyname]++
			}
			var pie *charts.Pie
			pie = PieChartAgg(tC, fmt.Sprintf("%s(%s)-%s", all[2], all[1], all[3]), true)
			allCharts = append(allCharts, pie)
			localLevelChartMap[fmt.Sprintf("%s__%s__%s", stateMap[all[0]], all[1], all[2])] = append(localLevelChartMap[fmt.Sprintf("%s__%s__%s", stateMap[all[0]], all[1], all[2])], pie)
		}
	}

	for key, value := range wardMap {
		all := strings.Split(key, "__")
		if all[3] == "वडा अध्यक्ष" {
			var pie *charts.Pie
			pie = PieChart(value, fmt.Sprintf("%s(%s)-%s(%s)", all[2], all[1], all[3], all[4]), false)
			allCharts = append(allCharts, pie)
			localLevelChartMap[fmt.Sprintf("%s__%s__%s", stateMap[all[0]], all[1], all[2])] = append(localLevelChartMap[fmt.Sprintf("%s__%s__%s", stateMap[all[0]], all[1], all[2])], pie)
		}
	}

	for key, value := range localLevelChartMap {
		all := strings.Split(key, "__")
		fileName := fmt.Sprintf("local-level-election/result/%s/%s/%s/", all[0], all[1], all[2])
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			err := os.MkdirAll(fileName, os.ModePerm)
			if err != nil {
				log.Println(err)
			}
		}
		CreateHTML(value, fileName)
	}

	for key, value := range districtLevelChartMap {
		all := strings.Split(key, "__")
		fileName := fmt.Sprintf("local-level-election/result/%s/%s/", all[0], all[1])
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			err := os.MkdirAll(fileName, os.ModePerm)
			if err != nil {
				log.Println(err)
			}
		}
		CreateHTML(value, fileName)
	}

	for key, value := range stateLevelChartMap {
		all := strings.Split(key, "__")
		fileName := fmt.Sprintf("local-level-election/result/%s/", all[0])
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			err := os.MkdirAll(fileName, os.ModePerm)
			if err != nil {
				log.Println(err)
			}
		}
		CreateHTML(value, fileName)
	}

	for _, value := range countryLevelChartMap {
		fileName := fmt.Sprintf("local-level-election/result/")
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			err := os.MkdirAll(fileName, os.ModePerm)
			if err != nil {
				log.Println(err)
			}
		}
		CreateHTML(value, fileName)
	}
}

func generatePieItemsAgg(sector map[string]int, isAggregated bool) []opts.PieData {
	items := make([]opts.PieData, 0)
	for k, v := range sector {
		items = append(items, opts.PieData{Value: v, Name: k})
	}
	return items
}

func PieChartAgg(aggregateDate map[string]int, title string, isAggregated bool) *charts.Pie {
	pie := charts.NewPie()
	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: title, Right: "20"}),
	)

	pie.AddSeries("pie", generatePieItemsAgg(aggregateDate, isAggregated)).
		SetSeriesOptions(charts.WithLabelOpts(
			opts.Label{
				Show:      true,
				Formatter: "{b}: {c}",
			}),
		)
	return pie
}

func PieChart(aggregateDate []ElectionData, title string, isAggregated bool) *charts.Pie {
	pie := charts.NewPie()
	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: title, Right: "20"}),
	)

	pie.AddSeries("pie", generatePieItems(aggregateDate, isAggregated)).
		SetSeriesOptions(charts.WithLabelOpts(
			opts.Label{
				Show:      true,
				Formatter: "{b}: {c}",
			}),
		)
	return pie
}

func generatePieItems(sector []ElectionData, isAggregated bool) []opts.PieData {
	items := make([]opts.PieData, 0)
	for _, v := range sector {
		var value int
		var name string
		name = fmt.Sprintf("%s- %s", v.Candidatename, v.Politicalpartyname)
		value = v.Totalvotesrecieved
		if isAggregated {
			name = v.Politicalpartyname
		}
		items = append(items, opts.PieData{Value: value, Name: name})
	}
	return items
}

func CreateHTML(pieChart []*charts.Pie, fileName string) {
	page := components.NewPage()
	for _, v := range pieChart {
		page.AddCharts(v)
	}
	f, err := os.Create(fmt.Sprintf("%s%s.html", fileName, "result"))
	if err != nil {
		log.Fatal(err)
	}
	page.Render(f)
}
