package usecase

import "time"

// kalaYok is the day-level portion of the Thai Kala Yok calculation.  The
// current auspicious finder deliberately uses this as a supporting signal;
// it does not attempt to produce a full electional chart.
type kalaYok struct {
	thongChai bool
	athibodi  bool
	ubat      bool
	lokawinat bool
}

// kalaYokForDate follows the common simplified calendar convention: the
// Thai astrological year changes on 16 April.  The calculation is based on
// Chula Sakarat (B.E. - 1181), with a remainder of zero representing
// Saturday.  Keeping it here makes the date boundary explicit and testable.
func kalaYokForDate(date time.Time) kalaYok {
	thaiYear := date.Year() + 543
	if date.Month() < time.April || (date.Month() == time.April && date.Day() < 16) {
		thaiYear--
	}
	chulaSakarat := thaiYear - 1181
	weekday := date.Weekday()

	return kalaYok{
		thongChai: weekday == kalaYokWeekday(chulaSakarat*10+3),
		athibodi:  weekday == kalaYokWeekday(chulaSakarat%498),
		ubat:      weekday == kalaYokWeekday(chulaSakarat*10+2),
		lokawinat: weekday == kalaYokWeekday(chulaSakarat+1120),
	}
}

// Kala Yok numbers label Sunday as 1 through Saturday as 7. time.Weekday
// labels Sunday as 0 through Saturday as 6, so its index is shifted by one.
// A remainder of zero represents Saturday.
func kalaYokWeekday(value int) time.Weekday {
	remainder := value % 7
	return time.Weekday((remainder + 6) % 7)
}
