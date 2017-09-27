package tools

import (
	"bufio"
	"code.google.com/p/mahonia"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/extrame/xls"
	"github.com/tealeg/xlsx"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"path"
)

// //处理数据
func HandleData(types int32, paramMap map[string]string, c *CallBack) (ret string, err error) {
	var isJson bool
	var isHead bool
	var upload bool
	if types == 1 {
		paths, ok := paramMap["path"]
		value, oks := paramMap["type"]
		//如果没则参数错
		if !ok || !oks {
			c.RBack["status"] = ParamsError
			c.RBack["message"] = Message[ParamsError]
			c.RBack["detail"] = ""
			ret, err = c.RanderJson()
			return
		}
		if value == "json" {
			isJson = true
		} else if value == "path" {
			isJson = false
		} else {
			c.RBack["status"] = ParamsError
			c.RBack["message"] = Message[ParamsError]
			c.RBack["detail"] = ""
			ret, err = c.RanderJson()
			return
		}
		v, okh := paramMap["isHead"]
		if !okh {
			isHead = true
		} else if v == "false" {
			isHead = false
		} else if v != "false" {
			isHead = true
		}
		up, okup := paramMap["upload"]
		if !okup {
			upload = false
		} else if up == "true" {
			upload = true
		} else if v != "true" {
			upload = false
		}

		ret, err = CallBackData(paths, isJson, isHead, upload, c)
	} else if types == 2 {
		ret, err = CallbackCheck(paramMap, c) //营销平台和方舟验证规则
	} else if types == 3 {
		ret, err = CallCheckPortrait(paramMap, c) //验证画像规则
	} else if types == 4 {
		ret, err = CallBackCreateCluesData(paramMap, c)
	} else if types == 5 {
		ret, err = CallCreateExcel(paramMap, c)
	} else {
		c.RBack["status"] = ParamsError
		c.RBack["message"] = Message[ParamsError]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
	}
	return
}

/**\
生成excel
 */
func CallCreateExcel(paramMap map[string]string, c *CallBack) (ret string, err error) {
	paths, ok := paramMap["path"]
	Head, oks := paramMap["title"]
	_, oken := paramMap["encrypt"]
	if !ok {
		c.RBack["status"] = ParamsError
		c.RBack["message"] = Message[ParamsError]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}
	originPath := paths
	ishas := strings.Contains(paths, CommonPath)
	if !ishas { //判断是否相对路径
		paths = ReadValue("rootPath", "rootPath") + paths
	}
	if !Exist(paths) {
		c.RBack["status"] = NoExist
		c.RBack["message"] = Message[NoExist]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}

	dataClue, err := readCluesFileLine(paths)
	if err != nil {
		c.RBack["status"] = ParamsError
		c.RBack["message"] = Message[ParamsError]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}
	var file *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell

	file = xlsx.NewFile()
	sheet, err = file.AddSheet("Sheet1")
	//添加头部
	if oks && strings.Trim(Head, " ") != "" {
		row = sheet.AddRow()
		HeadSlice := strings.Split(Head, "|")
		for _, v := range HeadSlice {
			cell = row.AddCell()
			cell.Value = v
		}
	}
	//生成数据key
	mkey := strconv.FormatInt(time.Now().UnixNano(), 10)
	for _, v := range dataClue {
		row = sheet.AddRow()
		for k, vt := range v {
			cell = row.AddCell()
			if vt == "-" {
				cell.Value = ""
			} else {
				if k == 0 && oken {
					cell.Value = Md5(mkey + vt)
				} else {
					cell.Value = vt
				}

			}
		}
	}
	fileName := ""
	dirName := ""
	strName := ""
	dateDir := ""
	//ishas := strings.Contains(paths, CommonPath)
	if ishas { //如果是绝对路径
		//生成文件名称
		strName = strconv.FormatInt(time.Now().UnixNano(), 10) + RandNum(1000)
		dirName = ReadValue("createFile", "privatePath")
		dateDirs, status := CreateDir(dirName, false)
		dateDir = dateDirs
		if !status {
			c.RBack["status"] = CreateDirError
			c.RBack["message"] = Message[CreateDirError]
			c.RBack["detail"] = ""
			ret, err = c.RanderJson()
			return
		}
		fileName = dirName + dateDir + strName + Suffix
	} else {
		fileName = path.Dir(paths) + "/" + strings.Trim(path.Base(paths), path.Ext(paths)) + Suffix
	}
	err = file.Save(fileName)
	if err != nil {
		fmt.Printf(err.Error())
		c.RBack["status"] = CreateError
		c.RBack["message"] = Message[CreateError]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}
	c.RBack["status"] = Success
	c.RBack["message"] = Message[Success]
	if ishas {
		c.RBack["detail"] = dateDir + strName + Suffix
	} else {
		relativePath := path.Dir(originPath) + "/" + strings.Trim(path.Base(originPath), path.Ext(originPath)) + Suffix
		c.RBack["detail"] = relativePath
	}
	ret, err = c.RanderJson()
	if err != nil {
		fmt.Printf(err.Error())
		c.RBack["status"] = NoKnown
		c.RBack["message"] = Message[NoKnown]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}
	return
}

