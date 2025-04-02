package services

import (
	"github.com/XiaoSGentle/echo-server-core/core"
	"time"

	"github.com/tencentyun/qcloud-cos-sts-sdk/go"
)

var txc *TencentCloud

func NewTencentCloudService() *TencentCloud {
	if txc == nil {
		txc.Init()
	}
	return txc
}

type TencentCloud struct {
	CosClient           *sts.Client
	CosClientRedisCache *core.RedisCache[TencentCloudCosTmpKey]
}

// TencentCloudCosTmpKey 结构体用于封装腾讯云COS（Cloud Object Storage）的临时密钥信息。
// 它不仅包含了获取临时密钥的结果，还包含了与COS存储桶相关的元数据。
// 该结构体主要用于简化对象存储服务中的身份验证和资源定位过程。
type TencentCloudCosTmpKey struct {
	// CredentialResult 嵌入式字段，继承了临时密钥的相关信息，如访问密钥ID、秘密访问密钥等。
	*sts.CredentialResult

	// Bucket 表示COS存储桶的名称，用于定位资源。
	Bucket string `json:"Bucket"`

	// Region 表示存储桶所在的地理区域，用于路由请求到正确的数据中心。
	Region string `json:"Region"`

	// CdnUrl 表示内容分发网络（CDN）的URL，用于加速内容的分发和访问。
	CdnUrl string `json:"CdnUrl"`
}

func (t *TencentCloud) Init() {
	tencentConfig := core.GetConfig().Tencent
	cosClient := sts.NewClient(tencentConfig.Cos.SecretId, tencentConfig.Cos.SecretKey, nil)
	txc = &TencentCloud{
		CosClient:           cosClient,
		CosClientRedisCache: core.GetRedisCache[TencentCloudCosTmpKey]("tencent-cos-temp-key-cache-key"),
	}
}

// GetTempCosKey 用于获取腾讯云COS的临时密钥。
// 该方法根据腾讯云的配置信息，生成一个具有限定权限和时效性的COS访问密钥。
// 返回值是一个包含临时密钥信息的 TencentCloudCosTmpCKey 结构体和一个错误对象。
func (t *TencentCloud) GetTempCosKey() (TencentCloudCosTmpKey, error) {

	have, value := t.CosClientRedisCache.XGet()
	if have {
		return value, nil
	}
	// 从全局配置中获取腾讯云COS的配置信息。
	tencentCosConfig := core.GetConfig().Tencent.Cos

	// 初始化获取临时密钥的选项。
	opt := &sts.CredentialOptions{
		Policy: &sts.CredentialPolicy{Statement: []sts.CredentialPolicyStatement{
			{
				// 定义允许的操作，包括文件上传、分片上传等。
				Action: []string{
					"name/cos:PutObject", "name/cos:PostObject", "name/cos:sliceUploadFile", "name/cos:InitiateMultipartUpload",
					"name/cos:ListMultipartUploads", "name/cos:ListParts", "name/cos:UploadPart", "name/cos:CompleteMultipartUpload", "name/cos:AbortMultipartUpload",
				},
				Effect: "allow",
				// 定义允许访问的资源路径前缀，根据实际需求设置。
				Resource: []string{
					"qcs::cos:" + tencentCosConfig.Region + ":uid/" + tencentCosConfig.AppId + ":" + tencentCosConfig.Bucket + "/*",
				},
			},
		}},
		// 设置临时密钥的有效期为1小时。
		Region:          tencentCosConfig.Region,
		DurationSeconds: int64(time.Hour.Seconds()),
	}

	// 调用COS客户端的GetCredential方法获取临时密钥。
	credentialResult, err := t.CosClient.GetCredential(opt)
	key := TencentCloudCosTmpKey{
		CredentialResult: credentialResult,
		Bucket:           tencentCosConfig.Bucket,
		Region:           tencentCosConfig.Region,
		CdnUrl:           tencentCosConfig.CdnUrl,
	}

	t.CosClientRedisCache.XSetEX(key, time.Hour-time.Minute*10)
	// 返回包含临时密钥信息的结构体和可能的错误。
	return key, err
}
