package id

import (
	"crypto/md5"
	"fmt"
	"io"
	"math/rand"
	"time"

	"strconv"

	"strings"

	"github.com/gofrs/uuid"
)

var timeLayout = "2006-01-02-15-04-05"

// GenReadableUint64ID generate readable uint64 id 19(bit)
func GenReadableUint64ID() uint64 {

	now := time.Now().UTC()
	formattedNow := now.Format(timeLayout)

	idStr := strings.ReplaceAll(formattedNow, "-", "") + fmt.Sprintf("%03d", now.Nanosecond()/1000000) + fmt.Sprintf("%02d", rand.Intn(100))

	id, _ := strconv.ParseUint(idStr, 10, 64)

	return id
}

// GenUint64ID generate normal uint64 id 19(bit)
func GenUint64ID() uint64 {
	now := time.Now().UTC().UnixNano()
	ran := rand.Intn(1000)

	idNum := now + int64(ran)

	return uint64(idNum)
}

// GenTraceID new normal traceID
func GenTraceID() string {
	return GenUUIDString()
}

// TraceIDFrom new traceID from text
func TraceIDFrom(text string) string {
	return UUIDFromString(text)
}

// GenUUIDString new uuid
func GenUUIDString() string {
	return uuid.Must(uuid.NewV4()).String()
}

// Num2Str convert uint64 to number string
func Num2Str(id uint64) string {
	return strconv.FormatUint(id, 10)
}

// Str2Num convert number string to uint64
func Str2Num(idStr string) uint64 {
	v, _ := strconv.ParseUint(idStr, 10, 64)
	return v
}

// UUIDByName new uuid string from name
func UUIDByName(uuidStr, name string) string {
	ns, e := uuid.FromString(uuidStr)
	if e != nil {
		panic(e)
	}

	return uuid.NewV5(ns, name).String()
}

// UUIDFromString  new uuid string from string
func UUIDFromString(text string) string {
	h := md5.New()
	io.WriteString(h, text)
	sum := h.Sum(nil)
	sum[6] = (sum[6] & 0x0f) | 0x30
	sum[8] = (sum[8] & 0x3f) | 0x80
	// fmt.Printf(":::%x\n", sum)
	return uuid.FromBytesOrNil(sum).String()
}
