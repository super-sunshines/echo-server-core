package core

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"time"
)

var innerRedis *redis.Client
var ctx = context.Background()

func initRedis() {
	innerRedis = redis.NewClient(&redis.Options{
		Addr:     config.Redis.Addr,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})
	_, err := innerRedis.Ping(context.Background()).Result()
	if err != nil {
		zap.L().Error("redis connect error", zap.Error(err))
		panic(err)
	}
}

func GetRedisCache[T any](key string) *RedisCache[T] {
	if innerRedis == nil {
		initRedis()
	}
	return &RedisCache[T]{
		Client: innerRedis,
		key:    key,
	}
}

// RedisCache
// 结构体需要将字段导出才可以正常使用
type RedisCache[T any] struct {
	*redis.Client
	key string
}

func (client *RedisCache[T]) Marshal(value T) string {
	if marshal, err := json.Marshal(value); err != nil {
		zap.L().Error(err.Error())
		return ""
	} else {
		return string(marshal)
	}
}

func (client *RedisCache[T]) UnMarshal(str string) T {
	var value = new(T)
	if err := json.Unmarshal([]byte(str), &value); err != nil {
		zap.L().Error(err.Error())
	}
	return *value
}

// XSet  设置 key的值
func (client *RedisCache[T]) XSet(value T) bool {
	result, err := client.Set(ctx, client.key, client.Marshal(value), 0).Result()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return result == "OK"
}

// XSetEX 设置 key的值并指定过期时间
func (client *RedisCache[T]) XSetEX(value T, ex time.Duration) bool {
	result, err := client.Set(ctx, client.key, client.Marshal(value), ex).Result()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return result == "OK"
}

// XSetCodeEX 设置 指定Code值的值并指定过期时间
func (client *RedisCache[T]) XSetCodeEX(appendCode string, value T, ex time.Duration) bool {
	result, err := client.Set(ctx, client.key+appendCode, client.Marshal(value), ex).Result()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return result == "OK"
}

// XCodeExists 查询附加的Code是否存在
func (client *RedisCache[T]) XCodeExists(appendCode string) bool {
	result, err := client.Exists(ctx, client.key+appendCode).Result()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return result == 1
}

// XCodeGet 附加code获取
func (client *RedisCache[T]) XCodeGet(appendCode string) (have bool, value T) {
	result, err := client.Get(ctx, client.key+appendCode).Result()
	if err != nil {
		fmt.Println(err)
		return false, client.UnMarshal(result)
	}
	return true, client.UnMarshal(result)
}

// XCodeDel 附加code获取
func (client *RedisCache[T]) XCodeDel(appendCode string) (have bool) {
	affectRows, err := client.Del(ctx, client.key+appendCode).Result()
	if err != nil {
		return false
	}
	return affectRows > 0
}

// XGet 获取 key的值
func (client *RedisCache[T]) XGet() (have bool, value T) {
	result, err := client.Get(ctx, client.key).Result()
	if err != nil {
		fmt.Println(err)
		return false, client.UnMarshal(result)
	}
	return true, client.UnMarshal(result)
}

// XGetSet 设置新值获取旧值
func (client *RedisCache[T]) XGetSet(value T) (bool, T) {
	oldValue, err := client.GetSet(ctx, client.key, client.Marshal(value)).Result()
	if err != nil {
		fmt.Println(err)
		return false, client.UnMarshal(oldValue)
	}
	return true, client.UnMarshal(oldValue)
}

// XIncr key值每次加一 并返回新值
func (client *RedisCache[T]) XIncr() int64 {
	val, err := client.Incr(ctx, client.key).Result()
	if err != nil {
		fmt.Println(err)
	}
	return val
}

// XIncrBy key值每次加指定数值 并返回新值
func (client *RedisCache[T]) XIncrBy(incr int64) int64 {
	val, err := client.IncrBy(ctx, client.key, incr).Result()
	if err != nil {
		fmt.Println(err)
	}
	return val
}

// XIncrByFloat key值每次加指定浮点型数值 并返回新值
func (client *RedisCache[T]) XIncrByFloat(incrFloat float64) float64 {
	val, err := client.IncrByFloat(ctx, client.key, incrFloat).Result()
	if err != nil {
		fmt.Println(err)
	}
	return val
}

