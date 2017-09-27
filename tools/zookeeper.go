package tools

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"time"
	"strings"
)

//zookeeper struct
type Zook struct {
	zooCon *zk.Conn
}

//使用 zookeeper
func (zoo *Zook) Zookeeper() error {
	var zkList []string
	zooHost := ReadValue("zookeeper", "zooHost")
	zooPort := ReadValue("zookeeper", "zooProt")
	//把zookeeper地址加入
	zkList = append(zkList, zooHost+":"+zooPort)
	//节点地址
	node := ReadValue("zookeeper", "zooParentNode") + "/" + ReadValue("addr", "ip") + ":" + ReadValue("addr", "port")
	//连接
	err := zoo.connZookeeper(zkList)
	if err != nil {
		return nil
	}
	err = zoo.CreateNode(node)
	if err != nil {
		return nil
	}
	return err

}

//连接 con 连接

func (zoo *Zook) connZookeeper(zkList []string) error {
	//连接zookeeper
	zk, _, err := zk.Connect(zkList, time.Second*10)
	if err != nil {
		fmt.Println(err)
	}
	zoo.zooCon = zk
	return err
}

//创建子节点
func (zoo *Zook) CreateNode(node string) error {

	if !strings.Contains(node, "/") {
		fmt.Println("节点填写错误")
	}
	parentString := strings.TrimLeft(node, "/")
	var box []string
	if strings.Contains(parentString, "/") {
		box = strings.Split(parentString, "/")

	} else {
		fmt.Println(parentString)
		box = append(box, parentString)
	}
	str := ""
	for k, v := range box {
		str += "/" + v
		//检测节点
		ret, _, err := zoo.zooCon.Exists(str)
		if err != nil {
			fmt.Println("节点查询失败")
			return err
		}
		if ret {
			fmt.Println("节点查询已存在")
			continue
		}
		var status int32 = 0
		if len(box)-1 == k {
			status = zk.FlagEphemeral
		}
		strback, err := zoo.zooCon.Create(str, nil, status, zk.WorldACL(zk.PermAll))
		if err != nil {
			fmt.Println("创建节点失败" + str)
			return err
		}
		fmt.Println(strback)
	}
	return nil

}

//关闭资源
func (zoo *Zook) ZookClose() {
	zoo.zooCon.Close()
}
