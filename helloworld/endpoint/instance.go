package endpoint

import "fmt"

type Server struct {
	Name    string `json:"name"`
	Addr    string `json:"addr"`    //服务地址
	Version string `json:"version"` //服务版本
	Weight  int64  `json:"weight"`  //服务权重
}

func BuildPrefix(info Server) string {
	return fmt.Sprintf("/%s/%s/", info.Name, info.Addr)
}
