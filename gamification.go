package main

import (
	"time"
)

// get3AMDay returns the "day" string based on 3AM cutoff
// Times before 3AM belong to the previous calendar day
func get3AMDay(t time.Time) string {
	// If time is before 3AM, it belongs to the previous day
	if t.Hour() < 3 {
		t = t.AddDate(0, 0, -1)
	}
	return t.Format("2006-01-02")
}

func updateTaskStreak(daily *Daily) {
	now := time.Now()
	today := get3AMDay(now)

	// Check if already completed today using LastCompleted
	if !daily.LastCompleted.IsZero() && get3AMDay(daily.LastCompleted) == today {
		return // Already completed today, don't update streak
	}

	// Calculate yesterday based on 3AM cutoff
	yesterdayTime := now.Add(-24 * time.Hour)
	yesterday := get3AMDay(yesterdayTime)

	// Update streak based on whether we completed yesterday
	if daily.LastCompleted.IsZero() {
		// First time completing this task
		daily.CurrentStreak = 1
	} else if get3AMDay(daily.LastCompleted) == yesterday {
		// Completed yesterday - continue streak
		daily.CurrentStreak++
	} else {
		// Streak broken - reset to 1
		daily.CurrentStreak = 1
	}

	// Update best streak if current is higher
	if daily.CurrentStreak > daily.BestStreak {
		daily.BestStreak = daily.CurrentStreak
	}
}
