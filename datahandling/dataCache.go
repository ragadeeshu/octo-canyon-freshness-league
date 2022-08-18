package datahandling

import (
	"encoding/json"
	"io/ioutil"
	"sync"
	"time"
)

func GetOrFetchData(iksmMutex *sync.Mutex) (Results, error) {
	results, err := getResults()
	if err != nil || time.Now().Sub(results.Date) > 5*time.Minute {
		var league League

		league, err = getLeague(iksmMutex)
		if err != nil {
			return results, err
		}
		results = CalculateResults(league)
		err = saveResults(results)
		if err != nil {
			return results, err
		}
	}
	return results, err
}

func FetchContestant(iksmMutex *sync.Mutex, index int) (int, error) {
	index, err := getSingleContestant(iksmMutex, index)
	return index, err
}

func getResults() (results Results, err error) {
	byteValue, err := ioutil.ReadFile("results.json")
	if err != nil {
		return
	}
	err = json.Unmarshal(byteValue, &results)
	return
}

func saveResults(results Results) (err error) {
	output, err := json.MarshalIndent(results, "", "\t")
	if err != nil {
		return
	}
	err = ioutil.WriteFile("results.json", output, 0644)
	return
}
