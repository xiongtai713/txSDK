package main

import (
	"flag"
	"fmt"
	_ "net/http/pprof"
	_ "runtime/pprof"
)


func main() {
	flag.Usage= func() {  //输入 ./main -help 会显示这里的打印
		fmt.Println("输入ip -s ip")

	}
	var e string
	var s= flag.String("s","liu","ip")  // 使用方法 ./main -s nihao

	flag.StringVar(&e,"e","jiao","地址")  //./main -e xxxx

	flag.Parse()  //解析flag
	fmt.Println(*s,e)
	fmt.Println(flag.Args()) //使用方法 ./main 直接输入参数 []string

}
