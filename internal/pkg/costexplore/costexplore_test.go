package costexplore

import (
	"testing"
	"time"
)

func TestLookbackMonths(t *testing.T) {
	d1 := LookbackMonths(1, time.Date(2020, time.April, 1, 12, 12, 12, 0, time.UTC))
	if d1.Month() != time.March {
		t.Errorf("Lookback 1 month from April-01 should have returned March")
	}
	d1 = LookbackMonths(1, time.Date(2020, time.January, 1, 12, 12, 12, 0, time.UTC))
	if d1.Month() != time.December {
		t.Errorf("Lookback 1 month from January-01 should have returned December")
	}
}
