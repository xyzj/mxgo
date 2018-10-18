package mxgo

import (
	"bytes"
	"compress/zlib"
	"container/list"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	// _ "github.com/go-sql-driver/mysql"
)

const (
	// OSNAME from runtime
	OSNAME = runtime.GOOS
	// OSARCH from runtime
	OSARCH = runtime.GOARCH
)

// 字符串数组排序
type StringSliceSort struct {
	OneDimensional []string
	TwoDimensional [][]string
	Idx            int
}

func (arr *StringSliceSort) Len() int {
	if len(arr.OneDimensional) > 0 {
		return len(arr.OneDimensional)
	} else {
		return len(arr.TwoDimensional)
	}
}

func (arr *StringSliceSort) Swap(i, j int) {
	if len(arr.OneDimensional) > 0 {
		arr.OneDimensional[i], arr.OneDimensional[j] = arr.OneDimensional[j], arr.OneDimensional[i]
	} else {
		arr.TwoDimensional[i], arr.TwoDimensional[j] = arr.TwoDimensional[j], arr.TwoDimensional[i]
	}
}

func (arr *StringSliceSort) Less(i, j int) bool {
	if len(arr.OneDimensional) > 0 {
		return arr.OneDimensional[i] < arr.OneDimensional[j]
	} else {
		arr1 := arr.TwoDimensional[i]
		arr2 := arr.TwoDimensional[j]
		if arr.Idx > len(arr.TwoDimensional[0]) {
			arr.Idx = 0
		}
		return arr1[arr.Idx] < arr2[arr.Idx]
	}
}

// Queue queue for go
type Queue struct {
	Q *list.List
}

// Put put data to the end of the queue
func (mq *Queue) Put(value interface{}) {
	mq.Q.PushBack(value)
}

// Get get data from front
func (mq *Queue) Get() interface{} {
	if mq.Q.Len() == 0 {
		return nil
	}
	e := mq.Q.Front()
	if e != nil {
		mq.Q.Remove(e)
		return e.Value
	}
	return nil
}

// Len get queue len
func (mq *Queue) Len() int64 {
	return int64(mq.Q.Len())
}

// Empty check if empty
func (mq *Queue) Empty() bool {
	return mq.Q.Len() == 0
}

// Clean clean the queue
func (mq *Queue) Clean() {
	var n *list.Element
	for e := mq.Q.Front(); e != nil; e = n {
		n = e.Next()
		mq.Q.Remove(e)
	}
}

// GetAddrFromString get addr from config string
//  args:
//	straddr: something like "1,2,3-6"
//  return:
//	[]int64,something like []int64{1,2,3,4,5,6}
func GetAddrFromString(straddr string) ([]int64, error) {
	lst := strings.Split(strings.TrimSpace(straddr), ",")
	lstAddr := make([]int64, 0)
	for _, v := range lst {
		if strings.Contains(v, "-") {
			x := strings.Split(v, "-")
			x1, ex := strconv.ParseInt(strings.TrimSpace(x[0]), 10, 0)
			if ex != nil {
				return nil, ex
			}
			x2, ex := strconv.ParseInt(strings.TrimSpace(x[1]), 10, 0)
			if ex != nil {
				return nil, ex
			}
			for i := x1; i <= x2; i++ {
				lstAddr = append(lstAddr, i)
			}
		} else {
			if y, ex := strconv.ParseInt(strings.TrimSpace(v), 10, 0); ex != nil {
				return nil, ex
			} else {
				lstAddr = append(lstAddr, y)
			}
		}
	}
	return lstAddr, nil
}

// CheckIP check if the ipstring is legal
//  args:
//	ip: ipstring something like 127.0.0.1:10001
//  return:
//	true/false
func CheckIP(ip string) bool {
	regip := `^(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[1-9])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d)\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d)\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d)$`
	regipwithport := `^(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|[1-9])\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d)\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d)\.(1\d{2}|2[0-4]\d|25[0-5]|[1-9]\d|\d):\d{1,5}$`
	if strings.Contains(ip, ":") {
		if a, ex := regexp.Match(regipwithport, []byte(ip)); ex != nil {
			return false
		} else {
			return a
		}
	} else {
		if a, ex := regexp.Match(regip, []byte(ip)); ex != nil {
			return false
		} else {
			return a
		}
	}
}

