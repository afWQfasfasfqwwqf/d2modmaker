package main

import (
	"fmt"
	"os"
	"strconv"
	"math"

	charStats "github.com/tlentz/d2modmaker/internal/charStatsTxt"
	"github.com/tlentz/d2modmaker/internal/cubeMainTxt"
	"github.com/tlentz/d2modmaker/internal/d2file"
	itmRatio "github.com/tlentz/d2modmaker/internal/itemRatioTxt"
	levels "github.com/tlentz/d2modmaker/internal/levelsTxt"
	misc "github.com/tlentz/d2modmaker/internal/miscTxt"
	missiles "github.com/tlentz/d2modmaker/internal/missilesTxt"
	skills "github.com/tlentz/d2modmaker/internal/skillsTxt"
	"github.com/tlentz/d2modmaker/internal/superUniquesTxt"
	tc "github.com/tlentz/d2modmaker/internal/treasureClassExTxt"
	"github.com/tlentz/d2modmaker/internal/util"
)

var (
	dataDir string
	outDir  string
	cfgPath string
	mode    string
	version string
)

func withDefault(a, b string) string {
	if a == "" {
		return b
	}
	return a
}

func main() {
	if mode == "production" {
		dataDir = "./113c-data/"
		outDir = "./data/global/excel/"
		cfgPath = "./cfg.json"
	} else {
		dataDir = "../../assets/113c-data/"
		outDir = "../../dist/"
		cfgPath = "../../cfg.json"
	}

	if version == "" {
		version = "[Dev Build]"
	}
	line := "==============================="
	fmt.Println(line)
	fmt.Println("", "D2 Mod Maker", version)
	fmt.Println(line)
	makeMod()
	// printFile()
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func printFile() {
	d2files := d2file.D2Files{}
	f := d2file.GetOrCreateFile(dataDir, d2files, superUniquesTxt.FileName)
	for i := range f.Headers {
		fmt.Println(f.Headers[i], " = ", i)
	}
}

func makeMod() {
	cfg := ReadCfg(cfgPath)
	d2files := d2file.D2Files{}

	// fmt.Println("removing " + outDir)
	os.RemoveAll(outDir)
	// fmt.Println("creating " + outDir)
	err := os.MkdirAll(outDir, 0755)
	util.Check(err)

	if cfg.IncreaseStackSizes {
		increaseStackSizes(d2files)
	}

	if cfg.IncreaseMonsterDensity > 0 {
		increaseMonsterDensity(d2files, cfg.IncreaseMonsterDensity)
	}

	if cfg.EnableTownSkills {
		enableTownSkills(d2files)
	}
	if cfg.NoDropZero {
		noDropZero(d2files)
	}
	if cfg.QuestDrops {
		questDrops(d2files)
	}
	if cfg.Cowzzz {
		cowzzz(d2files)
	}

	if cfg.UniqueItemDropRate > 0 {
		uniqueItemDropRate(d2files, cfg.UniqueItemDropRate)
	}
	
	if cfg.RuneDropRate > 0 {
		runeDropRate(d2files, cfg.RuneDropRate);
	}

	if cfg.StartWithCube {
		startWithCube(d2files)
	}
	if cfg.RandomOptions.Randomize {
		Randomize(&cfg, d2files)
	}

	d2file.WriteFiles(d2files, outDir)
	writeSeed(cfg)
	util.PP(cfg)
	fmt.Println("===========================")
	fmt.Println("Done!")
	if cfg.EnterToExit {
		fmt.Println("\n[Press enter to exit]")
		fmt.Scanln() // wait for Enter Key
	}
}

func writeSeed(cfg ModConfig) {
	filePath := outDir + "Seed.txt"
	// fmt.Println("writing " + filePath)
	f, err := os.Create(filePath)
	util.Check(err)
	defer f.Close()
	f.WriteString(fmt.Sprintf("%d\n", cfg.RandomOptions.Seed))
}

func cowzzz(d2files d2file.D2Files) {
	// Add New Recipe for Cow Poral (tp scroll -> cow portal)
	cubeF := d2file.GetOrCreateFile(dataDir, d2files, cubeMainTxt.FileName)
	// newCubeRows := make([][]string, 0)
	for _, row := range cubeF.Rows {
		description := row[cubeMainTxt.Description]

		if description == cubeMainTxt.CowPortalWirt {
			tmp := make([]string, len(row))
			// // copy cow row to tmp
			copy(tmp, row)

			// change tmp to remove wirts leg
			tmp[cubeMainTxt.Description] = cubeMainTxt.CowPortalNoWirt
			tmp[cubeMainTxt.NumInputs] = "1"
			tmp[cubeMainTxt.Input1] = "tsc"
			tmp[cubeMainTxt.Input2] = ""

			cubeF.Rows = append(cubeF.Rows, tmp)

		}
	}

	// Enable ability to kill cow king and still create portal
	suF := d2file.GetOrCreateFile(dataDir, d2files, superUniquesTxt.FileName)
	for idx, row := range suF.Rows {
		name := row[superUniquesTxt.Name]

		if name == superUniquesTxt.CowKing {
			suF.Rows[idx][superUniquesTxt.HcIdx] = "1"
		}
	}
}

func startWithCube(d2files d2file.D2Files) {
	f := d2file.GetOrCreateFile(dataDir, d2files, charStats.FileName)
	itemOffset := charStats.Item1
	countOffset := 2
	for idx, row := range f.Rows {
		for i := itemOffset; i < len(row)-countOffset; i += 3 {
			if (row[i] == "0" || row[i] == "") && (row[i+countOffset] == "0" || row[i+countOffset] == "") {
				f.Rows[idx][i] = "box"
				f.Rows[idx][i+countOffset] = "1"
				break // we added a cube, we are done with this row
			}
		}
	}
}

func uniqueItemDropRate(d2files d2file.D2Files, d float64) {
	f := d2file.GetOrCreateFile(dataDir, d2files, itmRatio.FileName)

	one := func(n int) int {
		if n < 1 {
			return 1
		}
		return n
	}

	divU := func(n int) int {
		return one(int(float64(n) / d))
	}

	divM := func(n int) int {
		dM := 1 + (d - 1) / 10
		return one(int(float64(n) / dM))
	}

	for i := range f.Rows {

		// Uniques
		oldUnique, err1 := strconv.Atoi(f.Rows[i][itmRatio.Unique])
		oldUniqueMin, err2 := strconv.Atoi(f.Rows[i][itmRatio.UniqueMin])
		if err1 == nil && err2 == nil {
			newUnique := divU(oldUnique)
			newUniqueMin := divM(oldUniqueMin)
			f.Rows[i][itmRatio.Unique] = strconv.Itoa(newUnique)
			f.Rows[i][itmRatio.UniqueMin] = strconv.Itoa(newUniqueMin)
		}

		// Sets
		oldSet, err3 := strconv.Atoi(f.Rows[i][itmRatio.Set])
		oldSetMin, err4 := strconv.Atoi(f.Rows[i][itmRatio.SetMin])
		if err3 == nil && err4 == nil {
			newSet := divU(oldSet)
			newSetMin := divM(oldSetMin)
			f.Rows[i][itmRatio.Set] = strconv.Itoa(newSet)
			f.Rows[i][itmRatio.SetMin] = strconv.Itoa(newSetMin)
		}
	}
}

func runeDropRate(d2files d2file.D2Files, rateScale float64) {
	f := d2file.GetOrCreateFile(dataDir, d2files, tc.FileName)

	// Clip rateScale to valid 1-100 range
	if rateScale > 100.0 {
		rateScale = 100.0
	}
	if rateScale < 1.0 {
		rateScale = 1.0
	}
	// Convert rateScale to 0.0-1.0 range
	rateScale = ( rateScale - 1.0 ) / 99.0
	
	// Invert rate scale to get the drop rate multiplier
	rateMult := 1.0 - rateScale 
	
	targetProb12 := 1.0
	
	origProb1 := 3.0
	baseProb1 := math.Log2(targetProb12)
	rangeProb1 := math.Log2(origProb1) - math.Log2(targetProb12)
	newProb1 := math.Pow( 2, baseProb1 + rangeProb1 * rateMult )
	
	origProb2 := 2.0
	baseProb2 := math.Log2(targetProb12)
	rangeProb2 := math.Log2(origProb2) - math.Log2(targetProb12)
	newProb2 := math.Pow( 2, baseProb2 + rangeProb2 * rateMult )
	
	for idx, row := range f.Rows {
		treasureClass := row[tc.TreasureClass]
		if len(treasureClass) >= 5 && treasureClass[:5] == "Runes" {
			runeTc, err := strconv.Atoi(row[tc.TreasureClass][6:])
			
			if err == nil {			
				targetProb3	:= float64((runeTc - 1) * 2)
				var newProb3 float64 = 0
				if targetProb3 != 0 {	
					var origProb3 int
					if runeTc != 17 {
						origProb3, _ = strconv.Atoi(row[tc.Prob3])	
					} else {
						// Runes 17 doesn't have a second rune chance, and the next rune TC slot is in Prob2
						origProb3, _ = strconv.Atoi(row[tc.Prob2])	
					}
					baseProb3 := math.Log2(targetProb3)
					rangeProb3 := math.Log2(float64(origProb3)) - math.Log2(targetProb3)
					newProb3 = math.Pow( 2, baseProb3 + rangeProb3 * rateMult )
				}
				
				if runeTc != 17 {
					f.Rows[idx][tc.Prob1] = strconv.Itoa(int(newProb1+0.5))
					f.Rows[idx][tc.Prob2] = strconv.Itoa(int(newProb2+0.5))
					if runeTc != 1 {
						// Runes 1 does not have a next rune TC
						f.Rows[idx][tc.Prob3] = strconv.Itoa(int(newProb3+0.5))
					}
				} else {
					// Runes 17 doesn't have a second rune chance, and the next rune TC slot is in Prob2
					f.Rows[idx][tc.Prob2] = strconv.Itoa(int(newProb3+0.5))
				}
			}
		}
	}
}

func questDrops(d2files d2file.D2Files) {
	f := d2file.GetOrCreateFile(dataDir, d2files, tc.FileName)
	diffOffsets := []int{0, 1, 2} // norm, nm, hell
	bossQOffset := 3
	for idx, row := range f.Rows {
		switch row[tc.TreasureClass] {
		case tc.Andariel, tc.Duriel, tc.DurielBase, tc.Mephisto, tc.Diablo, tc.Baal:
			{
				for _, offset := range diffOffsets {
					tmp := make([]string, len(row))
					// copy quest drop row to tmp
					copy(tmp, f.Rows[idx+bossQOffset+offset])
					// copy all tmp values except 1st index to original row
					copy(f.Rows[idx+offset][tc.TreasureClass+1:], tmp[tc.TreasureClass+1:])
				}
			}
		}
	}
}

func noDropZero(d2files d2file.D2Files) {
	f := d2file.GetOrCreateFile(dataDir, d2files, tc.FileName)
	for idx, row := range f.Rows {
		if row[tc.NoDrop] != "" {
			f.Rows[idx][tc.NoDrop] = "0"
		}
	}
}

func increaseStackSizes(d2files d2file.D2Files) {
	f := d2file.GetOrCreateFile(dataDir, d2files, misc.FileName)
	for idx, row := range f.Rows {
		if row[misc.Name] == misc.TownPortalBook || row[misc.Name] == misc.IdentifyBook || row[misc.Name] == misc.SkeletonKey {
			f.Rows[idx][misc.MaxStack] = "100"
		}
		if row[misc.Name] == misc.Arrows || row[misc.Name] == misc.Bolts {
			f.Rows[idx][misc.MaxStack] = "511"
		}
	}

}

func enableTownSkills(d2files d2file.D2Files) {
	sktxt := d2file.GetOrCreateFile(dataDir, d2files, skills.FileName)
	for i := range sktxt.Rows {
		sktxt.Rows[i][skills.InTown] = "1"
	}
	missilestxt := d2file.GetOrCreateFile(dataDir, d2files, missiles.FileName)
	for i := range missilestxt.Rows {
		missilestxt.Rows[i][missiles.Town] = "1"
	}
}

func increaseMonsterDensity(d2files d2file.D2Files, m float64) {
	f := d2file.GetOrCreateFile(dataDir, d2files, levels.FileName)
	maxM := 30.0
	mult := util.MinFloat(m, maxM)
	maxDensity := 10000

	increaseNumByMult := func(n int) int {
		return int(float64(n) * mult)
	}

	diffOffsets := []int{0, 1, 2} // norm, nm, hell
	minMaxOffsets := []int{0, 1}  // min, max

	for rowIdx := range f.Rows {

		for _, diffOffset := range diffOffsets {

			// MonDen
			colIdx := levels.MonDen + diffOffset
			oldVal, err := strconv.Atoi(f.Rows[rowIdx][colIdx])
			if err == nil {
				newVal := util.MinInt(maxDensity, increaseNumByMult(oldVal))
				f.Rows[rowIdx][colIdx] = strconv.Itoa(newVal)
			}

			// MonUMin / MonUMax
			for _, minMaxOffset := range minMaxOffsets {
				colIdx := levels.MonUMin + minMaxOffset + diffOffset
				oldVal, err := strconv.Atoi(f.Rows[rowIdx][colIdx])
				if err == nil {
					newVal := increaseNumByMult(oldVal)
					f.Rows[rowIdx][colIdx] = strconv.Itoa(newVal)
				}
			}
		}
	}
}
