package main

import (
	"fmt"

	"github.com/ragadeeshu/octo-canyon-freshness-league/datahandling"
)

func main() {

	// results, err := datahandling.GetOrFetchData()
	_, err := datahandling.GetOrFetchData()
	if err != nil {
		fmt.Println(err)
	}

	// output, _ := json.MarshalIndent(results, "", "\t")
	// fmt.Println(string(output))

}
