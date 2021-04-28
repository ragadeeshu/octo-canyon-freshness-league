package datahandling

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"
)

var (
	titles           = []string{"Profreshional", "Fresh", "Raw", "Dry"}
	stagesBySector   = []int{4, 7, 7, 7, 7}
	stageIDsBySector = [][]int{{0, 1, 2, 3}, {4, 5, 6, 7, 8, 9, 10}, {11, 12, 13, 14, 15, 16, 17}, {18, 19, 20, 21, 22, 23, 24}, {25, 26, 27, 28, 29, 30, 31}}
	weaponNames      = []string{"Hero Shot", "Hero Roller", "Hero Charger", "Hero Dualies", "Hero Slosher", "Hero Splatling", "Hero Blaster", "Hero Brella", "Herobrush"}
	stages           = 32
	NO_TIME          = uint(9999)
	BEST_WEAPON      = 9
	stageIDs         = map[string]int{
		"1":   0,
		"2":   1,
		"3":   2,
		"101": 3,
		"4":   4,
		"5":   5,
		"6":   6,
		"7":   7,
		"8":   8,
		"9":   9,
		"102": 10,
		"10":  11,
		"11":  12,
		"12":  13,
		"13":  14,
		"14":  15,
		"15":  16,
		"103": 17,
		"16":  18,
		"17":  19,
		"18":  20,
		"19":  21,
		"20":  22,
		"21":  23,
		"104": 24,
		"22":  25,
		"23":  26,
		"24":  27,
		"25":  28,
		"26":  29,
		"27":  30,
		"105": 31,
	}
	stageIDList = []int{1, 2, 3, 101, 4, 5, 6, 7, 8, 9, 102, 10, 11, 12, 13, 14, 15, 103, 16, 17, 18, 19, 20, 21, 104, 22, 23, 24, 25, 26, 27, 105}
)

type Results struct {
	LeagueName     string         `json:"league_name"`
	Date           time.Time      `json:"time"`
	PlayerResults  []PlayerResult `json:"player_result"`
	StagesBySector [][]int        `json:"stages_by_sector"`
	StageIDList    []int
}

