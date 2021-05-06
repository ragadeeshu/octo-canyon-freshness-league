package datahandling

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"sync"
	"time"
)

var (
	statsURL       = "https://app.splatoon2.nintendo.net/api/records/hero"
	resultsURL     = "https://app.splatoon2.nintendo.net/api/results"
	nameAndIconURL = "https://app.splatoon2.nintendo.net/api/nickname_and_icon"
	useragent      = "Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_3 like Mac OS X) AppleWebKit/603.3.8 (KHTML, like Gecko) Mobile/14G60"
)

type League struct {
	LeagueName  string       `json:"league"`
	Contestants []Contestant `json:"contestants"`
}

type Contestant struct {
	Name         string       `json:"name"`
	Cookie       string       `json:"cookie,omitempty"`
	SessionToken string       `json:"session_token,omitempty"`
	SplatnetName string       `json:"splatnet_name"`
	ProxyURL     string       `json:"proxy,omitempty"`
	PictureURL   string       `json:"picture_url"`
	SplatnetData SplatnetData `json:"-"`
}

type ProxyResponse struct {
	Name         string       `json:"name"`
	Cookie       string       `json:"cookie"`
	SessionToken string       `json:"session_token"`
	SplatnetName string       `json:"splatnet_name"`
	PictureURL   string       `json:"picture_url"`
	Date         time.Time    `json:"time"`
	SplatnetData SplatnetData `json:"splatnet_Data"`
}

//Stuff for getting picture and name

type BattleResults struct {
	BattleResults [50]BattleResult `json:"results,omitempty"`
}

type BattleResult struct {
	PlayerResult PlayerBattleResult `json:"player_result,omitempty"`
}

type PlayerBattleResult struct {
	Player Player `json:"player,omitempty"`
}

type Player struct {
	PrincipalID string `json:"principal_id,omitempty"`
}

type SplatnetProfiles struct {
	SplatnetProfiles []SplatnetProfile `json:"nickname_and_icons"`
}

type SplatnetProfile struct {
	Name       string `json:"nickname"`
	PictureURL string `json:"thumbnail_url"`
}

//Stuff for getting times
type SplatnetData struct {
	SplatnetCampaignSummary SplatnetCampaignSummary  `json:"summary"`
	SplatnetStageClearDatas []SplatnetStageClearData `json:"stage_infos"`
}

type SplatnetStageClearData struct {
	SplatnetStage SplatnetStage                      `json:"stage"`
	ClearWeapons  map[string]SplatnetWeaponClearData `json:"clear_weapons"`
}

type SplatnetWeaponClearData struct {
	ClearTime uint `json:"clear_time"`
}

type SplatnetStage struct {
	ID     string `json:"id"`
	IsBoss bool   `json:"is_boss"`
	Area   string `json:"area"`
}

type SplatnetCampaignSummary struct {
	SplatnetHonor SplatnetHonor `json:"honor"`
	ClearRate     float64       `json:"clear_rate"`
}

type SplatnetHonor struct {
	Name string `json:"name"`
}

func getLeague() (League, error) {
	league, err := loadLeague()
	if err != nil {
		return league, err
	}
	err = loadSplatnetData(&league)
	if err != nil {
		//save whatever progress we made
		saveLeague(league)
		return league, err
	}
	err = saveLeague(league)
	if err != nil {
		return league, err
	}
	return league, nil
}

func loadLeague() (league League, err error) {
	byteValue, err := ioutil.ReadFile("contestants.json")
	if err != nil {
		return
	}
	err = json.Unmarshal(byteValue, &league)
	return
}

func saveLeague(league League) (err error) {
	output, err := json.MarshalIndent(league, "", "\t")
	if err != nil {
		return
	}
	err = ioutil.WriteFile("contestants.json", output, 0644)
	return
}

func loadSplatnetData(league *League) error {
	errors := make(chan error, len(league.Contestants))
	var wg sync.WaitGroup
	var mu sync.Mutex
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	for index := range league.Contestants {
		wg.Add(1)
		go fetchSplatnetData(&wg, &mu, index, &client, league, errors)
	}
	// Wait for finish
	wg.Wait()
	select {
	case err := <-errors:
		return err
	default:
		return nil
	}
}