// MakeRuntimeDirs make conf,log,cache dirs
//  Args：
//	rootpath： 输入路径
//  return：
// 	conf，log，cache三个文件夹的完整路径
func MakeRuntimeDirs(rootpath string) (string, string, string) {
	var basepath string
	if strings.Compare(rootpath, ".") == 0 {
		basepath = GetExecDir()
	} else {
		basepath = rootpath
	}
	os.MkdirAll(filepath.Join(basepath, "..", "conf"), 0775)
	os.MkdirAll(filepath.Join(basepath, "..", "log"), 0775)
	os.MkdirAll(filepath.Join(basepath, "..", "cache"), 0775)
	return filepath.Join(basepath, "..", "conf"), filepath.Join(basepath, "..", "log"), filepath.Join(basepath, "..", "cache")
}

// String2Bytes convert hex-string to []byte
//  args:
// 	data: 输入字符串
// 	sep： 用于分割字符串的分割字符
//  return:
// 	字节切片
func String2Bytes(data string, sep string) []byte {
	var z []byte
	a := strings.Split(data, sep)
	z = make([]byte, len(a))
	for k, v := range a {
		b, _ := strconv.ParseUint(v, 16, 8)
		z[k] = byte(b)
	}
	return z
}

// Bytes2String convert []byte to hex-string
//  args:
// 	data: 输入字节切片
// 	sep： 用于分割字符串的分割字符
//  return:
// 	字符串
func Bytes2String(data []byte, sep string) string {
	a := make([]string, len(data))
	for k, v := range data {
		a[k] = fmt.Sprintf("%02x", v)
	}
	return strings.Join(a, sep)
}

// String2Int convert string 2 int
//  args:
// 	s: 输入字符串
// 	t: 返回数值进制
//  Return：
// 	int
func String2Int(s string, t int) int {
	x, _ := strconv.ParseInt(s, t, 0)
	return int(x)
}

// String2Int8 convert string 2 int8
//  args:
// 	s: 输入字符串
// 	t: 返回数值进制
//  Return：
// 	int8
func String2Int8(s string, t int) byte {
	x, _ := strconv.ParseInt(s, t, 0)
	return byte(x)
}

// String2Int32 convert string 2 int32
//  args:
// 	s: 输入字符串
// 	t: 返回数值进制
//  Return：
// 	int32
func String2Int32(s string, t int) int32 {
	x, _ := strconv.ParseInt(s, t, 0)
	return int32(x)
}

// String2Int64 convert string 2 int64
//  args:
// 	s: 输入字符串
// 	t: 返回数值进制
//  Return：
// 	int64
func String2Int64(s string, t int) int64 {
	x, _ := strconv.ParseInt(s, t, 0)
	return x
}

//StringSlice2Int8 convert string Slice 2 int8
func StringSlice2Int8(bs []string) byte {
	return String2Int8(strings.Join(bs, ""), 2)
}

// CheckLrc check lrc data
func CheckLrc(d []byte) bool {
	rowdata := d[:len(d)-1]
	lrcdata := d[len(d)-1]

	c := CountLrc(&rowdata)
	if c == lrcdata {
		return true
	}
	return false
}

// CountLrc count lrc data
func CountLrc(data *[]byte) byte {
	a := byte(0)
	for _, v := range *data {
		a ^= v
	}
	return a
}

// CheckCrc16VB check crc16 data
func CheckCrc16VB(d []byte) bool {
	rowdata := d[:len(d)-2]
	crcdata := d[len(d)-2:]

	c := CountCrc16VB(&rowdata)
	if c[0] == crcdata[0] && c[1] == crcdata[1] {
		return true
	}
	return false
}

// CountCrc16VB count crc16 as vb6 do
func CountCrc16VB(data *[]byte) []byte {
	var z = make([]byte, 0)
	crc16lo := byte(0xFF)
	crc16hi := byte(0xFF)
	cl := byte(0x01)
	ch := byte(0xa0)
	for _, v := range *data {
		crc16lo ^= v
		for i := 0; i < 8; i++ {
			savehi := crc16hi
			savelo := crc16lo
			crc16hi /= 2
			crc16lo /= 2
			if (savehi & 0x01) == 0x01 {
				crc16lo ^= 0x80
			}
			if (savelo & 0x01) == 0x01 {
				crc16hi ^= ch
				crc16lo ^= cl
			}
		}
	}
	z = append(z, byte(crc16lo), byte(crc16hi))
	return z
}

