package costexplore

import (
	"time"
)

func LookbackMonths(lbmonths int, anchorTime time.Time) time.Time {
	endDate := anchorTime.AddDate(0, -1*lbmonths, 0)
	return endDate
}
