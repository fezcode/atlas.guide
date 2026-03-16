package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// FoodEntry represents a single food item logged.
type FoodEntry struct {
	Name     string `json:"name"`
	Calories int    `json:"calories"`
	Protein  int    `json:"protein"`
	Time     string `json:"time"`
}

// FastingWindow represents an intermittent fasting schedule.
type FastingWindow struct {
	Start  string `json:"start"`
	End    string `json:"end"`
	Active bool   `json:"active"`
}

// CalorieDay stores all calorie data for a single day.
type CalorieDay struct {
	Entries []FoodEntry   `json:"entries"`
	Fasting FastingWindow `json:"fasting"`
	Goal    int           `json:"goal"`
}

// GymActivity represents a single workout/activity.
type GymActivity struct {
	Name         string `json:"name"`
	DurationMin  int    `json:"duration_min"`
	CaloriesBurnt int   `json:"calories_burnt"`
	Time         string `json:"time"`
}

// GymDay stores all gym data for a single day.
type GymDay struct {
	Activities []GymActivity `json:"activities"`
}

// MoodEntry stores mood data for a single day.
type MoodEntry struct {
	Rating int    `json:"rating"`
	Note   string `json:"note"`
	Time   string `json:"time"`
}

// MedicineEntry represents a single medicine dose logged.
type MedicineEntry struct {
	Name   string `json:"name"`
	Dosage string `json:"dosage"`
	Time   string `json:"time"`
}

// MedicineDay stores all medicine data for a single day.
type MedicineDay struct {
	Entries []MedicineEntry `json:"entries"`
}

// DairyEntry represents a single dairy product logged.
type DairyEntry struct {
	Name   string `json:"name"`
	Amount int    `json:"amount"`
	Unit   string `json:"unit"`
	Time   string `json:"time"`
}

// DairyDay stores all dairy data for a single day.
type DairyDay struct {
	Entries []DairyEntry `json:"entries"`
}

func dataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".atlas", "atlas.guide.data")
}

func dayDir(date time.Time) string {
	return filepath.Join(dataDir(), date.Format("2006-01-02"))
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func loadJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, v)
}

