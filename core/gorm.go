package core

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLog "gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"reflect"
	"strings"
	"time"
)

type GormGlobalHook = func(db *gorm.DB)

// InitGormOptions 初始化的时候调用的参数
type InitGormOptions struct {
	GormGlobalHook GormGlobalHook
}

var initOptions *InitGormOptions
var db *gorm.DB

type GormConfig struct {
	Host            string
	Port            int64
	DataBase        string
	User            string
	Pass            string
	OtherSettings   string
	ConnMaxLifetime int
	MaxIdleTime     int
	SetMaxIdleConn  int
	SetMaxOpenConn  int
}

func initGormConfig(options InitGormOptions) {
	initOptions = &options
}

func GetGormDB() *gorm.DB {
	if db == nil {
		connectDataBase()
	}
	return db
}

func connectDataBase() *gorm.DB {
	options := GetConfig().DataBase
	connectUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&charset=utf8mb4%s",
		options.User, options.Pass, options.Host, options.Port, options.DataBase, options.OtherSettings)
	mysqlDialectic := mysql.Open(connectUrl)
	var gormLogger gormLog.Interface
	gormLogger = gormLog.Default.LogMode(gormLog.Info)
	gormDb, err := gorm.Open(mysqlDialectic, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		panic(err)
	}

	if initOptions.GormGlobalHook != nil {
		initOptions.GormGlobalHook(gormDb)
	}

	resolverConf := dbresolver.Config{
		Replicas: []gorm.Dialector{mysqlDialectic}, //  读 操作库，查询类
		Policy:   dbresolver.RandomPolicy{},        // sources/replicas 负载均衡策略适用于
	}
	err = gormDb.Use(dbresolver.Register(resolverConf).
		SetConnMaxIdleTime(time.Duration(options.MaxIdleTime) * time.Second).
		SetConnMaxLifetime(time.Duration(options.ConnMaxLifetime) * time.Second).
		SetMaxIdleConns(options.SetMaxIdleConn).
		SetMaxOpenConns(options.SetMaxOpenConn))
	if err != nil {
		panic(err)
	}
	db = gormDb
	return gormDb
}

type InjectServiceConfig struct {
	SpecialPrimaryKey []string
	PrimaryKeyField   string
	CreateDeptField   string
	CreateByField     string
	CreateTimeField   string
	UpdateTimeField   string
	UpdateByField     string
	DeleteTimeField   string
	limitOne          string
}

type PreGorm[M any, V any] struct {
	config InjectServiceConfig
}

func NewService[M any, V any](config ...InjectServiceConfig) PreGorm[M, V] {
	return PreGorm[M, V]{
		config: mergeInjectServiceDefaultConfig(config...),
	}
}

// WithContext 限制有Context的
// Context 是 echo.Context 封装好的 里面有 Gorm DB
// WithContext 的话就自动把 Context 里面的 DB 赋值给 Gorm
func (receiver PreGorm[M, V]) WithContext(c echo.Context) *Gorm[M, V] {
	g := Gorm[M, V]{
		DB:      GetContext[any](c).GetDB(),
		config:  receiver.config,
		context: GetContext[any](c),
	}
	g.DB = g.GetModelDb()
	return &g
}

// SetDB 因为没有 echo.Context 所以只能手动设置
// 增加一定的维护性
func (receiver PreGorm[M, V]) SetDB(db *gorm.DB) *Gorm[M, V] {
	g := Gorm[M, V]{
		DB:     db,
		config: receiver.config,
	}
	g.DB = g.GetModelDb()
	return &g
}

type Gorm[M any, V any] struct {
	*gorm.DB
	context *XContext[any]
	config  InjectServiceConfig
}

func (r *Gorm[M, V]) SkipGlobalHook() *Gorm[M, V] {
	r.context.Set(GormGlobalSkipHookKey, true)
	return r
}

func (r *Gorm[M, V]) CheckHasField(column string) error {
	structure := r.createModelInstance()
	v := reflect.ValueOf(structure)
	t := reflect.TypeOf(structure)
	// 遍历结构体的字段
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("gorm")
		existStr := fmt.Sprintf(`column:%s`, LowerCamelCaseToSnake(column))
		if strings.Contains(tag, existStr) {
			return nil
		}
	}
	return NewFrontShowErrMsg("需要排序的字段不存在！")
}