func fetchSplatnetData(wg *sync.WaitGroup, mu *sync.Mutex, index int, client *http.Client, league *League, errorch chan error) {
	if league.Contestants[index].ProxyURL != "" {
		fetchSplatnetDataFromProxy(wg, index, client, league, errorch)
	} else {
		fetchSplatnetDataFromNintendo(wg, mu, index, client, league, errorch)
	}
}

func fetchSplatnetDataFromProxy(wg *sync.WaitGroup, index int, client *http.Client, league *League, errorch chan error) {
	defer wg.Done()
	req, err := http.NewRequest("GET", league.Contestants[index].ProxyURL, nil)
	resp, err := client.Do(req)
	if err != nil {
		errorch <- err
		return
	}
	var response ProxyResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		errorch <- err
		return
	}
	league.Contestants[index].SplatnetData = response.SplatnetData
	league.Contestants[index].SplatnetName = response.SplatnetName
	league.Contestants[index].PictureURL = response.PictureURL
}

func fetchSplatnetDataFromNintendo(wg *sync.WaitGroup, mu *sync.Mutex, index int, client *http.Client, league *League, errorch chan error) {
	//stats
	req, err := http.NewRequest("GET", statsURL, nil)
	req.Header.Set("User-Agent", useragent)
	req.Header.Set("Cookie", "iksm_session="+league.Contestants[index].Cookie)
	resp, err := client.Do(req)
	if err != nil {
		errorch <- err
		wg.Done()
		return
	}
	err = json.NewDecoder(resp.Body).Decode(&league.Contestants[index].SplatnetData)
	if err != nil {
		errorch <- err
		wg.Done()
		return
	}
	//name and picture
	req, err = http.NewRequest("GET", resultsURL, nil)
	req.Header.Set("User-Agent", useragent)
	req.Header.Set("Cookie", "iksm_session="+league.Contestants[index].Cookie)
	resp, err = client.Do(req)
	if err != nil {
		errorch <- err
		wg.Done()
		return
	}
	var results BattleResults
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		errorch <- err
		wg.Done()
		return
	}

	req, err = http.NewRequest("GET", nameAndIconURL, nil)
	req.Header.Set("User-Agent", useragent)
	req.Header.Set("Cookie", "iksm_session="+league.Contestants[index].Cookie)
	q := req.URL.Query()
	q.Set("id", results.BattleResults[0].PlayerResult.Player.PrincipalID)
	req.URL.RawQuery = q.Encode()
	resp, err = client.Do(req)
	if err != nil {
		errorch <- err
		wg.Done()
		return
	}
	var profiles SplatnetProfiles
	err = json.NewDecoder(resp.Body).Decode(&profiles)
	if err != nil {
		errorch <- err
		wg.Done()
		return
	}
	if len(profiles.SplatnetProfiles) == 0 {
		newCookie, err := generateCookie(league.Contestants[index].SessionToken, league.Contestants[index].Name, mu)
		if err != nil {
			errorch <- err
			wg.Done()
			return
		}
		league.Contestants[index].Cookie = newCookie
		// Try again
		fetchSplatnetDataFromNintendo(wg, mu, index, client, league, errorch)
	} else {
		league.Contestants[index].SplatnetName = profiles.SplatnetProfiles[0].Name
		league.Contestants[index].PictureURL = profiles.SplatnetProfiles[0].PictureURL
		wg.Done()
	}
}

func generateCookie(sessionToken string, contestantName string, mu *sync.Mutex) (string, error) {
	mu.Lock()
	// Be kind to iksm
	time.Sleep(15 * time.Second)
	fmt.Printf("Invalid cookie for %s, attempting to generate new.\n", contestantName)
	defer mu.Unlock()
	cmd := exec.Command("python3.9", "datahandling/iksm.py", sessionToken)
	out, err := cmd.Output()
	return string(out), err
}
