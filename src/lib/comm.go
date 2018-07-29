package lib

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

//ScanLine 读取整行
func ScanLine() string {
	inputReader := bufio.NewReader(os.Stdin)
	input, _ := inputReader.ReadString('\n')
	return strings.Replace(input, "\n", "", -1)
}

//错误处理
func CheckErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func FmtStr(str string) string {
	// 去除空格
	str = strings.Replace(str, " ", "", -1)
	// 去除换行符
	str = strings.Replace(str, "\n", "", -1)
	// 去除换行符
	str = strings.Replace(str, "\r", "", -1)
	return str

}
