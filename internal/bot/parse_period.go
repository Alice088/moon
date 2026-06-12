package bot

import (
	"strconv"
	"strings"
	"time"
)

func parsePeriod(s string) (time.Duration, error) {
	switch s {
	case "hour":
		return time.Hour, nil
	case "day":
		return 24 * time.Hour, nil
	case "week":
		return 7 * 24 * time.Hour, nil
	case "month":
		return 30 * 24 * time.Hour, nil
	}

	if strings.HasSuffix(s, "d") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return 0, err
		}
		return time.Duration(n) * 24 * time.Hour, nil
	}

	return time.ParseDuration(s)
}
