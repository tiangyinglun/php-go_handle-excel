package tools

import (
	"os"
	"fmt"
	"bufio"
	"io"
	"strings"
	"time"
	"strconv"
	"encoding/json"
	"math/rand"
	"crypto/md5"
)

/**
获取文件内容
*/
func readCluesFile(file string) (dataBox [][]string, err error) {
	f, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
	}
	var box [][]string
	defer f.Close()
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		if err != nil || io.EOF == err { //遇到任何错误立即返回，并忽略 EOF 错误信息
			break
		}
		m := strings.Split(strings.TrimRight(line, "\n"), "\t")
		if m[0] == "" {
			continue
		}
		box = append(box, m)
	}
	return box, err
}

func readCluesFileLine(file string) (s [][]string, err error) {
	f, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
	}
	var box [][]string
	defer f.Close()
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		if err != nil || io.EOF == err { //遇到任何错误立即返回，并忽略 EOF 错误信息
			break
		}
		m := strings.Split(strings.TrimRight(line, "\n"), "\t")
		if len(m) <= 0 { //可以删除的
			continue
		}
		box = append(box, m)
	}
	return box, err
}

//创建日期目录
func CreateDir(path string, types bool) (string, bool) {
	if types {
		return "", true
	}
	year, month, day := time.Now().Date()
	years := fmt.Sprintf("%d", year)
	var monthLong string
	var dayLong string
	monthLong = fmt.Sprintf("%d", month)
	if len(monthLong) < 2 {
		monthLong = "0" + monthLong
	}
	dayLong = fmt.Sprintf("%d", day)
	if len(dayLong) < 2 {
		dayLong = "0" + dayLong
	}
	dirName := years + monthLong + dayLong
	if Exist(path + dirName) {
		return dirName + "/", true
	}
	err := os.Mkdir(path + dirName + "/", 0777)
	if err != nil {
		fmt.Println(err)
		return dirName + "/", false
	}
	return dirName + "/", true
}

//check json的内容是否合法

//验证后缀
func checkExtension(path, Suffix string) bool {
	return strings.HasSuffix(path, Suffix)
}

//检测文件是否存在
func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

//生成数据数
func RandNum(num int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return strconv.Itoa(r.Intn(num))
}

//把 c 转成json
func (c *CallBack) RanderJson() (jsonStr string, err error) {
	jsonS, err := json.Marshal(c.RBack)
	if err != nil {
		fmt.Println(err)
	}
	jsonStr = string(jsonS)
	return
}

/**
MD5加密
 */
func Md5(str string) string {
	data := []byte(str)
	return fmt.Sprintf("%x", md5.Sum(data))
}


/**
兴建数据
 */
func CreateFile(file, data string) (n int, err error) {

	f_h, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		fmt.Println(err)
	}
	n, err = io.WriteString(f_h, data)
	if err != nil {
		fmt.Println(err)
	}
	defer f_h.Close()
	return n, err
}
