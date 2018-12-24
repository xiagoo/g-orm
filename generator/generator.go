package generator

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/mijia/modelq/drivers"
)

type CodeConfig struct {
	PackageName string
	Template    string
}

type ModelField struct {
	Name            string
	ColumnName      string
	Type            string
	Tag             string
	IsPrimaryKey    bool
	IsUniqueKey     bool
	IsAutoIncrement bool
	DefaultValue    string
	Extra           string
	Comment         string
}

type ModelMeta struct {
	Name         string
	LowerName    string
	DbName       string
	TableName    string
	PrimaryField *ModelField
	Fields       []ModelField
	Uniques      []ModelField
	config       CodeConfig
}

func (m ModelMeta) getTemplate(tmpl *template.Template, name string, defaultTmpl *template.Template) *template.Template {
	if tmpl != nil {
		if definedTmpl := tmpl.Lookup(name); definedTmpl != nil {
			return definedTmpl
		}
	}
	return defaultTmpl
}

func (m ModelMeta) GenHeader(w *bufio.Writer, tmpl *template.Template, importTime bool) error {
	return m.getTemplate(tmpl, "header", tmHeader).Execute(w, map[string]interface{}{
		"DbName":     m.DbName,
		"TableName":  m.TableName,
		"PkgName":    m.config.PackageName,
		"ImportTime": importTime,
	})
}

func (m ModelMeta) GenStruct(w *bufio.Writer, tmpl *template.Template) error {
	return m.getTemplate(tmpl, "struct", tmStruct).Execute(w, m)
}

func (m ModelMeta) GenObjectApi(w *bufio.Writer, tmpl *template.Template) error {
	return m.getTemplate(tmpl, "obj_api", tmObjApi).Execute(w, m)
}

func GenerateModels(dbName string, dbSchema drivers.DbSchema, config CodeConfig) {
	for tName, schema := range dbSchema {
		generateModel(dbName, tName, schema, config, nil)
	}
}

func generateModel(dbName, tName string, schema drivers.TableSchema, config CodeConfig, tmpl *template.Template) error {
	file, err := os.Create(path.Join(config.PackageName, tName+".go"))
	if err != nil {
		return err
	}
	w := bufio.NewWriter(file)
	defer func() {
		w.Flush()
		file.Close()
	}()

	model := ModelMeta{
		Name:      toCapitalCase(tName, true),
		LowerName: toCapitalCase(tName, false),
		DbName:    dbName,
		TableName: tName,
		Fields:    make([]ModelField, len(schema)),
		Uniques:   make([]ModelField, 0, len(schema)),
		config:    config,
	}

	for i, col := range schema {
		field := ModelField{
			Name:            toCapitalCase(col.ColumnName, true),
			ColumnName:      col.ColumnName,
			Type:            col.DataType,
			Tag:             "",
			IsPrimaryKey:    strings.ToUpper(col.ColumnKey) == "PRI",
			IsUniqueKey:     strings.ToUpper(col.ColumnKey) == "UNI",
			IsAutoIncrement: strings.ToUpper(col.Extra) == "AUTO_INCREMENT",
			DefaultValue:    col.DefaultValue,
			Extra:           col.Extra,
			Comment:         col.Comment,
		}

		gormTag := ""
		tagArr := make([]string, 0)
		jsonTag := fmt.Sprintf("json:\"%s\"", col.ColumnName)
		tagArr = append(tagArr, jsonTag)
		if field.IsPrimaryKey {
			model.PrimaryField = &field
		}

		if col.ColumnName == "created_at" || col.ColumnName == "updated_at" {
			gormTag = "gorm:\"-\""
		} else {
			gormTag = fmt.Sprintf("gorm:\"column:%s\"", col.ColumnName)
		}
		tagArr = append(tagArr, gormTag)

		field.Tag = fmt.Sprintf("`%s`", strings.Join(tagArr, ","))
		model.Fields[i] = field
	}
	if err := model.GenHeader(w, tmpl, true); err != nil {
		return fmt.Errorf("[%s] Fail to gen model header, %s", tName, err)
	}
	if err := model.GenStruct(w, tmpl); err != nil {
		return fmt.Errorf("[%s] Fail to gen model struct, %s", tName, err)
	}
	if err := model.GenObjectApi(w, tmpl); err != nil {
		return fmt.Errorf("[%s] Fail to gen model object api, %s", tName, err)
	}
	return nil
}

func toCapitalCase(name string, firstLetterUpper bool) string {
	data := []byte(name)
	segStart := true
	endPos := 0
	isFirst := true
	lastUnderScore := false
	for i := 0; i < len(data); i++ {
		ch := data[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			if segStart {
				if ch >= 'a' && ch <= 'z' {
					if !isFirst || firstLetterUpper {
						ch = ch - 'a' + 'A'
					}
				}
				segStart = false
			} else {
				if ch >= 'A' && ch <= 'Z' {
					ch = ch - 'A' + 'a'
				}
			}
			data[endPos] = ch
			lastUnderScore = false
			endPos++
		} else if ch >= '0' && ch <= '9' {
			if lastUnderScore {
				data[endPos] = "_"[0]
				endPos++
			}
			data[endPos] = ch
			endPos++
			segStart = true
			lastUnderScore = false
		} else {
			lastUnderScore = true
			segStart = true
		}
		isFirst = false
	}
	return string(data[:endPos])
}