/**
生成线索文件
 */
func CallBackCreateCluesData(paramMap map[string]string, c *CallBack) (ret string, err error) {
	LevelOne, ok := paramMap["levelOne"]
	LevelTwo, oktwo := paramMap["levelTwo"]
	LevelThree, okthree := paramMap["levelThree"]
	paths, okpath := paramMap["path"]
	Mapk, okm := paramMap["mapkey"]
	_, oken := paramMap["encrypt"]
	originPath := paths
	if !ok || !oktwo || !okthree || !okpath || !okm {
		c.RBack["status"] = ParamsError
		c.RBack["message"] = Message[ParamsError]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}

	ishas := strings.Contains(paths, CommonPath)
	if !ishas { //判断是否相对路径做兼容
		paths = ReadValue("rootPath", "rootPath") + paths
	}
	if !Exist(paths) {
		c.RBack["status"] = NoExist
		c.RBack["message"] = Message[NoExist]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}

	LevelTitle := getLevelPoint()
	LevelOneArray, err := LevelTitle.sloveJson(LevelOne, 1)
	if err != nil {
		c.RBack["status"] = ParamsError
		c.RBack["message"] = Message[ParamsError]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}
	LevelTwoArray, err := LevelTitle.sloveJson(LevelTwo, 2)
	if err != nil {
		c.RBack["status"] = ParamsError
		c.RBack["message"] = Message[ParamsError]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}
	LevelThreeArray, err := LevelTitle.sloveJson(LevelThree, 3)
	if err != nil {
		c.RBack["status"] = ParamsError
		c.RBack["message"] = Message[ParamsError]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}
	//获取位置
	MapKey, err := mapKey(Mapk)
	if err != nil {
		c.RBack["status"] = ParamsError
		c.RBack["message"] = Message[ParamsError]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}
	//获取生成的路径
	ret, err = createExcelData(paths, originPath, LevelOneArray, LevelTwoArray, LevelThreeArray, MapKey, oken, c)
	if err != nil {
		c.RBack["status"] = ParamsError
		c.RBack["message"] = Message[ParamsError]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}
	return
}

