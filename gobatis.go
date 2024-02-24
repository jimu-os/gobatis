package gobatis

import (
	"bytes"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"github.com/beevik/etree"
	"go.uber.org/zap"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
)

var banner = "  ______       ______             _      \n / _____)     (____  \\       _   (_)     \n| /  ___  ___  ____)  ) ____| |_  _  ___ \n| | (___)/ _ \\|  __  ( / _  |  _)| |/___)\n| \\____/| |_| | |__)  | ( | | |__| |___ |\n \\_____/ \\___/|______/ \\_||_|\\___)_(___/ \n"

const (
	MySQL = iota
	PostgreSQL
)

func New(db *sql.DB) *GoBatis {
	if db == nil {
		panic("db nil")
	}
	err := db.Ping()
	if err != nil {
		panic(err)
	}
	return &GoBatis{
		db:         reflect.ValueOf(db),
		NameSpaces: map[string]*Sql{},
		Logger:     logger,
	}
}

type GoBatis struct {
	*zap.Logger
	db reflect.Value
	// SqlSource 用于保存 xml 配置的文件的根路径配置信息，Build会通过SqlSource属性去加载 xml 文件
	SqlSource string
	// NameSpaces 保存了每个 xml 配置的根元素构建出来的 Sql 对象
	NameSpaces map[string]*Sql
	// mapper 文件加载
	mapperFS embed.FS
	// io 加载 mapper 文件
	mappers []io.Reader
	Type    int
}

// Logs 切换日志实例
func (batis *GoBatis) Logs(logger *zap.Logger) {
	batis.Logger = logger
}

// Source 加载 mapper文件
// source 应当是项目中的 mapper 文件根路径文件夹名称
func (batis *GoBatis) Source(source string) {
	if source != "" {
		batis.SqlSource = source
	}
	//fmt.Print(banner)
	// 解析 xml
	if batis.mapperFS == (embed.FS{}) && batis.SqlSource != "" {
		getwd, err := os.Getwd()
		if err != nil {
			return
		}
		root := filepath.Join(getwd, batis.SqlSource)
		filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
			if strings.HasSuffix(path, ".xml") {
				document := etree.NewDocument()
				err = document.ReadFromFile(path)
				if err != nil {
					return err
				}
				element := document.Root()
				attr := element.SelectAttr("namespace")
				if attr.Value == "" {
					batis.Warn("There is an empty namespace. Skip loading this mapper")
					return nil
				}
				s := NewSql(element)
				s.LoadSqlElement()
				batis.NameSpaces[attr.Value] = s
				batis.Debug("load mapper file path:[" + path + "]")
			}
			return nil
		})
		return
	}

	if batis.mapperFS != (embed.FS{}) {
		dir, err := batis.mapperFS.ReadDir(batis.SqlSource)
		if err != nil {
			panic(err)
		}
		batis.walk(batis.SqlSource, dir, batis.mapperFS, batis.NameSpaces)
	}

	if batis.mappers != nil {
		for _, mapperIo := range batis.mappers {
			document := etree.NewDocument()
			_, err := document.ReadFrom(mapperIo)
			if err != nil {
				panic(err)
			}
			element := document.Root()
			attr := element.SelectAttr("namespace")
			if attr.Value == "" {
				batis.Warn("There is an empty namespace. Skip loading this mapper")
				continue
			}
			s := NewSql(element)
			s.LoadSqlElement()
			batis.NameSpaces[attr.Value] = s
			batis.Debug("load mapper io namespace: " + attr.Value)
		}
	}
}

// Load 加载 mapper 静态文件
func (batis *GoBatis) Load(files embed.FS) {
	batis.mapperFS = files
}

// LoadByRootPath 根据提供的更路径及其 files 加载 mapper 文件
// {root:表示根路径,根路径应该和提供的 files 嵌入文件资源对应}
func (batis *GoBatis) LoadByRootPath(root string, files embed.FS) {
	dir, err := files.ReadDir(root)
	if err != nil {
		return
	}
	batis.walk(root, dir, files, batis.NameSpaces)
}

