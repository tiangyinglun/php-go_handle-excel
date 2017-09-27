package tools

type LevelType struct {
	Title string
	Num   int
}

var LevelTypeOne []LevelType
var LevelTypeTwo []LevelType
var LevelTypeThree []LevelType

const (
	Suffix     = ".xlsx"
	SuffCsv    = ".csv"
	SuffXls    = ".xls"
	Portrait   = 3
	CommonPath = "storage"
)

//定义Excel返回类型
type DataContent struct {
	Data [][]string
}

//定义返回类型
type CallBack struct {
	RBack map[string]interface{}
}

type HeadTitle struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Ext  int    `json:"ext"`
}

var Maps map[string]int