// XDecr key值每次递减 1 并返回新值
func (client *RedisCache[T]) XDecr() int64 {
	val, err := client.Decr(ctx, client.key).Result()
	if err != nil {
		fmt.Println(err)
	}
	return val
}

// XDecrBy key值每次递减指定数值 并返回新值
func (client *RedisCache[T]) XDecrBy(incr int64) int64 {
	val, err := client.DecrBy(ctx, client.key, incr).Result()
	if err != nil {
		fmt.Println(err)
	}
	return val
}

// XDel 删除 key
func (client *RedisCache[T]) XDel() bool {
	result, err := client.Del(ctx, client.key).Result()
	if err != nil {
		return false
	}
	return result == 1
}

// XExpire 设置 key的过期时间
func (client *RedisCache[T]) XExpire(ex time.Duration) bool {
	result, err := client.Expire(ctx, client.key, ex).Result()
	if err != nil {
		return false
	}
	return result
}

/*------------------------------------ list 操作 ------------------------------------*/

// XLPush 从列表左边插入数据，并返回列表长度
func (client *RedisCache[T]) XLPush(date ...T) int64 {
	var dateAny []string
	for _, t := range date {
		dateAny = append(dateAny, client.Marshal(t))
	}
	result, err := client.LPush(ctx, client.key, dateAny).Result()
	if err != nil {
		fmt.Println(err)
	}
	return result
}

// XRPush 从列表右边插入数据，并返回列表长度
func (client *RedisCache[T]) XRPush(date ...T) int64 {
	var dateAny []string
	for _, t := range date {
		dateAny = append(dateAny, client.Marshal(t))
	}
	result, err := client.RPush(ctx, client.key, dateAny).Result()
	if err != nil {
		fmt.Println(err)
	}
	return result
}

// XLPop 从列表左边删除第一个数据，并返回删除的数据
func (client *RedisCache[T]) XLPop() (bool, T) {
	val, err := client.LPop(ctx, client.key).Result()
	if err != nil {
		fmt.Println(err)
		return false, client.UnMarshal(val)
	}
	return true, client.UnMarshal(val)
}

// XRPop 从列表右边删除第一个数据，并返回删除的数据
func (client *RedisCache[T]) XRPop() (bool, T) {
	val, err := client.RPop(ctx, client.key).Result()
	if err != nil {
		fmt.Println(err)
		return false, client.UnMarshal(val)
	}
	return true, client.UnMarshal(val)
}

// XLIndex 根据索引坐标，查询列表中的数据
func (client *RedisCache[T]) XLIndex(index int64) (bool, string) {
	val, err := client.LIndex(ctx, client.key, index).Result()
	if err != nil {
		fmt.Println(err)
		return false, ""
	}
	return true, val
}

// XLLen 返回列表长度
func (client *RedisCache[T]) XLLen() int64 {
	val, err := client.LLen(ctx, client.key).Result()
	if err != nil {
		fmt.Println(err)
	}
	return val
}

// XLRange 返回列表的一个范围内的数据，也可以返回全部数据
func (client *RedisCache[T]) XLRange(start, stop int64) []T {
	vales, err := client.LRange(ctx, client.key, start, stop).Result()
	if err != nil {
		fmt.Println(err)
	}
	var valesT []T
	for _, t := range vales {
		valesT = append(valesT, client.UnMarshal(t))
	}
	return valesT
}

// XLRem 从列表左边开始，删除元素data， 如果出现重复元素，仅删除 count次
func (client *RedisCache[T]) XLRem(count int64, data T) bool {
	_, err := client.LRem(ctx, client.key, count, client.Marshal(data)).Result()
	if err != nil {
		fmt.Println(err)
	}
	return true
}