//创建 excel
func createExcelData(paths, originPath string, LevelOneArray, LevelTwoArray, LevelThreeArray []LevelType, MapKey map[string]int, encrypt bool, c *CallBack) (ret string, err error) {
	var file *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell
	var style *xlsx.Style
	file = xlsx.NewFile()
	sheet, err = file.AddSheet("Sheet1")
	if err != nil {
		c.RBack["status"] = ParamsError
		c.RBack["message"] = Message[ParamsError]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return

	}
	//excel样式设置
	style = xlsx.NewStyle()
	//设置样式
	style.Fill = *xlsx.NewFill("solid", "0095DD", "0095DD")
	//设置边框
	style.Border = *xlsx.NewBorder("thin", "thin", "thin", "thin")
	//添加第一行
	row = sheet.AddRow()

	for _, v := range LevelOneArray {
		style.Alignment.Horizontal = "center"
		style.Alignment.Vertical = "middle"
		cell = row.AddCell() //添加列
		cell.Value = v.Title
		cell.Merge(v.Num-1, 0)
		cell.SetStyle(style)
		for t := 0; t < v.Num-1; t++ {
			cell = row.AddCell()
			cell.Value = ""
		}
	}
	//添加第二行
	row = sheet.AddRow()
	for _, v := range LevelTwoArray {
		style.Alignment.Horizontal = "center"
		style.Alignment.Vertical = "middle"
		cell = row.AddCell()
		cell.Value = v.Title
		cell.Merge(v.Num-1, 0)
		cell.SetStyle(style)
		for t := 0; t < v.Num-1; t++ {
			cell = row.AddCell()
			cell.Value = ""
		}
	}
	//添加第三行
	row = sheet.AddRow()
	for _, v := range LevelThreeArray {
		style.Alignment.Horizontal = "center"
		style.Alignment.Vertical = "middle"
		cell = row.AddCell()
		cell.Value = v.Title
		cell.Merge(v.Num-1, 0)
		cell.SetStyle(style)
	}
	fileData, err := readCluesFile(paths)
	if err != nil {
		c.RBack["status"] = NoKnown
		c.RBack["message"] = Message[NoKnown]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}
	//生成数据key
	mkey := strconv.FormatInt(time.Now().UnixNano(), 10)
	for _, v := range fileData {
		slice := make([]string, len(MapKey))
		if len(v) < 1 { //如果空数组跳过
			continue
		}
		for k, vt := range v {
			if vt == "-" { //如果内容为 - 就跳出
				continue
			}
			if k == 0 {
				slice[k] = vt
			}
			tag := strings.Split(vt, "|")
			if len(tag) < 2 {
				continue
			}
			vl, ok := MapKey[tag[0]]
			if ok {
				slice[vl] = tag[1]
			}

		}
		if slice[0] == "" {
			continue
		}
		row = sheet.AddRow()
		for m, sv := range slice {
			cell = row.AddCell()
			if encrypt && m == 0 {
				cell.Value = Md5(mkey + sv)
			} else {
				cell.Value = sv
			}
		}
	}
	fileName := path.Dir(paths) + "/" + strings.Trim(path.Base(paths), path.Ext(paths)) + Suffix
	err = file.Save(fileName)
	if err != nil {
		fmt.Printf(err.Error())
		c.RBack["status"] = CreateError
		c.RBack["message"] = Message[CreateError]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}
	c.RBack["status"] = Success
	c.RBack["message"] = Message[Success]
	relativePath := path.Dir(originPath) + "/" + strings.Trim(path.Base(originPath), path.Ext(originPath)) + Suffix
	c.RBack["detail"] = relativePath
	ret, err = c.RanderJson()
	if err != nil {
		fmt.Println(err.Error())
		c.RBack["status"] = NoKnown
		c.RBack["message"] = Message[NoKnown]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}
	return
}

/**
把json类型转化成LevelType
 */
func (L *LevelType) sloveJson(jsonstr string, types int) (Level []LevelType, err error) {
	if types == 1 {
		m := &LevelTypeOne
		err = json.Unmarshal([]byte(jsonstr), m)
		if err != nil {
			fmt.Println(err)
		}
		return *m, err
	} else if types == 2 {
		m := &LevelTypeTwo
		err = json.Unmarshal([]byte(jsonstr), m)
		if err != nil {
			fmt.Println(err)
		}
		return *m, err
	} else if types == 3 {
		m := &LevelTypeThree
		err = json.Unmarshal([]byte(jsonstr), m)
		if err != nil {
			fmt.Println(err)
		}
		return *m, err
	}
	return
}

/**
  获取标签的相对位置
 */
func mapKey(str string) (m map[string]int, err error) {
	err = json.Unmarshal([]byte(str), &m)
	if err != nil {
		fmt.Println(err)
	}
	return
}

//获取LevelType指针
func getLevelPoint() *LevelType {
	return &LevelType{}
}

