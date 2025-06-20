package core

import (
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"time"
)

var tokenManager *TokenManager

func GetTokenManager() *TokenManager {
	if tokenManager == nil {
		initTokenManger()
	}
	return tokenManager
}

type TokenManager struct {
	*RedisCache[TokenInfo]
}
type TokenInfo struct {
	Token    string `json:"token"`
	ExpireAt int64  `json:"expireAt"`
}
type ClaimsAdditions struct {
	UID          int64    `json:"UID"`          // 账号
	Username     string   `json:"username"`     // 昵称
	DepartmentId int64    `json:"departmentId"` // 部门ID
	NickName     string   `json:"nickName"`     // 昵称
	RoleCodes    []string `json:"roleCodes"`    // 角色码
	Platform     string   `json:"platform"`
}
type Claims struct {
	ClaimsAdditions
	jwt.RegisteredClaims
}

func initTokenManger() {
	tokenManager = &TokenManager{
		RedisCache: GetRedisCache[TokenInfo]("sys:token:info:"),
	}
}
func (j TokenManager) GetJwtExpirationTime(tokenStr string) int64 {
	claims := &Claims{}
	_, _ = jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(GetConfig().Jwt.JwtKey), nil
	})
	return claims.ExpiresAt.Unix()
}
func (j TokenManager) ParseJwt(token string, platform ...string) (Claims, *CodeError) {
	selectPlatform := AdditionFirst(platform, "")
	key := GetConfig().Jwt.JwtKey
	claims := Claims{}
	tkn, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})
	if (err != nil) || (tkn != nil && !tkn.Valid) {
		zap.L().Info(fmt.Sprintf("Token 解析出错！%#v", err))
		return claims, NewErrCodeMsg(TOKEN_EXPIRE_ERROR, err.Error())
	}
	// 检测严格模式
	if j.GetPlatformStrict(selectPlatform) {
		exists := GetTokenManager().XHExists(fmt.Sprintf("%d:%s", claims.UID, claims.Platform))
		if !exists {
			return claims, NewErrCodeMsg(TOKEN_EXPIRE_ERROR, "token已过期")
		}
	}
	return claims, nil
}

func (j TokenManager) GetPlatformExpiration(platform string) int64 {
	config := GetConfig().Jwt
	if platform == "" {
		return config.Expire
	}
	// 查找指定平台的过期时间配置
	if v, ok := slice.FindBy(config.SpecifiedConfig, func(i int, item SpecifiedPlatform) bool {
		return item.Platform == platform
	}); ok && v.Expire > 0 {
		return v.Expire
	}
	return config.Expire
}

func (j TokenManager) GetPlatformStrict(platform string) bool {
	config := GetConfig().Jwt
	if platform == "" {
		return config.Strict
	}
	// 查找指定平台的严格模式配置
	for _, item := range config.SpecifiedConfig {
		if item.Platform == platform {
			return item.Strict
		}
	}
	return config.Strict
}

func (j TokenManager) GenJwtString(platform string, user ClaimsAdditions) (string, error) {
	userJwt := j.GetUserJwt(user.UID, platform)
	expiration := j.GetPlatformExpiration(platform)

	// 当前的Token 有效 且 过期时间 大于 当前时间 加上 缓存时间/2
	if userJwt.Token != "" && userJwt.ExpireAt > GetNowLocalTime().Unix()+expiration/2 {
		return userJwt.Token, nil
	}
	// 声明 token 的过期时间
	expirationTime := GetNowLocalTime().Add(time.Second * time.Duration(expiration))
	// 创建 JWT claims ，其中包括用户名和过期时间
	claims := &Claims{
		ClaimsAdditions: user,
		RegisteredClaims: jwt.RegisteredClaims{
			// 在 JWT 中，过期时间表示为 Unix 毫秒时间戳
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	// 声明使用的算法和 claims 来创建 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 创建 JWT 字符串
	signedString, err := token.SignedString([]byte(GetConfig().Jwt.JwtKey))
	if err != nil {
		zap.L().Error(fmt.Sprintf("生成Token出错！%#v", err))
		return "", err
	}
	j.SetUserJwt(user.UID, platform, signedString, expirationTime.Unix())
	return signedString, nil
}
func (j TokenManager) SetUserJwt(uid int64, platform, token string, expireAt int64) {
	j.XHSet(fmt.Sprintf("%d:%s", uid, platform), TokenInfo{
		Token: token, ExpireAt: expireAt,
	})
}
func (j TokenManager) GetUserJwt(uid int64, platform string) TokenInfo {
	return j.XHGet(fmt.Sprintf("%d:%s", uid, platform))
}
func (j TokenManager) ValidToken(uid int64, platform, token string) bool {
	tokenInfo := j.XHGet(fmt.Sprintf("%d:%s", uid, platform))
	return token != "" && tokenInfo.ExpireAt > GetNowTimeUnix()
}
func (j TokenManager) RemoveTokenByUid(uid int64) bool {
	return j.XHDel(fmt.Sprintf("%d:*", uid))
}
func (j TokenManager) RemoveToken(uid int64, platform string) bool {
	return j.XHDel(fmt.Sprintf("%d:%s", uid, platform))
}
