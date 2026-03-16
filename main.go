package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"atlas.guide/internal/app"
)

var Version = "dev"

func main() {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "-v" || arg == "--version" {
			fmt.Printf("atlas.guide v%s\n", Version)
			return
		}
		if arg == "-h" || arg == "--help" || arg == "help" {
			showHelp()
			return
		}
	}

	m := app.NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running atlas.guide: %v\n", err)
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println("Atlas Guide - A personal health and wellness tracker.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  atlas.guide              Start the interactive TUI")
	fmt.Println("  atlas.guide help         Show this help information")
	fmt.Println("  atlas.guide --version    Show version info")
	fmt.Println()
	fmt.Println("Tabs:")
	fmt.Println("  1 - Overview    Day summary with all trackers at a glance")
	fmt.Println("  2 - Monthly     Monthly aggregates and day-by-day heatmap")
	fmt.Println("  3 - Calories    Track food, calorie budget, intermittent fasting")
	fmt.Println("  4 - Gym         Log workouts, duration, calories burnt")
	fmt.Println("  5 - Mood        Rate your day (1-5) with notes")
	fmt.Println("  6 - Medicine    Track medication name & dosage")
	fmt.Println("  7 - Dairy       Track dairy products & amounts")
	fmt.Println()
	fmt.Println("Controls:")
	fmt.Println("  1-7, Tab         Switch tabs")
	fmt.Println("  h/l, Left/Right  Navigate days")
	fmt.Println("  j/k, Up/Down     Move cursor")
	fmt.Println("  t                Jump to today")
	fmt.Println("  c                Open calendar picker")
	fmt.Println("  a                Add entry")
	fmt.Println("  d                Delete entry")
	fmt.Println("  g                Set calorie goal (Calories tab)")
	fmt.Println("  f                Toggle fasting (Calories tab)")
	fmt.Println("  ?                Show help in TUI")
	fmt.Println("  q, Ctrl+C        Quit")
	fmt.Println()
	fmt.Println("Data is stored in ~/.atlas/atlas.guide.data/")
}
