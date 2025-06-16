package routers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/super-sunshines/echo-server-core/core"
	"github.com/super-sunshines/echo-server-core/vben/vo"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var SysFileRouterGroup = core.NewRouterGroup("", NewSysFileRouter, func(rg *echo.Group, group *core.RouterGroup) error {
	return group.Reg(func(m *SysFileRouter) {
		server := core.GetConfig().Server
		if server.BaseStaticFolder == "" {
			server.BaseStaticFolder = "./static/"
		}
		rg.Static("/static", server.BaseStaticFolder)
		rg.POST("/upload", m.upload)
	})

})

type SysFileRouter struct {
}

func (r SysFileRouter) generateFilename(rawFilename string) string {
	// 移除路径部分（只获取文件名）
	nameWithoutPath := filepath.Base(rawFilename)
	// 分割文件名和扩展名
	ext := filepath.Ext(nameWithoutPath)
	basename := strings.TrimSuffix(nameWithoutPath, ext)
	// 处理没有扩展名的情况
	if ext == "" {
		return fmt.Sprintf("%s-%d", basename, time.Now().Unix())
	}

	// 生成带时间戳的新文件名
	return fmt.Sprintf("%s-%d%s", basename, time.Now().Unix(), ext)
}

// buildTargetPath 构建安全的文件保存路径
func (r SysFileRouter) buildTargetPath(fullFileName, originalFilename, baseFolder string) (string, error) {
	// 处理文件名
	var filename string
	if fullFileName != "" {
		filename = r.generateFilename(filepath.Base(fullFileName))
	} else {
		filename = r.generateFilename(originalFilename)
	}

	// 构建路径
	var pathBuilder string
	if fullFileName != "" {
		// 清理路径并确保以/开头
		cleanPath := filepath.Clean("/" + strings.TrimPrefix(fullFileName, "/"))
		pathBuilder = filepath.Dir(cleanPath)
	} else {
		pathBuilder = "/"
	}

	// 合并基础目录和最终路径
	finalPath := filepath.Join(baseFolder, pathBuilder, filename)

	return filepath.Clean(finalPath), nil
}

// @Summary	上传文件
// @Tags		[系统]文件系统
// @Success	200	{object}	core.ResponseSuccess{data=vo.FileUploadVo}
// @Router		/upload [POST]
// @Param 		file	formData	file	true	"文件"
// @Param 		fullName	formData	string	true	"文件名"
func (r SysFileRouter) upload(c echo.Context) error {
	server := core.GetConfig().Server
	context := core.GetAnyContext(c)
	file, err := c.FormFile("file")
	fullFileName := c.FormValue("fullName")
	if err != nil {
		return err
	}
	targetPath, err := r.buildTargetPath(fullFileName, file.Filename, server.BaseStaticFolder)
	if err != nil {
		return err
	}
	// 4. 创建目录
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	// 6. 创建目标文件
	dst, err := os.Create(targetPath)
	defer func(src multipart.File) {}(src)
	if err != nil {
		return err
	}
	defer func(dst *os.File) {}(dst)
	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	relativePath, _ := strings.CutPrefix(
		strings.ReplaceAll(fmt.Sprintf("/%s", filepath.ToSlash(targetPath)), "//", "/"),
		strings.ReplaceAll(filepath.ToSlash(server.BaseStaticFolder), "./", "/"))
	// 处理文件名
	return context.Success(vo.FileUploadVo{
		RelativePath: relativePath,
		BasePath:     server.ServerDomain + "/api/static",
		FullPath:     strings.ReplaceAll(fmt.Sprintf("%s/api/static%s", server.ServerDomain, relativePath), "\\", "/"),
	})
}

func NewSysFileRouter() *SysFileRouter {
	return &SysFileRouter{}
}