func (r *Gorm[M, V]) FindOneByPrimaryKey(id int64) (error, M) {
	t, db := r.ModelAndDB()
	db = db.First(&t, id)
	if db.RowsAffected == 0 {
		return NewErrCode(CURD_DATA_NOT_EXIST_ERROR), t
	}
	return db.Error, t
}

func (r *Gorm[M, V]) FindOneVoByPrimaryKey(id int64) (error, V) {
	err, t := r.FindOneByPrimaryKey(id)
	if err != nil {
		return err, r.createViewInstance()
	}
	err, v := r.CopyViewFromModel(t)
	return err, v
}

func (r *Gorm[M, V]) FindOne(conditions ...func(*gorm.DB) *gorm.DB) (error, M) {
	db := r.DBWithConditions(conditions...)
	result := r.createModelInstance()
	count := int64(0)
	db.Count(&count)
	if count == 0 {
		return NewErrCode(CURD_DATA_NOT_EXIST_ERROR), result
	}
	err := db.First(&result).Error
	return err, result
}

func (r *Gorm[M, V]) FindOneVo(conditions ...func(*gorm.DB) *gorm.DB) (error, V) {
	err, t := r.FindOne(conditions...)
	if err != nil {
		return err, r.createViewInstance()
	}
	err, v := r.CopyViewFromModel(t)
	return err, v
}

func (r *Gorm[M, V]) FindListByPage(param PageParam, conditions ...func(*gorm.DB) *gorm.DB) (error, PageResultList[M]) {
	var result PageResultList[M]
	result.PageParam = param
	db := r.DBWithConditions(conditions...)
	resultList := r.ModelList()
	tx := db.Offset((param.Page - 1) * param.PageSize).Limit(param.PageSize).Find(&resultList)
	if tx.Error != nil {
		return tx.Error, result
	}
	result.Items = resultList
	if size := len(resultList); 0 < param.PageSize && 0 < size && size < param.PageSize {
		result.Total = int64(size + (param.Page-1)*param.PageSize)
		return nil, result
	}
	tx = db.Offset(-1).Limit(-1).Count(&result.Total)
	if tx.Error != nil {
		return tx.Error, result
	}
	return nil, result
}

func (r *Gorm[M, V]) FindList(conditions ...func(*gorm.DB) *gorm.DB) (error, []M) {
	db := r.DBWithConditions(conditions...)
	result := r.ModelList()
	db.Find(&result)
	return db.Error, result
}

func (r *Gorm[M, V]) FindVoList(conditions ...func(*gorm.DB) *gorm.DB) (error, []V) {
	err, i2 := r.FindList(conditions...)
	err, i3 := r.CopyViewListFromModelList(i2)
	if err != nil {
		return err, nil
	}
	return err, i3
}

func (r *Gorm[M, V]) FindVoListByPage(param PageParam, conditions ...func(*gorm.DB) *gorm.DB) (error, PageResultList[V]) {
	var result PageResultList[V]
	var err error
	result.PageParam = param
	db := r.DBWithConditions(conditions...)
	resultList := r.ModelList()
	err = db.Offset((param.Page - 1) * param.PageSize).Limit(param.PageSize).Find(&resultList).Error
	if err != nil {
		return err, result
	}
	err, result.Items = r.CopyViewListFromModelList(resultList)
	if err != nil {
		return err, result
	}
	if size := len(resultList); 0 < param.PageSize && 0 < size && size < param.PageSize {
		result.Total = int64(size + (param.Page-1)*param.PageSize)
		return nil, result
	}
	err = db.Offset(-1).Limit(-1).Count(&result.Total).Error
	if err != nil {
		return err, result
	}
	return nil, result
}