//验证画像规则
func CallCheckPortrait(paramMap map[string]string, c *CallBack) (ret string, err error) {
	paths, ok := paramMap["path"]
	if !ok {
		c.RBack["status"] = ParamsError
		c.RBack["message"] = Message[ParamsError]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}
	//ishas := strings.Contains(paths, CommonPath)
	//if !ishas {
	//	paths = ReadValue("rootPath", "rootPath") + paths
	//}
	//如果文件不存在
	if !Exist(paths) {
		c.RBack["status"] = NoExist
		c.RBack["message"] = Message[NoExist]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}

	f, err := os.Open(paths)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	bfRd := bufio.NewReader(f)
	i := 0
	message := false
	empty := false
	allLine := 0
	keepStatus := false
	for {
		allLine++
		line, err := bfRd.ReadString('\n')
		if err != nil || io.EOF == err { //遇到任何错误立即返回，并忽略 EOF 错误信息
			break
		}
		//去除 \n 割成数组
		data := strings.Split(strings.TrimRight(line, "\n"), "\t")
		if data[0] != "-" {
			empty = true
			//简单验证电话号码
			phone, err := strconv.ParseInt(data[0], 10, 0)
			if err != nil {
				message = true
				break
			}
			if (phone > 10000000000 && phone < 11000000000) || (phone > 13000000000 && phone < 16000000000) || (phone > 17000000000 && phone < 19000000000) {
				if keepStatus == true {
					message = true
					break
				}
				i++
				continue
			} else {
				//fmt.Println(phone)
				//fmt.Println(i)
				message = true
				break
			}
		} else if data[0] == "-" {
			if keepStatus == false {
				keepStatus = true
			}

		}

		if len(data) > 0 {
			empty = true
		}
	}

	if empty == false {
		c.RBack["status"] = EmptyFile
		c.RBack["message"] = Message[EmptyFile]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}
	if message == true || i == 0 {
		c.RBack["status"] = PhoneError
		c.RBack["message"] = Message[PhoneError]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
	} else {
		c.RBack["status"] = Success
		c.RBack["message"] = Message[Success]
		c.RBack["detail"] = strconv.Itoa(i)
		ret, err = c.RanderJson()
	}
	return
}

//验证检测返回
func CallbackCheck(paramMap map[string]string, c *CallBack) (ret string, err error) {
	subscript, ok := paramMap["subscript"]
	mark, oks := paramMap["mark"]
	head, okc := paramMap["head"]
	paths, okp := paramMap["path"]
	//验证参数
	if !ok || !oks || !okc || !okp {
		c.RBack["status"] = ParamsError
		c.RBack["message"] = Message[ParamsError]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}
	//解析传过来的参数 subscript
	subscriptByte := []byte(subscript)
	var subBox []int
	err = json.Unmarshal(subscriptByte, &subBox)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(subBox)
	//解析参数 mark
	markBox := make(map[string][]int)
	markByte := []byte(mark)
	err = json.Unmarshal(markByte, &markBox)
	if err != nil {
		fmt.Println(err)
	}
	//解析参数 head
	headByte := []byte(head)
	var headbox []HeadTitle
	err = json.Unmarshal(headByte, &headbox)
	if err != nil {
		fmt.Println(err)
	}
	//读取文件
	//ishas := strings.Contains(paths, CommonPath)
	//if !ishas { //如果是相对路径
	//	paths = ReadValue("rootPath", "rootPath") + paths
	//}
	f, err := os.Open(paths)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	bfRd := bufio.NewReader(f)
	//存放所有的错误
	message := make(map[string]map[string]string)
	i := 0
	for {
		i++
		line, err := bfRd.ReadString('\n')
		//fmt.Println(line)
		//行数下标
		errLine := strconv.Itoa(i)
		if err != nil || io.EOF == err { //遇到任何错误立即返回，并忽略 EOF 错误信息
			break
		}
		//把字符串变成数组
		data := strings.Split(strings.TrimRight(line, "\n"), "\t")
		//存错误
		m := make(map[string]string)
		if len(data[0]) == 11 {

			//简单验证电话号码
			phone, err := strconv.ParseInt(data[0], 10, 0)
			//验证电话号码
			if err != nil {
				m["phone"] = "1"
				message[errLine] = m
				//continue
			}

			if phone < 13000000000 || phone > 19000000000 || (phone > 16000000000 && phone < 17000000000) {
				m["phone"] = "1"
				message[errLine] = m
				//continue
			}

		} else if len(data[0]) != 32 {
			m["phone"] = "1"
			message[errLine] = m
			//continue
		}

		var iszero bool = false
		dataLen := len(data) - 1
		//fmt.Println(subBox)
		//循环阶段
		for k, v := range subBox {

			//如果长度小于第一个下标 跳出循环
			if dataLen < v && k == 0 {
				m["level"] = "1"
				message[errLine] = m
				break
			}

			//如果长多小于v 并且iszero true 就跳过

			if dataLen < v { //阶段错误
				if dataLen == (v-1) && data[v-1] == "0" {
					m["level"] = "1"
					message[errLine] = m
					break
				}
				break
			}
			//为了兼容wps   和 microsoft
			//如果最后一个 为空 跳出 并且结束了 就跳出
			if (dataLen == v && data[v] == "" && iszero) || (data[v] == "" && iszero) {
				continue
			}

			//验证没到结束还是成功
			if dataLen == v && !iszero && len(subBox) != k+1 && data[v] == "0" {
				m["level"] = "1"
				message[errLine] = m
				continue
			}

			//已经结束 后面不为空|| 如果 内容为空 并且没有结束
			if iszero && data[v] != "" {
				m["level"] = "1"
				message[errLine] = m
				break
			}

			if data[v] == "" && !iszero {
				m["level"] = "1"
				message[errLine] = m
				break
			}
			dataV := data[v]

			//转成数字32 位
			value, err := strconv.Atoi(data[v])
			//入托不是数字 就跳出
			if err != nil {
				m["level"] = "1"
				message[errLine] = m
				break
			}
			//比较数量 和value值的大小
			lmark := len(markBox[headbox[v].Type])-1 < value
			if lmark {
				m["level"] = "1"
				message[errLine] = m
				break
			}
			//如果已经结束
			if iszero {
				m["level"] = "1"
				message[errLine] = m
				break
			}
			//如果 值不是0 就true
			if dataV != "0" {
				iszero = true
			}
		}
	}
	//需要去除表头的第一行
	linenum := i - 2
	c.RBack["message"] = linenum
	c.RBack["status"] = Success
	c.RBack["detail"] = message
	ret, err = c.RanderJson()
	return
}

