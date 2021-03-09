package datahandling

import "strconv"

var (
	titles         = []string{"Profreshional", "Fresh", "Raw", "Dry"}
	stagesBySector = []int{4, 7, 7, 7, 7}
	stages         = 32
	NO_TIME        = uint(9999)
	BEST_WEAPON    = 9
)

type Resutls struct {
	LeagueName    string         `json:"league_name"`
	PlayerResults []PlayerResult `json:"player_result"`
}

//9 Is the overall
type PlayerResult struct {
	PlayerName      string         `json:"player_name"`
	PlayerImage     string         `json:"player_image"`
	PlayerHonor     string         `json:"player_honor"`
	PlayerTitle     string         `json:"player_title"`
	PlayerClearRate float64        `json:"player_clear_rate"`
	TotalScores     ScoreSummary   `json:"total_score"`
	SectorScores    []ScoreSummary `json:"sector_scores"`
	StageScores     []ScoreSummary `json:"stage_scores"`
}

type WeaponScore struct {
	PlayerScore   int  `json:"score"`
	PlayerRanking int  `json:"ranking"`
	PlayerTime    uint `json:"time,omitempty"`
	Weapon        int  `json:"weapon"`
}

type ScoreSummary struct {
	ScoreByWeapon []WeaponScore `json:"score_by_weapon"`
}

func collectWeaponTimes(stage SplatnetStageClearData) (summary ScoreSummary) {
	best := WeaponScore{
		PlayerTime: NO_TIME,
		Weapon:     BEST_WEAPON,
	}
	for weaponID := 0; weaponID < BEST_WEAPON; weaponID++ {
		var time uint
		spatnetWeapon, found := stage.ClearWeapons[strconv.Itoa(weaponID)]
		if !found {
			time = NO_TIME
		} else {
			time = spatnetWeapon.ClearTime
		}
		weaponScore := WeaponScore{
			PlayerTime: time,
			Weapon:     weaponID,
		}
		if time < best.PlayerTime {
			best = weaponScore
		}
		summary.ScoreByWeapon = append(summary.ScoreByWeapon, weaponScore)
	}
	summary.ScoreByWeapon = append(summary.ScoreByWeapon, best)
	return
}

func collectPlayerData(contestant Contestant) (playerResult PlayerResult) {
	playerResult.PlayerImage = contestant.PictureURL
	playerResult.PlayerHonor = contestant.SplatnetData.SplatnetCampaignSummary.SplatnetHonor.Name
	playerResult.PlayerName = contestant.SplatnetName
	playerResult.PlayerClearRate = contestant.SplatnetData.SplatnetCampaignSummary.ClearRate
	for _, stage := range contestant.SplatnetData.SplatnetStageClearDatas {
		playerResult.StageScores = append(playerResult.StageScores, collectWeaponTimes(stage))
	}
	return
}

func CalculateResults(league League) (results Resutls) {
	results.LeagueName = league.LeagueName
	for _, contestant := range league.Contestants {
		results.PlayerResults = append(results.PlayerResults, collectPlayerData(contestant))
	}
	// for stageIndex := 0; stageIndex < stages; stageIndex++ {
	// var StageScorePointers []*ScoreSummary
	// for _, player := range results.PlayerResults {

	// }

	// }
	return
}