// UpdateByPrimaryKey   更新非零字段 false 0 "" 均不会被更新
func (r *Gorm[M, V]) UpdateByPrimaryKey(id int64, entity M) (error, int64) {
	db := r.GetModelDb()
	r.removePrimaryKey(&entity)
	tx := db.Where(fmt.Sprintf("%s = ?", r.config.PrimaryKeyField), id).
		Updates(entity)
	return tx.Error, tx.RowsAffected
}

// UpdateBy   更新非零字段 false 0 "" 均不会被更新
// 若指定了selectKey 则只会更新 selectKey的字段
func (r *Gorm[M, V]) UpdateBy(entity M, conditions func(*gorm.DB) *gorm.DB, selectKey ...string) (error, int64) {
	r.removePrimaryKey(&entity)
	db := r.DBWithConditions(conditions)
	tx := db.Updates(entity)
	return tx.Error, tx.RowsAffected
}

// SaveByPrimaryKey 更新所有字段 除了omitKeys
func (r *Gorm[M, V]) SaveByPrimaryKey(id int64, entity M, omitKey ...string) (error, int64) {
	db := r.GetModelDb()
	r.removePrimaryKey(&entity)
	var omitFields = []string{
		r.config.PrimaryKeyField,
		r.config.CreateDeptField,
		r.config.CreateByField,
		r.config.CreateTimeField,
		r.config.UpdateByField,
		r.config.DeleteTimeField}
	if len(omitFields) > 0 {
		omitFields = append(omitFields, omitKey...)
	}
	tx := db.Where(fmt.Sprintf("%s = ?", r.config.PrimaryKeyField), id).Select("*").
		Omit(omitFields...).
		Updates(entity)
	return tx.Error, tx.RowsAffected
}

func (r *Gorm[M, V]) DeleteByPrimaryKeys(ids []int64) (error, int64) {
	result, db := r.ModelAndDB()
	db.Delete(&result, ids)
	return db.Error, db.RowsAffected
}
func (r *Gorm[M, V]) DeleteBy(conditions ...func(*gorm.DB) *gorm.DB) (error, int64) {
	db := r.DBWithConditions(conditions...)
	db.Delete(nil)
	return db.Error, db.RowsAffected
}

func (r *Gorm[M, V]) Exist(conditions ...func(*gorm.DB) *gorm.DB) bool {
	return r.Count(conditions...) > 0
}

func (r *Gorm[M, V]) InsertOne(entity M) (error, M) {
	db := r.GetModelDb()
	r.removePrimaryKey(&entity)
	db.Create(&entity)
	if db.RowsAffected == 0 {
		return db.Error, entity
	}
	return nil, entity
}

func (r *Gorm[M, V]) InsertBatch(entities []M) (error, []M) {
	if len(entities) == 0 {
		return nil, entities
	}
	db := r.GetModelDb()
	for index := range entities {
		r.removePrimaryKey(&entities[index])
	}
	db.Create(&entities)
	if db.RowsAffected == 0 {
		return db.Error, entities
	}
	return db.Error, entities
}

func (r *Gorm[M, V]) Count(conditions ...func(*gorm.DB) *gorm.DB) int64 {
	db := r.DBWithConditions(conditions...)
	count := int64(0)
	db.Count(&count)
	return count
}

