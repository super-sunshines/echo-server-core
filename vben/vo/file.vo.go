package vo

type FileUploadVo struct {
	RelativePath string `json:"relativePath"`
	BasePath     string `json:"basePath"`
	FullPath     string `json:"fullPath"`
}
