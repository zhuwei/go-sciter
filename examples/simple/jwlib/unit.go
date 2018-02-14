package jwlib

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	base64Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
)

var AppDebug string

var mutexLog sync.Mutex
var mutexHttpLog sync.Mutex
var mutexSqlLog sync.Mutex

var coder = base64.NewEncoding(base64Table)

func base64Encode(src []byte) []byte {
	return []byte(coder.EncodeToString(src))
}

func base64Decode(src []byte) ([]byte, error) {
	return coder.DecodeString(string(src))
}

func Mcrypt(data string, mode string) string {

	var passcrypt string
	key := "justware2013"
	if mode == "decode" {
		if len(data) < 3 {
			return ""
		}
		if data[0:2] != getOrd(data[2:]+key) {
			return ""
		}
		str := strings.Replace(data[2:], " ", "+", -1)
		str = strings.Replace(str, "-_", "+/", -1)

		enbyte, err := base64Decode([]byte(str))
		if err != nil {
			passcrypt = ""
		} else {
			passcrypt = string(enbyte)
		}
	}

	if mode == "encode" {
		enbyte := base64Encode([]byte(data))

		passcrypt = string(enbyte)
		passcrypt = strings.Replace(passcrypt, "+/", "-_", -1)
		passcrypt = getOrd(passcrypt+key) + passcrypt
	}
	return passcrypt
}

func getOrd(str string) string {
	sum := 0
	if len(str) == 0 {
		return "00"
	} else {
		sum = int(str[0])
		var i64 int64
		i64 = int64(sum % 265)
		ret := strconv.FormatInt(i64, 16)
		if len(ret) == 1 {
			return "0" + ret
		} else {
			return ret
		}
	}
}

func ReturnJson(ret map[string]interface{}) string {
	b, _ := json.Marshal(ret)
	return Mcrypt(string(b), "encode")
}

//校验字符串是否整数
func IsInt(str string) bool {

	if str == "" {
		return false
	}

	reg := regexp.MustCompile(`[^\d]`)

	if len(reg.FindAllString(str, -1)) > 0 {
		return false
	} else {
		return true
	}
}

/**
 * 单引号替换
 */
func FormatSql(str string) string {
	return strings.Replace(str, "'", "''", -1)
}

func FormatTimeToStr(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.000000000")
}

func printT(t time.Time) time.Time {
	return t.Add(100)
}

//土方法判断进程是否启动
func AppInit() {
	iManPid := fmt.Sprint(os.Getpid())
	tmpDir := os.TempDir()
	fmt.Println(tmpDir)
	if err := ProcExsit(tmpDir); err == nil {
		pidFile, _ := os.Create(tmpDir + "/eposPack.pid")
		defer pidFile.Close()
		pidFile.WriteString(iManPid)
	} else {
		fmt.Println(err)
		os.Exit(1)
	}
}

// 判断进程是否启动
func ProcExsit(tmpDir string) (err error) {
	iManPidFile, err := os.Open(tmpDir + "/eposPack.pid")
	defer iManPidFile.Close()
	if err == nil {
		filePid, err := ioutil.ReadAll(iManPidFile)
		if err == nil {
			pidStr := fmt.Sprintf("%s", filePid)
			pid, _ := strconv.Atoi(pidStr)
			_, err := os.FindProcess(pid)
			if err == nil {
				return errors.New("[ERROR] epos已启动.")
			}
		}
	}
	return nil
}

/**
 * 判断文件是否存在  存在返回 true 不存在返回false
 */
func CheckFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

/**
 * 判断文件夹是否存在  存在返回 true 不存在返回false
 */
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//获取程序路径
func GetAppPath() string {
	//os.Getwd()
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
		return ""
	} else {
		return dir
	}
}