// IPUint2String change ip int64 data to string format
func IPUint2String(ipnr uint) string {
	return fmt.Sprintf("%d.%d.%d.%d", (ipnr>>24)&0xFF, (ipnr>>16)&0xFF, (ipnr>>8)&0xFF, ipnr&0xFF)
}

// IPInt642String change ip int64 data to string format
func IPInt642String(ipnr int64) string {
	return fmt.Sprintf("%d.%d.%d.%d", (ipnr)&0xFF, (ipnr>>8)&0xFF, (ipnr>>16)&0xFF, (ipnr>>24)&0xFF)
}

// IPInt642Bytes change ip int64 data to string format
func IPInt642Bytes(ipnr int64) []byte {
	return []byte{byte((ipnr) & 0xFF), byte((ipnr >> 8) & 0xFF), byte((ipnr >> 16) & 0xFF), byte((ipnr >> 24) & 0xFF)}
}

// IPUint2Bytes change ip int64 data to string format
func IPUint2Bytes(ipnr int64) []byte {
	return []byte{byte((ipnr >> 24) & 0xFF), byte((ipnr >> 16) & 0xFF), byte((ipnr >> 8) & 0xFF), byte((ipnr) & 0xFF)}
}

// IP2Uint change ip string data to int64 format
func IP2Uint(ipnr string) uint {
	// ex := errors.New("wrong ip address format")
	bits := strings.Split(ipnr, ".")
	if len(bits) != 4 {
		return 0
	}
	var intip uint
	for k, v := range bits {
		i, ex := strconv.Atoi(v)
		if ex != nil || i > 255 || i < 0 {
			return 0
		}
		intip += uint(i) << uint(8*(3-k))
	}
	return intip
}

// IP2Int64 change ip string data to int64 format
func IP2Int64(ipnr string) int64 {
	// ex := errors.New("wrong ip address format"
	bits := strings.Split(ipnr, ".")
	if len(bits) != 4 {
		return 0
	}
	var intip uint
	for k, v := range bits {
		i, ex := strconv.Atoi(v)
		if ex != nil || i > 255 || i < 0 {
			return 0
		}
		intip += uint(i) << uint(8*(k))
	}
	return int64(intip)
}

// IsExist file is exist or not
func IsExist(p string) bool {
	_, err := os.Stat(p)
	return err == nil || os.IsExist(err)
}

// GetExecDir get current file path
func GetExecDir() string {
	a, _ := os.Executable()
	execdir := filepath.Dir(a)
	if strings.Contains(execdir, "go-build") {
		execdir, _ = filepath.Abs(".")
	}
	return execdir
}

//SplitDateTime SplitDateTime
func SplitDateTime(sdt int64) (y, m, d, h, mm, s, wd byte) {
	if sdt == 0 {
		sdt = time.Now().Unix()
	}
	if sdt > 621356256000000000 {
		sdt = SwitchStamp(sdt)
	}
	tm := time.Unix(sdt, 0)
	stm := tm.Format("2006-01-02 15:04:05 Mon")
	dt := strings.Split(stm, " ")
	dd := strings.Split(dt[0], "-")
	tt := strings.Split(dt[1], ":")
	return byte(String2Int32(dd[0], 10) - 2000),
		String2Int8(dd[1], 10),
		String2Int8(dd[2], 10),
		String2Int8(tt[0], 10),
		String2Int8(tt[1], 10),
		String2Int8(tt[2], 10),
		byte(tm.Weekday())
}

// Stamp2Time convert stamp to datetime string
func Stamp2Time(t int64) string {
	tm := time.Unix(t, 0)
	return tm.Format("2006-01-02 15:04:05")
}