//验证 text 的格式是否正确
func CallBackData(paths string, isJson, isHead, upload bool, c *CallBack) (ret string, err error) {
	//如果后缀名错误
	if !checkExtension(paths, Suffix) && !checkExtension(paths, SuffCsv) && !checkExtension(paths, SuffXls) {
		c.RBack["status"] = SuffixError
		c.RBack["message"] = Message[SuffixError]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}
	//如果文件不存在
	ishas := strings.Contains(paths, CommonPath)
	if !ishas { //如果是相对路径
		paths = ReadValue("rootPath", "rootPath") + paths
	}
	if !Exist(paths) {
		c.RBack["status"] = NoExist
		c.RBack["message"] = Message[NoExist]
		c.RBack["detail"] = ""
		ret, err = c.RanderJson()
		return
	}
	retData := new(DataContent)
	//读取Excel
	if checkExtension(paths, Suffix) {
		retData, err = slove(paths, &DataContent{})
		if err != nil {
			fmt.Println(err)
		}
	} else if checkExtension(paths, SuffCsv) {
		retData, err = sloveCsv(paths, &DataContent{})
		if err != nil {
			fmt.Println(err)
		}
	} else if checkExtension(paths, SuffXls) {
		retData, err = SloveXls(paths, &DataContent{})
		if err != nil {
			fmt.Println(err)
		}
	}

	//判断是否返回json
	if isJson {
		c.RBack["status"] = Success
		c.RBack["message"] = Message[Success]
		c.RBack["detail"] = retData.Data
		ret, err = c.RanderJson()
	} else {
		//生成文件名
		strName := strconv.FormatInt(time.Now().UnixNano(), 10) + RandNum(1000)
		var dirName string
		if isHead {
			dirName = ReadValue("createFile", "path")
		} else {
			dirName = ReadValue("createFile", "pathArk")
		}

		if !isHead && upload {
			dirName = ReadValue("createFile", "pathArkUpload")
		}

		dateDir, status := CreateDir(dirName, isHead)
		if !status {
			c.RBack["status"] = CreateDirError
			c.RBack["message"] = Message[CreateDirError]
			c.RBack["detail"] = ""
			ret, err = c.RanderJson()
		}
		fileName := dirName + dateDir + strName + ".txt"
		//fileNameHead := dirName + dateDir + strName + "_head.txt"

		f, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
		defer f.Close()
		if err != nil {
			fmt.Println(err)
		}
		w := bufio.NewWriter(f)
		var firstLen int
		for i, v := range retData.Data {
			var str string
			if !isHead {
				if i == 0 { //第一行去除头部
					if err != nil {
						fmt.Println(err)
					}
					firstLen = len(v)
					if firstLen < 1 {
						break
					}
					continue
				}

				if len(v) == 0 { //如果为空跳出
					continue
				}
				for i := 0; i < firstLen; i++ {
					if i <= len(v)-1 {
						if v[i] == "" {
							str += "-" + "\t"
						} else if i > 2&&!isHead {
							str += retData.Data[0][i] + "|" + v[i] + "\t"
						} else {
							str += v[i] + "\t"
						}
					} else {
						str += "-" + "\t"
					}
				}
				//}
				str = strings.TrimRight(str, "\t")
			} else {
				str = strings.Join(v, "\t")
				//写入文件
			}
			w.WriteString(str + "\n")

		}
		w.Flush()
		c.RBack["status"] = Success
		if len(retData.Data) == 0 { //如果为空时
			c.RBack["message"] = ""
		} else {
			c.RBack["message"] = retData.Data[0]
		}
		c.RBack["detail"] = fileName
		ret, err = c.RanderJson()
	}
	return
}