//http log
func WriteHttpLog(r *http.Request) {

	mutexHttpLog.Lock()
	defer mutexHttpLog.Unlock()

	dir := GetAppPath()
	if dir == "" {
		return
	}

	logPath := dir + "/log/"
	if ok, err := PathExists(logPath); !ok {
		err = os.Mkdir(logPath, os.ModePerm) //在当前目录下生成md目录
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	now := time.Now()
	logFile := "http_" + now.Format("200601") + ".log"

	fd, err := os.OpenFile(logPath+logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		fmt.Println(err)
		return
	}

	str := `[` + now.Format("2006-01-02 15:04:05.0000") + `] ` + r.RemoteAddr + ` - "` + r.Method + ` ` + r.RequestURI + ` ` + r.Proto + `"` + "\n"
	buf := []byte(str)
	fd.Write(buf)
	fd.Close()

	fmt.Print(str)
	/*fmt.Println(r.Host)        //121.41.48.197:9090
	fmt.Println(r.Method)      //POST
	fmt.Println(r.RemoteAddr)  //59.172.170.173:19625
	fmt.Println(r.RequestURI)  // /
	fmt.Println(r.URL)         // /
	fmt.Println(r.UserAgent()) //
	fmt.Println(r.Referer())   //
	fmt.Println(r.Proto)       // HTTP/1.1
	*/
}

//log
func WriteLog(s string) {
	WriteLogTag(``, s)
}

func WriteLogTag(tag string, s string) {

	mutexLog.Lock()
	defer mutexLog.Unlock()

	dir := GetAppPath()
	if dir == "" {
		return
	}

	logPath := dir + "/log/"
	if ok, err := PathExists(logPath); !ok {
		err = os.Mkdir(logPath, os.ModePerm) //在当前目录下生成md目录
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	var logFile string
	now := time.Now()
	if len(tag) > 0 {
		logFile = tag + `_` + now.Format("200601") + ".log"
	} else {
		logFile = now.Format("200601") + ".log"
	}

	fd, err := os.OpenFile(logPath+logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		fmt.Println(err)
		return
	}

	var str string
	if len(tag) > 0 {
		str = `[` + now.Format("2006-01-02 15:04:05.0000") + `] [` + tag + `]` + s + "\n"
	} else {
		str = `[` + now.Format("2006-01-02 15:04:05.0000") + `] ` + s + "\n"
	}

	buf := []byte(str)
	fd.Write(buf)
	fd.Close()

	fmt.Print(str)
}

//sql log
func WriteSqlLog(s string) {

	mutexSqlLog.Lock()
	defer mutexSqlLog.Unlock()

	dir := GetAppPath()
	if dir == "" {
		return
	}

	logPath := dir + "/log/"
	if ok, err := PathExists(logPath); !ok {
		err = os.Mkdir(logPath, os.ModePerm) //在当前目录下生成md目录
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	now := time.Now()
	logFile := "sql_" + now.Format("200601") + ".log"

	fd, err := os.OpenFile(logPath+logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		fmt.Print(err)
		return
	}

	str := `[` + now.Format("2006-01-02 15:04:05.0000") + `] ` + s + "\n"
	buf := []byte(str)
	fd.Write(buf)
	fd.Close()
}

//err := SendToMail("yang**@yun*.com", "***", "smtp.exmail.qq.com:25", "397685131@qq.com", subject, body, "html")
func SendToMail(user, password, host, to, subject, body, mailtype string) error {
	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", user, password, hp[0])
	var content_type string
	if mailtype == "html" {
		content_type = "Content-Type: text/" + mailtype + "; charset=UTF-8"
	} else {
		content_type = "Content-Type: text/plain" + "; charset=UTF-8"
	}

	msg := []byte("To: " + to + "\r\nFrom: " + user + ">\r\nSubject: " + "\r\n" + content_type + "\r\n\r\n" + body)
	send_to := strings.Split(to, ";")
	err := smtp.SendMail(host, auth, user, send_to, msg)
	return err
}

//获取上一级函数名称等信息
func GetCurrFuncInfo() string {

	pc, _, line, ok := runtime.Caller(1)
	if ok {
		f := runtime.FuncForPC(pc)
		return `(Func: ` + f.Name() + `;line: ` + strconv.Itoa(line) + `;)`
	} else {
		return ""
	}
}

func DownloadImg(url string, filename string) (int64, string) {

	if len(url) <= 0 {
		return 0, ``
	}

	var resultErr error

	//for i := 0; i < 20; i++ {
	i := 0
	for true {
		if i > 0 {
			WriteLog(`请求(` + url + `)失败, 重试中...(` + strconv.Itoa(i) + `)`)
			time.Sleep(200 * time.Millisecond)
		} else {
			i++
		}
		client := &http.Client{}
		reqest, err := http.NewRequest("GET", url, nil)

		if err != nil {
			resultErr = err
			continue
		}

		reqest.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
		reqest.Header.Add("Accept-Encoding", "gzip, deflate")
		reqest.Header.Add("Accept-Language", "zh-CN,zh;q=0.8")
		reqest.Header.Add("Cache-Control", "max-age=0")
		reqest.Header.Add("Connection", "keep-alive")
		//reqest.Header.Add("Host", "www.dianping.com")
		//reqest.Header.Add("Referer", "http://www.dianping.com/wuhan")
		reqest.Header.Add("Upgrade-Insecure-Requests", "1")
		reqest.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36")
		response, err := client.Do(reqest)
		if err != nil {
			resultErr = err
			continue
		}

		if response.StatusCode == 200 || response.StatusCode == 304 {

			defer response.Body.Close()

			var bodyByte []byte
			ext := `jpg`

			if response.Header.Get("Content-Type") == "image/jpeg" {
				ext = `jpg`
				bodyByte, _ = ioutil.ReadAll(response.Body)
			} else if response.Header.Get("Content-Type") == "image/png" {
				ext = `png`
				bodyByte, _ = ioutil.ReadAll(response.Body)
			}

			out, _ := os.Create(filename + `.` + ext)
			defer out.Close()

			n, _ := io.Copy(out, bytes.NewReader(bodyByte))

			return n, filename + `.` + ext
		} else {
			response.Body.Close()
		}
	}

	/*suffixes := "avi|mpeg|3gp|mp3|mp4|wav|jpeg|gif|jpg|png|apk|exe|pdf|rar|zip|docx|doc"

	reg, _ := regexp.Compile(`(\w|\d|_)*.(` + suffixes + `)`)
	name := reg.FindStringSubmatch(url)[0]
	if name == "" {
		return 0
	}

	ext := strings.Split(name, ".")[1]

	index := strings.Index(url, name)

	//通过http请求获取图片的流文件
	resp, _ := http.Get(url[0:index] + name)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	out, _ := os.Create(filename + `.` + ext)
	defer out.Close()

	n, _ := io.Copy(out, bytes.NewReader(body))*/

	if resultErr != nil {
		WriteLog(`[DownloadImg]` + resultErr.Error())
	}

	return 0, ``
}

func GetPefixUrl(url string) string {

	suffixes := "avi|mpeg|3gp|mp3|mp4|wav|jpeg|gif|jpg|png|apk|exe|pdf|rar|zip|docx|doc"

	reg, _ := regexp.Compile(`(\w|\d|_)*.(` + suffixes + `)`)
	list := reg.FindStringSubmatch(url)
	if len(list) <= 0 {
		return GetPefixUrl2(url)
	}

	name := list[0]
	if name == "" {
		return GetPefixUrl2(url)
	}

	index := strings.Index(url, name)
	return url[0:index] + name
}

func GetPefixUrl2(url string) string {
	index := strings.Index(url, "%40")
	if index >= 0 {
		return url[0:index]
	} else {
		return url
	}
}

// 写超时警告日志 通用方法
func TimeoutWarning(tag, detailed string, start time.Time, timeLimit float64) {
	dis := time.Now().Sub(start).Seconds()
	if dis > timeLimit {
		//log.Warning(log.CENTER_COMMON_WARNING, tag, " detailed:", detailed, "TimeoutWarning using", dis, "s")
		//pubstr := fmt.Sprintf("%s count %v, using %f seconds", tag, count, dis)
		//stats.Publish(tag, pubstr)
		pubstr := fmt.Sprintf("%s(%s), using %f seconds", tag, detailed, dis)
		WriteLog(pubstr)
	}
}

func GetAddressLocation(region string, address string) (float64, float64, float64, error) {

	str := `http://apis.map.qq.com/ws/place/v1/search?`

	parameters := url.Values{}
	parameters.Add("keyword", address)
	parameters.Add("boundary", `region(`+region+`,0)`)
	parameters.Add("key", "LERBZ-VHXH6-TD5SB-M5P6M-YOG2Q-CFFVT")

	str += parameters.Encode()

	resp, err := http.Get(str)
	if err != nil {
		return 0, 0, -1, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, -1, errors.New("http返回statusCode:" + strconv.Itoa(resp.StatusCode))
	}
	body, err := ioutil.ReadAll(resp.Body)

	var dataMap map[string]interface{}
	err = json.Unmarshal(body, &dataMap)
	if err != nil {
		return 0, 0, -1, errors.New("返回数据JSON解析失败!")
	}

	var status float64 = -1
	if value, ok := dataMap["status"]; ok {
		status = value.(float64)
	}

	var message string = ``
	if value, ok := dataMap["message"]; ok {
		message = value.(string)
	}

	if status != 0 {
		return 0, 0, status, errors.New(message)
	}

	var data []interface{}
	if value, ok := dataMap["data"]; ok {
		data = value.([]interface{})
	} else {
		return 0, 0, status, errors.New("无法查询该地址(" + address + ")")
	}

	if len(data) == 0 {
		return 0, 0, status, errors.New("无法查询该地址(" + address + ")")
	}

	item := data[0].(map[string]interface{})

	var location map[string]interface{}
	if value, ok := item["location"]; ok {
		location = value.(map[string]interface{})
	} else {
		return 0, 0, status, errors.New("该地址没有经纬度坐标(" + address + ")")
	}

	var lat float64 = 0
	if value, ok := location["lat"]; ok {
		lat = value.(float64)
	} else {
		return 0, 0, status, errors.New("无法查询该地址lat(" + address + ")")
	}

	var lng float64 = 0
	if value, ok := location["lng"]; ok {
		lng = value.(float64)
	} else {
		return 0, 0, status, errors.New("无法查询该地址lng(" + address + ")")
	}

	return lat, lng, status, nil
}

func ReplaceSpecialChar(src string) string {

	isHave := false
	specialChar := `\/:*?"<>|` //特殊字符 windows文件名
	for i := 0; i < len(src); i++ {
		if strings.Contains(specialChar, src[i:i+1]) {
			isHave = true
			break
		}
	}

	newSrc := src
	if isHave {
		for i := 0; i < len(specialChar); i++ {
			newSrc = strings.Replace(newSrc, specialChar[i:i+1], `_`, -1)
		}
	}

	return newSrc
}

//左边填充字符串
func PadLeft(s string, l int, c string) string {
	return PadString(s, l, true, c)
}

//右边填充字符串
func PadRight(s string, l int, c string) string {
	return PadString(s, l, false, c)
}

func PadString(s string, l int, f bool, c string) string {

	var result string

	for i := 0; i < l-len(s); i++ {
		result = result + c
	}

	if f {
		return result + s
	} else {
		return s + result
	}
}

func Md5File(path string) (string, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return "", err
	}

	h := md5.New()
	_, err = io.Copy(h, file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func SHA1File(path string) (string, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return "", err
	}

	h := sha1.New()
	_, err = io.Copy(h, file)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func SHA256File(path string) (string, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return "", err
	}

	h := sha256.New()
	_, err = io.Copy(h, file)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
