package datahandling

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
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
	Cookie       string       `json:"cookie"`
	SessionToken string       `json:"session_token"`
	SplatnetName string       `json:"splatnet_name"`
	PictureURL   string       `json:"picture_url"`
	SplatnetData SplatnetData `json:"-"`
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

func GetLeague() (League, error) {
	league, err := loadLeague()
	if err != nil {
		return league, err
	}
	err = loadSplatnetData(&league)
	if err != nil {
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
	client := http.Client{
		Timeout: 2 * time.Second,
	}

	for index := range league.Contestants {
		//stats
		req, err := http.NewRequest("GET", statsURL, nil)
		req.Header.Set("User-Agent", useragent)
		req.Header.Set("Cookie", "iksm_session="+league.Contestants[index].Cookie)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		err = json.NewDecoder(resp.Body).Decode(&league.Contestants[index].SplatnetData)
		if err != nil {
			return err
		}
		//name and picture
		req, err = http.NewRequest("GET", resultsURL, nil)
		req.Header.Set("User-Agent", useragent)
		req.Header.Set("Cookie", "iksm_session="+league.Contestants[index].Cookie)
		resp, err = client.Do(req)
		if err != nil {
			return err
		}
		var results BattleResults
		err = json.NewDecoder(resp.Body).Decode(&results)
		if err != nil {
			return err
		}

		req, err = http.NewRequest("GET", nameAndIconURL, nil)
		req.Header.Set("User-Agent", useragent)
		req.Header.Set("Cookie", "iksm_session="+league.Contestants[index].Cookie)
		q := req.URL.Query()
		q.Set("id", results.BattleResults[0].PlayerResult.Player.PrincipalID)
		req.URL.RawQuery = q.Encode()
		resp, err = client.Do(req)
		if err != nil {
			return err
		}
		var profiles SplatnetProfiles
		err = json.NewDecoder(resp.Body).Decode(&profiles)
		if err != nil {
			return err
		}
		if len(profiles.SplatnetProfiles) == 0 {
			fmt.Printf("Invalid cookie for %s, attempting to generate new.\n", league.Contestants[index].Name)
			newCookie, err := generateCookie(league.Contestants[index].SessionToken)
			if err != nil {
				return err
			}
			league.Contestants[index].Cookie = newCookie
			err = saveLeague(*league)
			if err != nil {
				return err
			}
			return loadSplatnetData(league)
		}
		league.Contestants[index].SplatnetName = profiles.SplatnetProfiles[0].Name
		league.Contestants[index].PictureURL = profiles.SplatnetProfiles[0].PictureURL

	}
	return nil
}

func generateCookie(sessionToken string) (string, error) {
	cmd := exec.Command("python3.9", "datahandling/iksm.py", sessionToken)
	out, err := cmd.Output()
	return string(out), err
}