// ScanMappers 扫描解析
func (batis *GoBatis) ScanMappers(mappers ...any) {
	batis.Debug("Start scanning the mapper mapping function")
	group := sync.WaitGroup{}
	group.Add(len(mappers))
	for i := 0; i < len(mappers); i++ {
		go func(mapper any) {
			defer group.Done()
			vf := reflect.ValueOf(mapper)
			if vf.Kind() != reflect.Pointer {
				panic("")
			}
			if vf.Elem().Kind() != reflect.Struct {
				panic("")
			}
			vf = vf.Elem()
			namespace := vf.Type().String()
			namespace = Namespace(namespace)
			batis.Debug(fmt.Sprint("Starts loading the '" + namespace + "' mapping resolution"))
			wait := sync.WaitGroup{}
			wait.Add(vf.NumField())
			for j := 0; j < vf.NumField(); j++ {
				key := make([]string, 0)
				key = append(key, namespace)
				go func(field reflect.Value, structField reflect.StructField) {
					defer wait.Done()
					if !structField.IsExported() || structField.Type.Kind() != reflect.Func {
						return
					}
					// mapper 函数校验规范
					if flag, err := MapperCheck(field); !flag {
						batis.Panic(fmt.Sprint(namespace+"."+structField.Name, ",", field.Type().String(), ",", err.Error()))
					}
					key = append(key, structField.Name)
					batis.initMapper(key, field)
					fun := field.Type().String()
					index := strings.Index(fun, "(")
					fun = " " + fun[index:]
					batis.Debug(fmt.Sprint(namespace+"."+structField.Name, fun))
				}(vf.Field(j), vf.Type().Field(j))
			}
			wait.Wait()
		}(mappers[i])
	}
	group.Wait()
}

func (batis *GoBatis) get(id []string, value map[string]any) (string, string, string, []any, error) {
	if len(id) != 2 {
		return "", "", "", nil, errors.New("id error")
	}
	//ctx := toMap(value)
	if sql, b := batis.NameSpaces[id[0]]; b {
		if element, f := sql.Statement[id[1]]; f {
			analysis, tag, tempSql, params, err := Analysis(element, value)
			if err != nil {
				return "", "", "", nil, err
			}
			join := strings.Join(analysis, " ")
			temp := strings.Join(tempSql, " ")
			return join, tag, temp, params, nil
		}
	}
	return "", "", "", nil, fmt.Errorf("not found sql statement element")
}

// Analysis 解析xml标签
func Analysis(element *etree.Element, ctx map[string]any) ([]string, string, []string, []any, error) {
	var err error
	var t string
	var params []any
	args := make([]any, 0)
	SQL := make([]string, 0)
	template := make([]string, 0)
	// 解析根标签 开始之后的文本
	sqlStar := element.Text()
	// 处理字符串前后空格
	sqlStar = strings.TrimSpace(sqlStar)
	//更具标签类型，对应解析字符串
	sqlStar, t, params, err = Element(element, sqlStar, ctx)
	if err != nil {
		return nil, "", nil, nil, err
	}
	if sqlStar != "" && t != "" {
		SQL = append(SQL, sqlStar)
		template = append(template, t)
	}
	args = append(args, params...)
	// if 标签解析 逻辑不通过
	if sqlStar != "" && err == nil {
		// 解析子标签内容
		child := element.ChildElements()
		var childAnalysis, childTempSql []string
		var childParams []any
		for _, childElement := range child {
			childAnalysis, _, childTempSql, childParams, err = Analysis(childElement, ctx)
			if err != nil {
				return nil, "", childTempSql, params, fmt.Errorf("%s -> %s error,%s", element.Tag, childElement.Tag, err.Error())
			}
			SQL = append(SQL, childAnalysis...)
			template = append(template, childTempSql...)
			args = append(args, childParams...)
		}
		// 子节点解析内容为空 根据情况处理当前节点
		//if len(childAnalysis) <= 1 {
		//	SQL, template = sqlTag(element, SQL, template)
		//}
		SQL, template = sqlTag(element, SQL, template)
	}
	endSql := element.Tail()
	endSql = strings.TrimSpace(endSql)
	if endSql != "" {
		endSql, t, params, err = Element(element.Parent(), endSql, ctx)
		if err != nil {
			return nil, "", nil, nil, err
		}
		SQL = append(SQL, endSql)
		template = append(template, t)
		args = append(args, params...)
	}
	return SQL, element.Tag, template, args, nil
}