// Time2Stampf 可根据制定的时间格式和时区转换为当前时区的Unix时间戳
//  fmt：
//  year：2006
//  month：01
//  day：02
//  hour：15
//  minute：04
//  second：05
//  tz：0～12,超范围时使用本地时区
func Time2Stampf(s, fmt string, tz float32) int64 {
	if fmt == "" {
		fmt = "2006-01-02 15:04:05"
	}
	if tz > 12 || tz < 0 {
		_, t := time.Now().Zone()
		tz = float32(t / 3600)
	}
	var loc *time.Location
	loc = time.FixedZone("", int((time.Duration(tz) * time.Hour).Seconds()))
	tm, ex := time.ParseInLocation(fmt, s, loc)
	if ex != nil {
		return 0
	}
	return tm.Unix()
}

// Time2Stamp convert datetime string to stamp
func Time2Stamp(s string) int64 {
	return Time2Stampf(s, "", 8)
}

// Time2StampNB 电信NB平台数据时间戳转为本地unix时间戳
func Time2StampNB(s string) int64 {
	return Time2Stampf(s, "20060102T150405Z", 0)
}

// SwitchStamp switch stamp format between unix and c#
func SwitchStamp(t int64) int64 {
	y := int64(621356256000000000)
	z := int64(10000000)
	if t > y {
		return (t - y) / z
	}
	return t*z + y
}

// Byte2Bytes int8 to bytes
func Byte2Bytes(v byte, reverse bool) []byte {
	s := fmt.Sprintf("%08b", v)
	if reverse {
		s = ReverseString(s)
	}
	b := make([]byte, 0)
	for _, v := range s {
		if v == 48 {
			b = append(b, 0)
		} else {
			b = append(b, 1)
		}
	}
	return b
}

// Byte2Int32s int8 to int32 list
func Byte2Int32s(v byte, reverse bool) []int32 {
	s := fmt.Sprintf("%08b", v)
	if reverse {
		s = ReverseString(s)
	}
	b := make([]int32, 0)
	for _, v := range s {
		if v == 48 {
			b = append(b, 0)
		} else {
			b = append(b, 1)
		}
	}
	return b
}

// Bcd2Int8 bcd to int
func Bcd2Int8(v byte) byte {
	return ((v&0xf0)>>4)*10 + (v & 0x0f)
}

// Int82Bcd int to bcd
func Int82Bcd(v byte) byte {
	return ((v / 10) << 4) | (v % 10)
}

// ReverseString ReverseString
func ReverseString(s string) string {
	runes := []rune(s)
	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}
	return string(runes)
}

// DecodeString 解码混淆字符串，兼容python算法
func DecodeString(s string) string {
	s = SwapCase(s)
	var ns bytes.Buffer
	ns.Write([]byte{120, 156})
	if x := 4 - len(s)%4; x != 4 {
		for i := 0; i < x; i++ {
			s += "="
		}
	}
	if y, ex := base64.StdEncoding.DecodeString(s); ex == nil {
		x := String2Int8(string(y[0])+string(y[1]), 0)
		z := y[2:]
		for i := len(z) - 1; i >= 0; i-- {
			if z[i] >= x {
				ns.WriteByte(z[i] - x)
			} else {
				ns.WriteByte(byte(int(z[i]) + 256 - int(x)))
			}
		}
		return ReverseString(string(DoZlibUnCompress(ns.Bytes())))
	} else {
		return "You screwed up."
	}
}

// DoZlibUnCompress zlib uncompress
func DoZlibUnCompress(src []byte) []byte {
	b := bytes.NewReader(src)
	var out bytes.Buffer
	r, _ := zlib.NewReader(b)
	io.Copy(&out, r)
	return out.Bytes()
}

// DoZlibCompress zlib compress
func DoZlibCompress(src []byte) []byte {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	w.Write(src)
	w.Close()
	return in.Bytes()
}

// SwapCase swap char case
func SwapCase(s string) string {
	var ns bytes.Buffer
	for _, v := range s {
		// println(v, string(v))
		if v >= 65 && v <= 90 {
			ns.WriteString(string(int(v) + 32))
		} else if v >= 97 && v <= 122 {
			ns.WriteString(string(int(v) - 32))
		} else {
			ns.WriteString(string(v))
		}
	}
	return ns.String()
}

// VersionInfo show something
//  args:
// 	p: program name
// 	v: program version
// 	gv: golang version
// 	bd: build datetime
// 	pl: platform info
// 	a: auth name
func VersionInfo(p, v, gv, bd, pl, a string) string {
	return fmt.Sprintf("\n%s\r\nVersion:\t%s\r\nGo version:\t%s\r\nBuild date:\t%s\r\nBuild OS:\t%s\r\nCode by:\t%s", p, v, gv, pl, bd, a)
}

