package oracle

import "time"

// GetCurrentBlockTime 获取当前区块时间
func GetCurrentBlockTime() time.Time {
	//TODO： default: time.Now()
	return time.Now()
}