/**
解析xlsx
 */
func SloveXls(paths string, m *DataContent) (c *DataContent, err error) {

	xlFile, err := xls.Open(paths, "utf-8")
	if err != nil {
		fmt.Println(err)
	}
	if sheet1 := xlFile.GetSheet(0); sheet1 != nil {
		sheetlen := int(sheet1.MaxRow)
		data := make([][]string, len(sheet1.Rows))
		for i := 0; i < sheetlen; i++ {
			if sheet1.Rows[uint16(i)] == nil {
				continue
			}
			row := sheet1.Rows[uint16(i)]
			var slice []string
			var lenmp []int
			for k, _ := range row.Cols {
				lenmp = append(lenmp, int(k))
			}
			sort.Ints(lenmp)
			for t := 0; t <= lenmp[len(lenmp)-1]; t++ {
				colsObj, ok := row.Cols[uint16(t)]
				if !ok {
					slice = append(slice, "")
					continue
				}
				cols := colsObj.String(xlFile)
				if t == 0 {
					slice = append(slice, cols[0])
				} else if t == 1 {
					//fmt.Println(cols)
					for _, v := range cols {
						slice = append(slice, v)
					}
				} else {
					slice = append(slice, cols[0])
				}

			}
			data[i] = slice
		}
		m.Data = data
	}
	return m, err
}

//解析xlsx
func slove(paths string, m *DataContent) (ret *DataContent, err error) {
	//打开文件
	xlFile, err := xlsx.OpenFile(paths)
	if err != nil {
		panic(err)
	}
	//循环 sheet
	for sk, sheet := range xlFile.Sheets {
		if sk > 0 {
			break
		}
		//定义slice 类型interface
		data := make([][]string, len(sheet.Rows))
		for k, row := range sheet.Rows {
			arr := make([]string, len(row.Cells))
			for s, cell := range row.Cells {
				str, e := cell.String()
				if e != nil { //如果不等于nil 恐慌
					panic(e)
				}
				arr[s] = str
			}
			data[k] = arr
		}
		m.Data = data
	}
	return m, err
}

//golang 读取csv文件
func sloveCsv(paths string, m *DataContent) (ret *DataContent, err error) {
	//打开文件
	f, err := os.Open(paths)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	//释放
	defer f.Close()
	decoder := mahonia.NewDecoder("GBK")
	//读取
	reader := csv.NewReader(decoder.NewReader(f))
	var data [][]string
	for {
		//读取一条记录
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return m, err
		}
		arr := make([]string, len(record))
		for k, v := range record {
			arr[k] = v
		}
		data = append(data, arr)
	}
	m.Data = data
	return m, err
}