// WriteVersionInfo write version info to .ver file
//  args:
// 	p: program name
// 	v: program version
// 	gv: golang version
// 	bd: build datetime
// 	pl: platform info
// 	a: auth name
func WriteVersionInfo(p, v, gv, bd, pl, a string) {
	fn, _ := os.Executable()
	f, _ := os.OpenFile(fmt.Sprintf("%s.ver", fn), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0444)
	defer f.Close()
	f.WriteString(fmt.Sprintf("\n%s\r\nVersion:\t%s\r\nGo version:\t%s\r\nBuild date:\t%s\r\nBuild OS:\t%s\r\nCode by:\t%s\r\n", p, v, gv, pl, bd, a))
}

// CalculateSecurityCode calculate security code
//  args:
//	t: calculate type "h"-按小时计算，当分钟数在偏移值范围内时，同时计算前后一小时的值，"m"-按分钟计算，同时计算前后偏移量范围内的值
//	salt: 拼接用字符串
//	offset: 偏移值，范围0～59
//  return:
//	32位小写md5码切片
func CalculateSecurityCode(t, salt string, offset int) []string {
	var sc []string
	if offset < 0 {
		offset = 0
	}
	if offset > 59 {
		offset = 59
	}
	tt := time.Now()
	mm := tt.Minute()
	switch t {
	case "h":
		sc = make([]string, 0, 3)
		sc = append(sc, GetMD5(tt.Format("2006010215")+salt))
		if mm < offset || 60-mm < offset {
			sc = append(sc, GetMD5(tt.Add(-1*time.Hour).Format("2006010215")+salt))
			sc = append(sc, GetMD5(tt.Add(time.Hour).Format("2006010215")+salt))
		}
	case "m":
		sc = make([]string, 0, offset*2)
		if offset > 0 {
			tts := tt.Add(time.Duration(-1*(offset)) * time.Minute)
			for i := 0; i < offset*2+1; i++ {
				sc = append(sc, GetMD5(tts.Add(time.Duration(i)*time.Minute).Format("200601021504")+salt))
			}
		} else {
			sc = append(sc, GetMD5(tt.Format("200601021504")+salt))
		}
	}
	return sc
}

// GetMD5 生成32位md5字符串
func GetMD5(text string) string {
	ctx := md5.New()
	ctx.Write([]byte(text))
	return hex.EncodeToString(ctx.Sum(nil))
}

// GetRandomString 生成随机字符串
func GetRandomString(l int64) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*()_+[]{};:<>,./?-="
	bb := []byte(str)
	var rs bytes.Buffer
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := int64(0); i < l; i++ {
		rs.WriteByte(bb[r.Intn(len(bb))])
	}
	return rs.String()
}

