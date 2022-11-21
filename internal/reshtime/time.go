package reshtime

import (
	"fmt"
	"math"
	"time"
)

func TimeToReshString(t time.Time) string {
	return fmt.Sprintf("%.4f", float64(t.Unix())+(float64(t.Nanosecond())*math.Pow(10, -9)))
}
