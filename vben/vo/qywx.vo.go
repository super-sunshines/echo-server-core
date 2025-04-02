package vo

type SignatureVo struct {
	Timestamp int64  `json:"timestamp"` //token
	NonceStr  string `json:"nonceStr"`  // 过期时间
	Signature string `json:"signature"` // 过期时间
}