func Element(element *etree.Element, template string, ctx map[string]any) (string, string, []any, error) {
	// 检擦 节点标签类型
	tag := element.Tag
	switch tag {
	case For:
		return ForElement(element, template, ctx)
	case If:
		return IfElement(element, template, ctx)
	case Select, Update, Delete, Insert:
		return StatementElement(element, template, ctx)
	case Mapper:
		// 对根标签不做任何处理
		return "", "", nil, nil
	default:
		return keyTag(element, template, ctx)
	}
	//return "", "", nil, errors.New("error")
}

func Namespace(namespace string) string {
	if index := strings.LastIndex(namespace, "."); index != -1 {
		if end := strings.Index(namespace, "["); end != -1 {
			return namespace[index+1 : end]
		}
		return namespace[index+1:]
	}
	return namespace
}

// MapperCheck 检查 不同类别的sql标签 Mapper 函数是否符合规范
// 规则: 入参只能有一个并且只能是 map 或者 结构体，对返回值最后一个参数必须是error
func MapperCheck(fun reflect.Value) (bool, error) {
	// 至少有一个返回值
	if fun.Type().NumOut() < 1 {
		return false, errors.New("at least one return value is required")
	}
	// 只有一个参数返回时候，只能是 error
	if fun.Type().NumOut() == 1 {
		err := fun.Type().Out(0)
		if !err.Implements(reflect.TypeOf(new(error)).Elem()) {
			return false, errors.New("the second return value must be error")
		}
	}
	return true, nil
}

func (batis *GoBatis) walk(root string, list []fs.DirEntry, files embed.FS, NameSpaces map[string]*Sql) {
	for _, dirEntry := range list {
		path := filepath.Join(root, dirEntry.Name())
		path = filepath.ToSlash(path)
		if dirEntry.IsDir() {
			dir, err := files.ReadDir(path)
			if err != nil {
				panic(err)
			}
			batis.walk(path, dir, files, NameSpaces)
		}
		if strings.HasSuffix(path, ".xml") {
			b, err := files.ReadFile(path)
			if err != nil {
				panic(err)
			}
			buf := bytes.NewBuffer(b)
			document := etree.NewDocument()
			_, err = document.ReadFrom(buf)
			if err != nil {
				panic(err)
			}
			element := document.Root()
			attr := element.SelectAttr("namespace")
			s := NewSql(element)
			s.LoadSqlElement()
			NameSpaces[attr.Value] = s
			batis.Debug(fmt.Sprint("load mapper file path:[" + path + "]"))
		}
	}
}

func sqlTag(element *etree.Element, sqls, tmeplate []string) ([]string, []string) {
	if sqls == nil || len(sqls) == 0 {
		return sqls, tmeplate
	}
	switch element.Tag {
	case Where, WHERE, Value, Values, VALUE, VALUES:
		// 对最后一部分数据进行校验
		for i, s := range sqls {
			if s == Where || s == WHERE || s == Value || s == VALUE || s == Values || s == VALUES {
				if i == len(sqls)-1 {
					return sqls[:i], tmeplate[:i]
				}
			}
		}
		return sqls, tmeplate
	}
	return sqls, tmeplate
}