//9 Is the overall
type PlayerResult struct {
	PlayerName      string         `json:"player_name"`
	PlayerImage     string         `json:"player_image"`
	PlayerHonor     string         `json:"player_honor"`
	PlayerTitle     string         `json:"player_title"`
	PlayerClearRate string         `json:"player_clear_rate"`
	BestWeapon      string         `json:"best_weapon"`
	WorstWeapon     string         `json:"worst_weapon"`
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
		var weapon int
		spatnetWeapon, found := stage.ClearWeapons[strconv.Itoa(weaponID)]
		if !found {
			time = NO_TIME
			weapon = BEST_WEAPON
		} else {
			time = spatnetWeapon.ClearTime
			weapon = weaponID
		}
		weaponScore := WeaponScore{
			PlayerTime: time,
			Weapon:     weapon,
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
	clearPrecentage := int(math.Round(contestant.SplatnetData.SplatnetCampaignSummary.ClearRate * 100))
	playerResult.PlayerClearRate = fmt.Sprintf("%d%%", clearPrecentage)
	playerResult.StageScores = make([]ScoreSummary, stages)

	for stageIndex := 0; stageIndex < stages; stageIndex++ {
		var emptyScore ScoreSummary
		for weaponID := 0; weaponID <= BEST_WEAPON; weaponID++ {
			noWeaponData := WeaponScore{
				PlayerTime: NO_TIME,
				Weapon:     BEST_WEAPON,
			}
			emptyScore.ScoreByWeapon = append(emptyScore.ScoreByWeapon, noWeaponData)
		}
		playerResult.StageScores[stageIndex] = emptyScore
	}
	for _, stage := range contestant.SplatnetData.SplatnetStageClearDatas {
		playerResult.StageScores[stageIDs[stage.SplatnetStage.ID]] = collectWeaponTimes(stage)
	}
	return
}

func calculateStageResults(results *Results) {
	// for each stage
	for stageIndex := 0; stageIndex < stages; stageIndex++ {
		var stageScorePointers []*ScoreSummary
		// gather pointers to the result for each player
		for _, player := range results.PlayerResults {
			stageScorePointers = append(stageScorePointers, &player.StageScores[stageIndex])
		}
		// for each weapon
		for weaponID := 0; weaponID <= BEST_WEAPON; weaponID++ {
			// sort the results
			sort.Slice(stageScorePointers, func(i, j int) bool {
				return stageScorePointers[i].ScoreByWeapon[weaponID].PlayerTime < stageScorePointers[j].ScoreByWeapon[weaponID].PlayerTime
			})
			// assign ranks and points.
			pointsForWinning := 1
			if weaponID == BEST_WEAPON {
				pointsForWinning = 10
			}
			rank := 1
			prevTime := stageScorePointers[0].ScoreByWeapon[weaponID].PlayerTime
			for playerIndex := range stageScorePointers {

				if stageScorePointers[playerIndex].ScoreByWeapon[weaponID].PlayerTime > prevTime {
					rank++
					prevTime = stageScorePointers[playerIndex].ScoreByWeapon[weaponID].PlayerTime
				}
				stageScorePointers[playerIndex].ScoreByWeapon[weaponID].PlayerRanking = rank
				if rank == 1 {
					stageScorePointers[playerIndex].ScoreByWeapon[weaponID].PlayerScore += pointsForWinning
					if weaponID != BEST_WEAPON {
						stageScorePointers[playerIndex].ScoreByWeapon[BEST_WEAPON].PlayerScore += pointsForWinning
					}
				}
			}
		}
	}
}

func calculateSectorResults(results *Results) {
	stagePositionOffset := 0
	// for each sector
	for sector, sectorSize := range stagesBySector {
		var sectorScorePointers []*ScoreSummary
		// gather the score for each player
		for playerIndex := range results.PlayerResults {
			var sectorScore ScoreSummary
			sectorScore.ScoreByWeapon = make([]WeaponScore, 10)
			// for each stage in the sector
			for stageInSector := stagePositionOffset; stageInSector < stagePositionOffset+sectorSize; stageInSector++ {
				// add the score for each weapon to the total for that weapon
				for weaponID := 0; weaponID <= BEST_WEAPON; weaponID++ {
					sectorScore.ScoreByWeapon[weaponID].Weapon = weaponID
					sectorScore.ScoreByWeapon[weaponID].PlayerScore += results.PlayerResults[playerIndex].StageScores[stageInSector].ScoreByWeapon[weaponID].PlayerScore
				}

			}
			results.PlayerResults[playerIndex].SectorScores = append(results.PlayerResults[playerIndex].SectorScores, sectorScore)
			sectorScorePointers = append(sectorScorePointers, &results.PlayerResults[playerIndex].SectorScores[sector])
		}
		// sort the player restults to get the rank
		// for each weapon
		for weaponID := 0; weaponID <= BEST_WEAPON; weaponID++ {
			// sort the results
			sort.Slice(sectorScorePointers, func(i, j int) bool {
				return sectorScorePointers[i].ScoreByWeapon[weaponID].PlayerScore > sectorScorePointers[j].ScoreByWeapon[weaponID].PlayerScore
			})
			// assign ranks
			rank := 1
			prevScore := sectorScorePointers[0].ScoreByWeapon[weaponID].PlayerScore
			for playerIndex := range sectorScorePointers {
				if sectorScorePointers[playerIndex].ScoreByWeapon[weaponID].PlayerScore < prevScore {
					rank++
					prevScore = sectorScorePointers[playerIndex].ScoreByWeapon[weaponID].PlayerScore
				}
				sectorScorePointers[playerIndex].ScoreByWeapon[weaponID].PlayerRanking = rank
			}
		}
		// go to next sector
		stagePositionOffset += sectorSize
	}

}

func calculateTotalResults(results *Results) {
	var totalScorePointers []*ScoreSummary
	// gather the score for each player
	for playerIndex := range results.PlayerResults {
		var totalScore ScoreSummary
		totalScore.ScoreByWeapon = make([]WeaponScore, 10)
		// for each stage
		for stageIndex := 0; stageIndex < stages; stageIndex++ {
			// add the score for each weapon to the total for that weapon
			for weaponID := 0; weaponID <= BEST_WEAPON; weaponID++ {
				totalScore.ScoreByWeapon[weaponID].Weapon = weaponID
				totalScore.ScoreByWeapon[weaponID].PlayerScore += results.PlayerResults[playerIndex].StageScores[stageIndex].ScoreByWeapon[weaponID].PlayerScore
				totalScore.ScoreByWeapon[weaponID].PlayerTime += results.PlayerResults[playerIndex].StageScores[stageIndex].ScoreByWeapon[weaponID].PlayerTime
			}

		}
		results.PlayerResults[playerIndex].TotalScores = totalScore
		totalScorePointers = append(totalScorePointers, &results.PlayerResults[playerIndex].TotalScores)
	}
	// sort the player restults to get the rank
	// for each weapon
	for weaponID := 0; weaponID <= BEST_WEAPON; weaponID++ {
		// sort the results
		sort.Slice(totalScorePointers, func(i, j int) bool {
			return totalScorePointers[i].ScoreByWeapon[weaponID].PlayerScore > totalScorePointers[j].ScoreByWeapon[weaponID].PlayerScore
		})
		// assign ranks
		rank := 1
		prevScore := totalScorePointers[0].ScoreByWeapon[weaponID].PlayerScore
		for playerIndex := range totalScorePointers {
			if totalScorePointers[playerIndex].ScoreByWeapon[weaponID].PlayerScore < prevScore {
				rank++
				prevScore = totalScorePointers[playerIndex].ScoreByWeapon[weaponID].PlayerScore
			}
			totalScorePointers[playerIndex].ScoreByWeapon[weaponID].PlayerRanking = rank
		}
	}
	// Best and worst weapon
	for i := range results.PlayerResults {
		best := results.PlayerResults[i].TotalScores.ScoreByWeapon[0].PlayerTime
		bestID := 0
		worst := results.PlayerResults[i].TotalScores.ScoreByWeapon[0].PlayerTime
		worstID := 0
		for weaponID := 0; weaponID < BEST_WEAPON; weaponID++ {
			time := results.PlayerResults[i].TotalScores.ScoreByWeapon[weaponID].PlayerTime
			if time < best {
				best = time
				bestID = weaponID
			}
			if time > worst {
				worst = time
				worstID = weaponID
			}
		}
		results.PlayerResults[i].BestWeapon = weaponNames[bestID]
		results.PlayerResults[i].WorstWeapon = weaponNames[worstID]
	}
	// Rank titles
	for i := range results.PlayerResults {
		rank := results.PlayerResults[i].TotalScores.ScoreByWeapon[9].PlayerRanking
		titleindex := min(rank, len(titles)) - 1
		results.PlayerResults[i].PlayerTitle = titles[titleindex]
	}
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func CalculateResults(league League) (results Results) {
	results.LeagueName = league.LeagueName
	results.StagesBySector = stageIDsBySector
	results.StageIDList = stageIDList
	for _, contestant := range league.Contestants {
		results.PlayerResults = append(results.PlayerResults, collectPlayerData(contestant))
	}
	calculateStageResults(&results)
	calculateSectorResults(&results)
	calculateTotalResults(&results)
	results.Date = time.Now()
	return
}
