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

	header := []string{"PartyID", "StateID", "CandidateName", "Gender", "Age", "PoliticalPartyName", "DistrictName", "LocalBodyName", "WardNo", "PostName", "SerialNo", "TotalVotesReceived", "EStatus", "Rank"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, e := range electionData {
		var csvRow []string
		csvRow = append(csvRow, strconv.Itoa(e.Partyid), strconv.Itoa(e.Stateid), e.Candidatename, e.Gender, strconv.Itoa(e.Age), e.Politicalpartyname, e.Districtname, e.Localbodyname, e.Wardno, e.Postname, strconv.Itoa(e.Serialno), strconv.Itoa(e.Totalvotesrecieved), e.Estatus, strconv.Itoa(e.Rank))
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

	for _, data := range electionData {
		mainMap[fmt.Sprintf("%d-%s-%s-%s", data.Stateid, data.Districtname, data.Localbodyname, data.Postname)] = append(mainMap[fmt.Sprintf("%d-%s-%s-%s", data.Stateid, data.Districtname, data.Localbodyname, data.Postname)], data)
	}

	for key, value := range mainMap {
		all := strings.Split(key, "-")
		fileName := fmt.Sprintf("local-level-election/result/%s/%s/%s/", all[0], all[1], all[2])
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			err := os.MkdirAll(fileName, os.ModePerm)
			if err != nil {
				log.Println(err)
			}
		}
		csvFileName := fmt.Sprintf("%s%s.csv", fileName, all[3])
		os.Create(csvFileName)
		if err := convertJSONToCSV(value, csvFileName); err != nil {
			log.Fatal(err)
		}
		fmt.Println("fileName", fileName)
	}
}
