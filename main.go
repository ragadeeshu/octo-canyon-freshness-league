package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	league, _ := LoadLeague()
	// fmt.Println(league)
	err := LoadSplatnetData(&league)
	if err != nil {
		fmt.Println(err)
	}
	output, _ := json.MarshalIndent(league, "", "\t")
	fmt.Println(string(output))

}
