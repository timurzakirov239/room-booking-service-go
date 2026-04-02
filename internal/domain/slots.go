package domain

import "time"

const SlotDuration = 30 * time.Minute

type SlotWindow struct {
	StartAt time.Time
	EndAt   time.Time
}

func BuildRollingSlotWindows(daysOfWeek []int16, startTime time.Time, endTime time.Time, from time.Time, to time.Time) ([]SlotWindow, error) {
	fromUTC := from.UTC()
	toUTC := to.UTC()
	if !fromUTC.Before(toUTC) {
		return nil, ErrInvalidTimeRange
	}

	startUTC := startTime.UTC()
	endUTC := endTime.UTC()
	if !startUTC.Before(endUTC) {
		return nil, ErrInvalidTimeRange
	}

	allowedDays := make(map[time.Weekday]struct{}, len(daysOfWeek))
	for _, day := range daysOfWeek {
		if day < 1 || day > 7 {
			return nil, ErrInvalidTimeRange
		}
		allowedDays[isoDayToWeekday(day)] = struct{}{}
	}

	windows := make([]SlotWindow, 0)
	for day := truncateToUTCDay(fromUTC); day.Before(toUTC); day = day.AddDate(0, 0, 1) {
		if _, ok := allowedDays[day.Weekday()]; !ok {
			continue
		}

		windowStart := time.Date(day.Year(), day.Month(), day.Day(), startUTC.Hour(), startUTC.Minute(), startUTC.Second(), 0, time.UTC)
		windowEnd := time.Date(day.Year(), day.Month(), day.Day(), endUTC.Hour(), endUTC.Minute(), endUTC.Second(), 0, time.UTC)
		if !windowStart.Before(windowEnd) {
			return nil, ErrInvalidTimeRange
		}

		for slotStart := windowStart; slotStart.Add(SlotDuration) <= windowEnd; slotStart = slotStart.Add(SlotDuration) {
			slotEnd := slotStart.Add(SlotDuration)
			if slotEnd <= fromUTC || !slotStart.Before(toUTC) {
				continue
			}
			windows = append(windows, SlotWindow{StartAt: slotStart, EndAt: slotEnd})
		}
	}

	return windows, nil
}

func truncateToUTCDay(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func isoDayToWeekday(day int16) time.Weekday {
	switch day {
	case 1:
		return time.Monday
	case 2:
		return time.Tuesday
	case 3:
		return time.Wednesday
	case 4:
		return time.Thursday
	case 5:
		return time.Friday
	case 6:
		return time.Saturday
	default:
		return time.Sunday
	}
}
