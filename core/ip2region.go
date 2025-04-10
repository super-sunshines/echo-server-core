package core

import (
	"fmt"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"os"
)

var ip2RegionSearcher *xdb.Searcher

// LoadContentFromFile 从指定文件中加载内容
func LoadContentFromFile(filename string) ([]byte, error) {
	// 读取文件内容
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func initIp2Region() {
	var filePath = config.Ip2RegionConfig

	// 1、从 dbPath 加载整个 xdb 到内存
	cBuff, err := LoadContentFromFile(filePath.FilePath)
	if err != nil {
		fmt.Printf("failed to load content from `%s`: %s\n", filePath.FilePath, err)
		return
	}

	// 2、用全局的 cBuff 创建完全基于内存的查询对象。
	ip2RegionSearcher, err = xdb.NewWithBuffer(cBuff)
	if err != nil {
		fmt.Printf("failed to create searcher with content: %s\n", err)
		return
	}
}
func IPParse(ip string) (string, error) {
	if ip2RegionSearcher == nil {
		initIp2Region()
	}
	return ip2RegionSearcher.SearchByStr(ip)
}