func (r *Gorm[M, V]) CopyViewFromModel(model M) (error, V) {
	view := r.createViewInstance()
	err := copier.CopyWithOption(&view, &model, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	return err, view
}

func (r *Gorm[M, V]) CopyViewListFromModelList(models []M) (error, []V) {
	views := r.ViewList(len(models))
	err := copier.CopyWithOption(&views, &models, copier.Option{IgnoreEmpty: true})
	return err, views
}

func (r *Gorm[M, V]) removePrimaryKey(s *M) {
	for _, key := range r.config.SpecialPrimaryKey {
		if HasField(*s, key) {
			_ = SetField(s, key, int64(0))
		}
	}
}

func (r *Gorm[M, V]) createModelInstance() (result M) {
	return Instance[M]()
}

func (r *Gorm[M, V]) createViewInstance() (result V) {
	return Instance[V]()
}

func (r *Gorm[M, V]) GetModelDb() *gorm.DB {
	var db = r.context.GetDB()
	result := r.createModelInstance()
	return db.WithContext(r.context).Model(&result)
}

func (r *Gorm[M, V]) GetDb() *gorm.DB {
	var db = r.context.GetDB()
	return db.WithContext(r.context)
}

func (r *Gorm[M, V]) ModelAndDB() (result M, db *gorm.DB) {
	return r.createModelInstance(), r.GetModelDb()
}

func (r *Gorm[M, V]) ViewAndDB() (result V, db *gorm.DB) {
	return r.createViewInstance(), r.GetModelDb()
}

func (r *Gorm[M, V]) ModelList() (result []M) {
	sliceType := reflect.SliceOf(reflect.TypeOf((*M)(nil)).Elem())
	sliceValue := reflect.MakeSlice(sliceType, 0, 0)
	result = sliceValue.Interface().([]M)
	return result
}

func (r *Gorm[M, V]) ViewList(len int) (result []V) {
	sliceType := reflect.SliceOf(reflect.TypeOf((*V)(nil)).Elem())
	sliceValue := reflect.MakeSlice(sliceType, len, len)
	result = sliceValue.Interface().([]V)
	return result
}

func (r *Gorm[M, V]) DBWithConditions(conditions ...func(*gorm.DB) *gorm.DB) (db *gorm.DB) {
	contextDb := r.GetModelDb()
	for _, condition := range conditions {
		contextDb = condition(contextDb)
	}
	return contextDb
}

func mergeInjectServiceDefaultConfig(config ...InjectServiceConfig) InjectServiceConfig {
	var defaultConfig = InjectServiceConfig{
		SpecialPrimaryKey: []string{"id", "UID"},
		PrimaryKeyField:   "id",
		CreateDeptField:   "create_dept",
		CreateByField:     "create_by",
		CreateTimeField:   "create_time",
		UpdateByField:     "update_by",
		UpdateTimeField:   "update_time",
		DeleteTimeField:   "delete_time",
		limitOne:          "limit 1",
	}
	// 如果传入了配置，使用最后一个配置项覆盖默认值
	if len(config) > 0 {
		lastConfig := config[len(config)-1]
		if lastConfig.PrimaryKeyField != "" {
			defaultConfig.PrimaryKeyField = lastConfig.PrimaryKeyField
		}
		if lastConfig.CreateDeptField != "" {
			defaultConfig.CreateDeptField = lastConfig.CreateDeptField
		}
		if lastConfig.CreateByField != "" {
			defaultConfig.CreateByField = lastConfig.CreateByField
		}
		if lastConfig.CreateTimeField != "" {
			defaultConfig.CreateTimeField = lastConfig.CreateByField
		}
		if lastConfig.UpdateByField != "" {
			defaultConfig.UpdateByField = lastConfig.UpdateByField
		}
		if lastConfig.UpdateTimeField != "" {
			defaultConfig.UpdateTimeField = lastConfig.UpdateTimeField
		}
		if lastConfig.DeleteTimeField != "" {
			defaultConfig.DeleteTimeField = lastConfig.DeleteTimeField
		}
	}
	return defaultConfig
}

type Array[T string | int32 | int8 | int64] []T

// Scan 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (a *Array[T]) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to scan Array value:", value))
	}
	if len(bytes) > 0 {
		return json.Unmarshal(bytes, a)
	}
	*a = make([]T, 0)
	return nil
}

// Value 实现 driver.Valuer 接口，Value 返回 json value
func (a Array[T]) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	return convertor.ToString(a), nil
}

type IntBool bool

// Value implements the driver.Valuer interface,
// and turns the IntBool into an integer for MySQL storage.
func (i IntBool) Value() (driver.Value, error) {
	if i {
		return int64(1), nil // true -> 1
	}
	return int64(2), nil // false -> 2
}

// Scan implements the sql.Scanner interface,
// and turns the int incoming from MySQL into an IntBool
func (i *IntBool) Scan(src interface{}) error {
	v, ok := src.(int64)
	if !ok {
		return errors.New("bad int type assertion")
	}
	*i = v == int64(1) // 1 -> true, otherwise false
	return nil
}
