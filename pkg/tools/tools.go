package tools

import (
	"fmt"
	"math"
	"strconv"
)

const ONE_KB = 1024
const ONE_MB = ONE_KB * 1024
const ONE_GB = ONE_MB * 1024
const ONE_TB = ONE_GB * 1024
const ONE_PB = ONE_TB * 1024

// SizeFormat Converts byte number to correct unit.
func SizeFormat(size int, prc ...int) string {
	digitNumbers := 2
	if len(prc) > 0 {
		if prc[0] > 0 {
			digitNumbers = prc[0]
		}
	}
	size64 := float64(size)
	if size < ONE_KB {
		return fmt.Sprintf("%."+strconv.Itoa(digitNumbers)+"f", toFixed(size64, 0)) + " B"
	} else if size < ONE_MB {
		return fmt.Sprintf("%."+strconv.Itoa(digitNumbers)+"f", toFixed((size64/ONE_KB), 2)) + " KB"
	} else if size < ONE_GB {
		return fmt.Sprintf("%."+strconv.Itoa(digitNumbers)+"f", toFixed((size64/ONE_MB), 2)) + " MB"
	} else if size < ONE_TB {
		return fmt.Sprintf("%."+strconv.Itoa(digitNumbers)+"f", toFixed((size64/ONE_GB), 2)) + " GB"
	} else if size < ONE_PB {
		return fmt.Sprintf("%."+strconv.Itoa(digitNumbers)+"f", toFixed((size64/ONE_TB), 2)) + " TB"
	} else {
		return fmt.Sprintf("%."+strconv.Itoa(digitNumbers)+"f", toFixed((size64/ONE_PB), 2)) + " PB"
	}
}

func toFixed(num float64, n int) float64 {
	pow10 := math.Pow10(n + 1)
	return math.Floor(num*pow10) / pow10
}
