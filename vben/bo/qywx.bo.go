package bo

type BindRealNameBo struct {
	QywxUid  string `json:"qywxUid" zh_comment:"企业微信UID" en_comment:"uid" validate:"required"`
	RealName string `json:"realName" zh_comment:"真实姓名" en_comment:"real name" validate:"required"`
}
