package tools

const (
	Success        = 0
	SuffixError    = iota
	NoExist
	ParamsError
	EmptyFile
	PhoneError
	NoKnown
	CreateDirError
	Exception
	CreateError
)

var Message = map[int]string{
	Success:        "返回成功",
	SuffixError:    "文件后缀名错误",
	NoExist:        "文件不存在",
	ParamsError:    "参数错误",
	EmptyFile:      "上传模版内容为空，请重新上传",
	PhoneError:     "手机号格式有误，请重新上传！",
	NoKnown:        "未知错误",
	CreateDirError: "创建目录失败，检查是否有权限",
	Exception:      "文件或者数据格式出现问题请检查",
	CreateError:    "创建文件失败",
}