func saveJSON(path string, v interface{}) error {
	dir := filepath.Dir(path)
	if err := ensureDir(dir); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadCalories loads calorie data for a given date.
func LoadCalories(date time.Time) CalorieDay {
	path := filepath.Join(dayDir(date), "calories.json")
	var day CalorieDay
	_ = loadJSON(path, &day)
	if day.Goal == 0 {
		day.Goal = 2000
	}
	if day.Entries == nil {
		day.Entries = []FoodEntry{}
	}
	return day
}

// SaveCalories persists calorie data for a given date.
func SaveCalories(date time.Time, day CalorieDay) error {
	path := filepath.Join(dayDir(date), "calories.json")
	return saveJSON(path, day)
}

// LoadGym loads gym data for a given date.
func LoadGym(date time.Time) GymDay {
	path := filepath.Join(dayDir(date), "gym.json")
	var day GymDay
	_ = loadJSON(path, &day)
	if day.Activities == nil {
		day.Activities = []GymActivity{}
	}
	return day
}

// SaveGym persists gym data for a given date.
func SaveGym(date time.Time, day GymDay) error {
	path := filepath.Join(dayDir(date), "gym.json")
	return saveJSON(path, day)
}

// LoadMood loads mood data for a given date.
func LoadMood(date time.Time) MoodEntry {
	path := filepath.Join(dayDir(date), "mood.json")
	var entry MoodEntry
	_ = loadJSON(path, &entry)
	return entry
}

// SaveMood persists mood data for a given date.
func SaveMood(date time.Time, entry MoodEntry) error {
	path := filepath.Join(dayDir(date), "mood.json")
	return saveJSON(path, entry)
}

// LoadMedicine loads medicine data for a given date.
func LoadMedicine(date time.Time) MedicineDay {
	path := filepath.Join(dayDir(date), "medicine.json")
	var day MedicineDay
	_ = loadJSON(path, &day)
	if day.Entries == nil {
		day.Entries = []MedicineEntry{}
	}
	return day
}

// SaveMedicine persists medicine data for a given date.
func SaveMedicine(date time.Time, day MedicineDay) error {
	path := filepath.Join(dayDir(date), "medicine.json")
	return saveJSON(path, day)
}

// LoadDairy loads dairy data for a given date.
func LoadDairy(date time.Time) DairyDay {
	path := filepath.Join(dayDir(date), "dairy.json")
	var day DairyDay
	_ = loadJSON(path, &day)
	if day.Entries == nil {
		day.Entries = []DairyEntry{}
	}
	return day
}

// SaveDairy persists dairy data for a given date.
func SaveDairy(date time.Time, day DairyDay) error {
	path := filepath.Join(dayDir(date), "dairy.json")
	return saveJSON(path, day)
}

// TotalCaloriesConsumed returns total calories from food entries.
func TotalCaloriesConsumed(day CalorieDay) int {
	total := 0
	for _, e := range day.Entries {
		total += e.Calories
	}
	return total
}

// TotalProtein returns total protein from food entries.
func TotalProtein(day CalorieDay) int {
	total := 0
	for _, e := range day.Entries {
		total += e.Protein
	}
	return total
}

// TotalCaloriesBurnt returns total calories burnt from gym activities.
func TotalCaloriesBurnt(day GymDay) int {
	total := 0
	for _, a := range day.Activities {
		total += a.CaloriesBurnt
	}
	return total
}

// TotalGymDuration returns total minutes spent working out.
func TotalGymDuration(day GymDay) int {
	total := 0
	for _, a := range day.Activities {
		total += a.DurationMin
	}
	return total
}

// TotalDairyAmount returns total dairy amount for a day.
func TotalDairyAmount(day DairyDay) int {
	total := 0
	for _, e := range day.Entries {
		total += e.Amount
	}
	return total
}

// NetCalories returns consumed minus burnt.
func NetCalories(cal CalorieDay, gym GymDay) int {
	return TotalCaloriesConsumed(cal) - TotalCaloriesBurnt(gym)
}

// CalorieRemaining returns how many calories left in the budget.
func CalorieRemaining(cal CalorieDay, gym GymDay) int {
	return cal.Goal - NetCalories(cal, gym)
}

// DaySummary holds aggregated stats for a single day.
type DaySummary struct {
	Date          time.Time
	HasData       bool
	CalConsumed   int
	CalGoal       int
	CalBurnt      int
	CalNet        int
	Protein       int
	FastingActive bool
	GymDuration   int
	GymSessions   int
	MoodRating    int
	MedicineCount int
	DairyCount    int
}

// LoadDaySummary loads a lightweight summary for a single date.
func LoadDaySummary(date time.Time) DaySummary {
	cal := LoadCalories(date)
	gym := LoadGym(date)
	mood := LoadMood(date)
	med := LoadMedicine(date)
	dairy := LoadDairy(date)

	consumed := TotalCaloriesConsumed(cal)
	burnt := TotalCaloriesBurnt(gym)

	s := DaySummary{
		Date:          date,
		CalConsumed:   consumed,
		CalGoal:       cal.Goal,
		CalBurnt:      burnt,
		CalNet:        consumed - burnt,
		Protein:       TotalProtein(cal),
		FastingActive: cal.Fasting.Active,
		GymDuration:   TotalGymDuration(gym),
		GymSessions:   len(gym.Activities),
		MoodRating:    mood.Rating,
		MedicineCount: len(med.Entries),
		DairyCount:    len(dairy.Entries),
	}
	s.HasData = len(cal.Entries) > 0 || len(gym.Activities) > 0 || mood.Rating > 0 || len(med.Entries) > 0 || len(dairy.Entries) > 0
	return s
}

// LoadMonthSummaries loads summaries for every day in a given month.
func LoadMonthSummaries(year int, month time.Month) []DaySummary {
	days := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
	summaries := make([]DaySummary, days)
	for d := 1; d <= days; d++ {
		date := time.Date(year, month, d, 0, 0, 0, 0, time.Local)
		summaries[d-1] = LoadDaySummary(date)
	}
	return summaries
}