// XLInsert 在列表中 pivot 元素的后面插入 data
func (client *RedisCache[T]) XLInsert(pivot int64, data T) bool {
	err := client.LInsert(ctx, client.key, "after", pivot, client.Marshal(data)).Err()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

/*------------------------------------ set 操作 ------------------------------------*/

// XSAdd 添加元素到集合中
func (client *RedisCache[T]) XSAdd(data ...T) bool {
	var strSlice []string
	for _, datum := range data {
		strSlice = append(strSlice, client.Marshal(datum))
	}
	err := client.SAdd(ctx, client.key, strSlice).Err()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// XSCard 获取集合元素个数
func (client *RedisCache[T]) XSCard() int64 {
	size, err := client.SCard(ctx, client.key).Result()
	if err != nil {
		fmt.Println(err)
	}
	return size
}

// XSIsMember 判断元素是否在集合中
func (client *RedisCache[T]) XSIsMember(data T) bool {
	ok, err := client.SIsMember(ctx, client.key, client.Marshal(data)).Result()
	if err != nil {
		fmt.Println(err)
	}
	return ok
}

// XSMembers 获取集合所有元素
func (client *RedisCache[T]) XSMembers() []T {

	var valesList []T
	es, err := client.SMembers(ctx, client.key).Result()
	if err != nil {
		fmt.Println(err)
	}
	for _, e := range es {
		valesList = append(valesList, client.UnMarshal(e))
	}
	return valesList
}

// XSRem 删除 key集合中的 data元素
func (client *RedisCache[T]) XSRem(data ...T) bool {
	var _data []interface{}
	for _, datum := range data {
		_data = append(_data, client.Marshal(datum))
	}
	_, err := client.SRem(ctx, client.key, _data).Result()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// XSPopN 随机返回集合中的 count个元素，并且删除这些元素
func (client *RedisCache[T]) XSPopN(count int64) (values []T) {
	vales, err := client.SPopN(ctx, client.key, count).Result()
	if err != nil {
		fmt.Println(err)
	}
	var ts []T
	for _, vale := range vales {
		ts = append(ts, client.UnMarshal(vale))
	}
	return ts
}

/*------------------------------------ hash 操作 ------------------------------------*/

// XHSet 根据 key和 field字段设置，field字段的值
func (client *RedisCache[T]) XHSet(field string, value T) bool {
	err := client.HSet(ctx, client.key, field, client.Marshal(value)).Err()
	if err != nil {
		return false
	}
	return true
}

// XHGet 根据 key和 field字段，查询field字段的值
func (client *RedisCache[T]) XHGet(field string) T {
	val, err := client.HGet(ctx, client.key, field).Result()
	if err != nil {
		fmt.Println(err)
	}
	return client.UnMarshal(val)
}

// XHMGet 根据key和多个字段名，批量查询多个 hash字段值
func (client *RedisCache[T]) XHMGet(fields ...string) []T {
	vales, err := client.HMGet(ctx, client.key, fields...).Result()
	var valesList []T
	for i := range vales {
		if vales[i] == nil {
			continue
		}
		valesList = append(valesList, client.UnMarshal(vales[i].(string)))
	}
	if err != nil {
		zap.L().Error("XHMGet error", zap.Error(err))
	}
	return valesList
}

// XHGetAll 根据 key查询所有字段和值
func (client *RedisCache[T]) XHGetAll() map[string]T {
	data, err := client.HGetAll(ctx, client.key).Result()
	mapData := make(map[string]T)
	for k, v := range data {
		mapData[k] = client.UnMarshal(v)
	}
	if err != nil {
		fmt.Println(err)
	}
	return mapData
}

// XHKeys 根据 key返回所有字段名
func (client *RedisCache[T]) XHKeys() []string {
	fields, err := client.HKeys(ctx, client.key).Result()
	if err != nil {
		fmt.Println(err)
	}
	return fields
}

// XHLen 根据 key，查询hash的字段数量
func (client *RedisCache[T]) XHLen() int64 {
	size, err := client.HLen(ctx, client.key).Result()
	if err != nil {
		fmt.Println(err)
	}
	return size
}

// XHMSet  根据 key和多个字段名和字段值，批量设置 hash字段值
func (client *RedisCache[T]) XHMSet(data map[string]T) bool {
	m := make(map[string]interface{}, len(data))
	for k, v := range data {
		m[k] = client.Marshal(v)
	}
	result, err := client.HMSet(ctx, client.key, m).Result()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return result
}

// XHSetNX 如果 field字段不存在，则设置 hash字段值
func (client *RedisCache[T]) XHSetNX(field string, value T) bool {
	result, err := client.HSetNX(ctx, client.key, field, client.Marshal(value)).Result()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return result
}

// XHDel 根据 key和字段名，删除 hash字段，支持批量删除
func (client *RedisCache[T]) XHDel(fields ...string) bool {
	_, err := client.HDel(ctx, client.key, fields...).Result()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// XHExists 检测 hash字段名是否存在
func (client *RedisCache[T]) XHExists(field string) bool {
	result, err := client.HExists(ctx, client.key, field).Result()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return result
}