// CheckSQLInject 检查sql语句是否包含注入攻击
func CheckSQLInject(s string) bool {
	str := `(?:')|(?:--)|(/\\*(?:.|[\\n\\r])*?\\*/)|(\b(select|update|and|or|delete|insert|trancate|char|chr|into|substr|ascii|declare|exec|count|master|into|drop|execute)\b)`
	re, err := regexp.Compile(str)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

// Page404 一个简单的404页面
func Page404(rw http.ResponseWriter, r *http.Request) {
	var d = make(map[string]string, 0)
	d["img"] = "iVBORw0KGgoAAAANSUhEUgAAAYAAAADqCAYAAACr35I3AAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAyBpVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADw/eHBhY2tldCBiZWdpbj0i77u/IiBpZD0iVzVNME1wQ2VoaUh6cmVTek5UY3prYzlkIj8+IDx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IkFkb2JlIFhNUCBDb3JlIDUuMC1jMDYwIDYxLjEzNDc3NywgMjAxMC8wMi8xMi0xNzozMjowMCAgICAgICAgIj4gPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4gPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIgeG1sbnM6eG1wPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvIiB4bWxuczp4bXBNTT0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wL21tLyIgeG1sbnM6c3RSZWY9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9zVHlwZS9SZXNvdXJjZVJlZiMiIHhtcDpDcmVhdG9yVG9vbD0iQWRvYmUgUGhvdG9zaG9wIENTNSBXaW5kb3dzIiB4bXBNTTpJbnN0YW5jZUlEPSJ4bXAuaWlkOkE1QkYwM0NCMjY1NzExRTFBRDNDOEMzNUIyODU0MkUzIiB4bXBNTTpEb2N1bWVudElEPSJ4bXAuZGlkOkE1QkYwM0NDMjY1NzExRTFBRDNDOEMzNUIyODU0MkUzIj4gPHhtcE1NOkRlcml2ZWRGcm9tIHN0UmVmOmluc3RhbmNlSUQ9InhtcC5paWQ6QTVCRjAzQzkyNjU3MTFFMUFEM0M4QzM1QjI4NTQyRTMiIHN0UmVmOmRvY3VtZW50SUQ9InhtcC5kaWQ6QTVCRjAzQ0EyNjU3MTFFMUFEM0M4QzM1QjI4NTQyRTMiLz4gPC9yZGY6RGVzY3JpcHRpb24+IDwvcmRmOlJERj4gPC94OnhtcG1ldGE+IDw/eHBhY2tldCBlbmQ9InIiPz5Ss/0wAACHl0lEQVR42uxdB5wUVdLv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"
	t := template.New("my 404")
	t.Parse(`<html lang="zh-cn">
	    <head>
	      <style type="text/css">
	      .img_frame .center_tool{
	          display: inline-block;
	          vertical-align: middle;
	          height: 100% ;
	          width: 0px ;
	          border: none ;
	          padding: 0px ;
	          margin: 0px 0px 0px -0.8rem ;
	      }
	      .img_frame{
	          width: 100% ;
	          height:100%;
	          overflow: hidden ;
	          white-space: nowrap;
	      }
	      .img_frame .image_large{
	          max-width:100% ;
	          max-height:100% ;
	          vertical-align: middle;
	          display: inline-block ;
	          margin-left: 0.3rem;
	      }
	      </style>

	    <meta content="text/html; charset=utf-8" http-equiv="content-type" />
	    <title>Wrong Way!</title>
	    </head>
	    <body style="background: #67ACE4;">
	      <div style="text-align: center;" class="img_frame" id="img_frame">
	          <div class="center_tool"></div>
			  <a href="/">
	          	<img class="image_large" src="data:image/png;base64,{{.img}}">
			  </a>
	      </div>
	    </body>
	</html>
	`)
	t.Execute(rw, d)
}

// GetSqlConn 获取数据库连接实例，utf8字符集，连接超时10s
// username: 数据库连接用户名
// password： 数据库连接密码
// host：主机名/主机ip
// port：服务端口号，默认3306
// dbname：数据库名称，为空时表示不指定数据库
// maxOpenConns：连接池中最大连接数，有效范围1-200，超范围时强制为20
// multiStatements：允许执行多条语句，true or false
// readTimeout：I/O操作超时时间，单位秒，0-无超时
// func GetSqlConn(username, password, host, dbname string, port, maxOpenConns int, multiStatements bool, readTimeout uint32) (*sql.DB, error) {
// 	ms := "false"
// 	if multiStatements {
// 		ms = "true"
// 	}
// 	if port > 65535 || port < 1 {
// 		port = 3306
// 	}
// 	connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s"+
// 		"?multiStatements=%s"+
// 		"&readTimeout=%ds"+
// 		"&parseTime=true"+
// 		"&timeout=10s"+
// 		"&charset=utf8"+
// 		"&columnsWithAlias=true",
// 		username, password, host, port, dbname, ms, readTimeout)
// 	db, ex := sql.Open("mysql", strings.Replace(connString, "\n", "", -1))
//
// 	if ex != nil {
// 		return nil, ex
// 	}
//
// 	if maxOpenConns <= 0 || maxOpenConns > 200 {
// 		maxOpenConns = 20
// 	}
// 	if maxOpenConns < 2 {
// 		db.SetMaxIdleConns(maxOpenConns)
// 	} else {
// 		db.SetMaxIdleConns(maxOpenConns / 2)
// 	}
// 	db.SetMaxOpenConns(maxOpenConns)
//
// 	if ex := db.Ping(); ex != nil {
// 		return nil, ex
// 	}
// 	return db, nil
// }
