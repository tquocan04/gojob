package pkg

import "time"

const DateTimeLayout = "2006-01-02 15:04:05"

var LocVN *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		// Fallback to FixedZone GMT+7 if docker/scratch environment doen not have tzdat
		LocVN = time.FixedZone("ICT", 7*60*60)
		return
	}

	LocVN = loc
}

func ToVN(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}

	return t.In(LocVN)
}

func FormatVN(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.In(LocVN).Format(DateTimeLayout)
}

// Handle pointer *time.Time case (e.g. LockedAt can be null)
func FormatVNPtr(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	formatted := t.In(LocVN).Format(DateTimeLayout)
	return &formatted
}
