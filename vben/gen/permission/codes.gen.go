package main

import (
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// 权限码扫描生成行函数
func main() {
	// 要扫描的目录
	dir := "./vben/"
	// 扫描的函数
	functions := []string{"HavePermission", "HaveOneOfPermissions", "HaveAllPermissions"} // 可变函数名

	outputFile := dir + "const/permission.gen.go"

	packageName := "_const"

	codes, err := FindPermissions(dir, functions)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if len(codes) == 0 {
		fmt.Println("No permission codes found.")
	} else {
		fmt.Println("Permission codes found:")
		for _, code := range codes {
			fmt.Println(code)
		}
	}
	unique := slice.Unique(codes)

	err = WriteCodesToFile(unique, packageName, outputFile)
	if err != nil {
		fmt.Println(err)
	}

}

// WriteCodesToFile 将权限代码写入 Go 文件
func WriteCodesToFile(codes []string, packageName, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("package %s\n\nvar GeneratePermissionCodes = []string{\n", packageName))
	if err != nil {
		return err
	}

	for _, code := range codes {
		code = strings.Replace(code, `"`, "", -1)
		_, err = file.WriteString(fmt.Sprintf("\t%q,\n", code)) // 直接写入字符串
		if err != nil {
			return err
		}
	}

	_, err = file.WriteString("}\n") // 结束切片
	return err
}

// PermissionVisitor 结构体用于访问函数调用
type PermissionVisitor struct {
	Codes []string
	Funcs []string
}

// Visit 方法实现了 ast.Visitor 接口
func (v *PermissionVisitor) Visit(node ast.Node) ast.Visitor {
	if callExpr, ok := node.(*ast.CallExpr); ok {
		if fun, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
			// 检查是否为指定的函数
			if ident, ok := fun.X.(*ast.Ident); ok && ident.Name == "core" {
				for _, funcName := range v.Funcs {
					if fun.Sel.Name == funcName {
						for _, arg := range callExpr.Args {
							if strLit, ok := arg.(*ast.BasicLit); ok {
								v.Codes = append(v.Codes, strLit.Value)
							}
						}
					}
				}
			}
		}
	}
	return v
}

// FindPermissions 遍历指定目录，找到所有权限代码
func FindPermissions(relativeDir string, funcs []string) ([]string, error) {
	var codes []string

	// 获取当前工作目录
	baseDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	println(baseDir)
	// 组合成绝对路径
	dir := filepath.Join(baseDir, relativeDir)

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".go" {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
			if err != nil {
				return err
			}

			visitor := &PermissionVisitor{Funcs: funcs}
			ast.Walk(visitor, node)
			codes = append(codes, visitor.Codes...)
		}
		return nil
	})

	return codes, err
}
