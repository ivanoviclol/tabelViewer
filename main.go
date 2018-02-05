package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type Table struct {
	Header    []string
	Value     [][]string
	Soort     bool
	TableName string
}

var TableData Table
var Order string
var Destination string

func writer(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	fmt.Println(path, " func writer")
	Tables := "fail"
	var soort bool
	if path == "/" {
		http.Redirect(w, r, "/firma", 301)
	} else if path == "/firma" {
		Tables = "firma"
		soort = getTableSort(Tables)
	} else if path != "/favicon.ico" && path != "/" {
		Tables = strings.TrimPrefix(path, "/")
		soort = getTableSort(Tables)
	}
	t, err := template.ParseFiles("templates/index.html", "templates/header.html")
	checkErr(err)
	if Order != "" {
		TableData = constructTable(Tables, soort, Order, Destination)
		Order = ""
		Destination = ""
	} else {
		TableData = constructTable(Tables, soort, "", "")
	}
	t.ExecuteTemplate(w, "index", TableData)
}

func addANewColumn(w http.ResponseWriter, r *http.Request) {
	Tables := strings.TrimPrefix(r.Referer(), "http://localhost:3000/")
	TableData = constructTable(Tables, getTableSort(Tables), "", "")
	t, err := template.ParseFiles("templates/addANewColumn.html", "templates/header.html")
	checkErr(err)
	t.ExecuteTemplate(w, "write", TableData)
}

func newRecord(w http.ResponseWriter, r *http.Request) {
	Tables := strings.TrimPrefix(r.Referer(), "http://localhost:3000/")
	TableData = constructTable(Tables, getTableSort(Tables), "", "")
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/project")
	checkErr(err)
	err = db.Ping()
	checkErr(err)
	valuess := [][]string{}
	text := "SELECT `COLUMN_NAME` FROM `INFORMATION_SCHEMA`.`COLUMNS` WHERE `TABLE_SCHEMA`='project' AND `TABLE_NAME`='" + TableData.TableName + "';"
	rows, err := db.Query(text)
	value := []string{}
	for rows.Next() {
		var COLUMN_NAME string
		err = rows.Scan(&COLUMN_NAME)
		checkErr(err)
		value = append(value, COLUMN_NAME)
	}
	valuess = append(valuess, value)
	value = []string{}
	text = "SELECT `COLUMN_TYPE` FROM `INFORMATION_SCHEMA`.`COLUMNS` WHERE `TABLE_SCHEMA`='project' AND `TABLE_NAME`='" + TableData.TableName + "';"
	rows, err = db.Query(text)
	for rows.Next() {
		var COLUMN_NAME string
		err = rows.Scan(&COLUMN_NAME)
		checkErr(err)
		value = append(value, COLUMN_NAME)
	}
	valuess = append(valuess, value)
	value = []string{}
	text = "Insert into `" + TableData.TableName + "` ("
	for i := 0; i < len(valuess[0]); i++ {
		value = append(value, r.FormValue(valuess[0][i]))
		text = text + "`" + valuess[0][i] + "`"
		if i == len(valuess[0])-1 {
			text = text + ") values (\""
		} else {
			text = text + ", "
		}
	}
	valuess = append(valuess, value)
	for i := 0; i < len(valuess[2]); i++ {
		if TableData.Soort == true && TableData.TableName != "firma" && TableData.TableName != "firmaTemplates" && strings.HasPrefix(TableData.TableName, "TemplateOf") != true {
			if i == 0 {
				text = text + valuess[2][i] + strconv.Itoa(getFirmaCode(TableData.TableName))
			} else {
				text = text + valuess[2][i]
			}
		} else if TableData.TableName == "firmaTemplates" {
			text = text + "TemplateOf" + valuess[2][i]
		} else if strings.HasPrefix(TableData.TableName, "TemplateOf") && getTableComment(TableData.TableName) != "Table" {
			text = text + TableData.TableName + "-" + valuess[2][i]
		} else {
			text = text + valuess[2][i]
		}
		if i == len(valuess[2])-1 {
			text = text + "\");"
		} else {
			text = text + "\", \""
		}
	}
	fmt.Println(text)
	_, err = db.Exec(text)
	checkErr(err)
	if TableData.Soort == true && TableData.TableName == "firma" {
		text = "Create table `" + valuess[2][0] + "` like Default_Firma;"
		_, err = db.Exec(text)
	} else if TableData.Soort == true && TableData.TableName != "firma" && TableData.TableName != "firmaTemplates" && strings.HasPrefix(TableData.TableName, "TemplateOf") != true {
		text = "Create table `" + valuess[2][0] + strconv.Itoa(getFirmaCode(TableData.TableName)) + "` like Default_Table;"
		_, err = db.Exec(text)
	} else if TableData.TableName == "firmaTemplates" {
		text = "Create table `TemplateOf" + valuess[2][0] + "` like Default_Firma;"
		_, err = db.Exec(text)
	} else if strings.HasPrefix(TableData.TableName, "TemplateOf") && getTableComment(TableData.TableName) == "Firma" {
		text = "Create table `" + TableData.TableName + "-" + valuess[2][0] + "` like Default_Table;"
		_, err = db.Exec(text)
	}
	checkErr(err)
	http.Redirect(w, r, getTrueUrl(), 301)
}

func saveColumn(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	vartype := r.FormValue("type")
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/project")
	checkErr(err)
	err = db.Ping()
	checkErr(err)
	switch vartype {
	case "TEXT":
		vartype = vartype + " not null;"
		defer db.Exec("UPDATE " + TableData.TableName + " Set " + name + " = \"\";")
	case "TINYTEXT", "VARCHAR(64)":
		vartype = vartype + " not null;"
	case "INTEGER":
		vartype = vartype + " not null;"
	case "DOUBLE":
		vartype = vartype + " not null;"
	}
	text := "ALTER TABLE `" + TableData.TableName + "` ADD `" + name + "` " + vartype
	_, err = db.Exec(text)
	checkErr(err)
	http.Redirect(w, r, getTrueUrl(), 301)
}

func deleteColumn(w http.ResponseWriter, r *http.Request) {
	columnName := r.FormValue("columnName")
	Table := strings.TrimPrefix(r.Referer(), "http://localhost:3000/")
	if columnName == "Name" || columnName == "FirmaCode" || columnName == "RecordId" {
		http.Redirect(w, r, "/"+Table, 301)
	} else {
		text := "ALTER TABLE " + Table + " DROP COLUMN `" + columnName + "`;"
		db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/project")
		_, err = db.Exec(text)
		checkErr(err)
		http.Redirect(w, r, "/"+Table, 301)
	}
}

func shuffleUp(w http.ResponseWriter, r *http.Request) {
	Order = r.FormValue("columnName")
	Table := strings.TrimPrefix(r.Referer(), "http://localhost:3000/")
	Destination = "ASC"
	http.Redirect(w, r, "/"+Table, 301)
}

func shuffleDown(w http.ResponseWriter, r *http.Request) {
	Order = r.FormValue("columnName")
	Table := strings.TrimPrefix(r.Referer(), "http://localhost:3000/")
	Destination = "DESC"
	http.Redirect(w, r, "/"+Table, 301)
}

func deleteRecord(w http.ResponseWriter, r *http.Request) {
	recordName := r.FormValue("name")
	Table := strings.TrimPrefix(r.Referer(), "http://localhost:3000/")
	text := ""
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/project")
	switch getTableComment(Table) {
	case "Firmas":
		text = "SELECT NAME from `" + recordName + "`;"
		rows, err := db.Query(text)
		for rows.Next() {
			var COLUMN_NAME string
			err = rows.Scan(&COLUMN_NAME)
			checkErr(err)
			text = "DROP TABLE `" + COLUMN_NAME + "`;"
			_, err = db.Exec(text)
			checkErr(err)
		}
		text = "DELETE FROM `" + Table + "` WHERE Name= '" + recordName + "';"
		_, err = db.Exec(text)
		checkErr(err)
		text = "DROP TABLE `" + recordName + "`;"
	case "Firma":
		text = "DELETE FROM `" + Table + "` WHERE Name= '" + recordName + "';"
		_, err = db.Exec(text)
		checkErr(err)
		text = "DROP TABLE `" + recordName + "`;"
	case "Table":
		text = "DELETE FROM `" + Table + "` WHERE Name= '" + recordName + "';"
	}
	_, err = db.Exec(text)
	checkErr(err)
	http.Redirect(w, r, "/"+Table, 301)
}

func editRecord(w http.ResponseWriter, r *http.Request) {
	Tables := strings.TrimPrefix(r.Referer(), "http://localhost:3000/")
	TableData = constructTable(Tables, getTableSort(Tables), "", "")
	t, err := template.ParseFiles("templates/edit.html", "templates/header.html")
	checkErr(err)
	type Data struct {
		Header string
		Value  string
	}

	type Record struct {
		Name      string
		TableName string
		MyData    []Data
	}

	var name string
	var headers []string
	var values []string
	var record Record
	var myDataList []Data
	for j := 0; j < len(TableData.Value); j++ {

		if TableData.Value[j][0] == r.FormValue("name") {
			name = TableData.Value[j][0]
			i := 1
			if TableData.Header[1] == "FirmaCode" {
				i = 2
			}
			values = TableData.Value[j][i:]
			headers = TableData.Header[i:]

			for a := 0; a < len(values); a++ {
				data := Data{headers[a], values[a]}
				myDataList = append(myDataList, data)
			}

		}
	}
	record = Record{name, TableData.TableName, myDataList}
	t.ExecuteTemplate(w, "edit", record)
}

func saveRecord(w http.ResponseWriter, r *http.Request) {
	tableName := r.FormValue("tableName")
	recordName := r.FormValue("recordName")
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/project")
	checkErr(err)
	text := "SELECT `COLUMN_NAME` FROM `INFORMATION_SCHEMA`.`COLUMNS` WHERE `TABLE_SCHEMA`='project' AND `TABLE_NAME`='" + tableName + "';"
	rows, err := db.Query(text)
	headers := []string{}
	for rows.Next() {
		var COLUMN_NAME string
		err = rows.Scan(&COLUMN_NAME)
		checkErr(err)
		if COLUMN_NAME != "Name" && COLUMN_NAME != "FirmaCode" {
			headers = append(headers, COLUMN_NAME)
		}
	}
	values := []string{}
	for i := 0; i < len(headers); i++ {
		values = append(values, r.FormValue(headers[i]))
	}

	text = "Update `" + tableName + "` SET `"
	for i := 0; i < len(headers); i++ {
		text = text + headers[i] + "` = '" + values[i] + "'"
		if i != len(headers)-1 {
			text = text + ", `"
		} else {
			text = text + " "
		}
	}
	text = text + "Where Name = '" + recordName + "';"
	_, err = db.Exec(text)
	checkErr(err)
	http.Redirect(w, r, "/"+tableName, 301)
}

func useTemplate(w http.ResponseWriter, r *http.Request) {
	Tables := strings.TrimPrefix(r.Referer(), "http://localhost:3000/")
	t, err := template.ParseFiles("templates/createFromTemplate.html", "templates/header.html")
	checkErr(err)
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/project")
	checkErr(err)
	type dataFromTemplate struct {
		TableName     string
		TemplatesType string
		Options       []string
	}
	var templatesType string
	if getTableComment(Tables) == "Firmas" {
		templatesType = "firmaTemplates"
	} else if getTableComment(Tables) == "Firma" {
		templatesType = "tableTemplates"
	}
	text := "SELECT `Name` from `" + templatesType + "`;"

	rows, err := db.Query(text)
	templateNames := []string{}
	for rows.Next() {
		var COLUMN_NAME string
		err = rows.Scan(&COLUMN_NAME)
		checkErr(err)
		templateNames = append(templateNames, COLUMN_NAME)
	}
	data := dataFromTemplate{Tables, templatesType, templateNames}
	t.ExecuteTemplate(w, "createFromTemplate", data)
}

func usingTemplate(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/project")
	checkErr(err)
	if r.FormValue("templatesType") == "tableTemplates" {
		text := "CREATE TABLE `" + r.FormValue("NewTable") + strconv.Itoa(getFirmaCode(r.FormValue("tableName"))) + "`LIKE `" + r.FormValue("templateName") + "`;"
		_, err = db.Exec(text)
		checkErr(err)
		text = "INSERT `" + r.FormValue("NewTable") + strconv.Itoa(getFirmaCode(r.FormValue("tableName"))) + "` SELECT * FROM `" + r.FormValue("templateName") + "`;"
		_, err = db.Exec(text)
		checkErr(err)
		text = "INSERT INTO `" + r.FormValue("tableName") + "` (Name) VALUES ('" + r.FormValue("NewTable") + strconv.Itoa(getFirmaCode(r.FormValue("tableName"))) + "');"
		_, err = db.Exec(text)
		checkErr(err)

	} else if r.FormValue("templatesType") == "firmaTemplates" {
		text := "CREATE TABLE `" + r.FormValue("NewTable") + "` LIKE `" + r.FormValue("templateName") + "`;"
		_, err = db.Exec(text)
		checkErr(err)
		text = "INSERT INTO `firma` (Name) values ('" + r.FormValue("NewTable") + "');"
		_, err = db.Exec(text)
		checkErr(err)
		text = "SELECT `Name` FROM `" + r.FormValue("templateName") + "`;"
		rows, err := db.Query(text)
		checkErr(err)
		for rows.Next() {
			var COLUMN_NAME string
			err = rows.Scan(&COLUMN_NAME)
			checkErr(err)
			var name string
			name = strings.TrimLeft(COLUMN_NAME, r.FormValue("templateName")+"-")
			text = "CREATE TABLE `" + name + strconv.Itoa(getFirmaCode(r.FormValue("NewTable"))) + "` like `" + COLUMN_NAME + "`;"
			_, err = db.Exec(text)
			checkErr(err)
			text = "INSERT `" + name + strconv.Itoa(getFirmaCode(r.FormValue("NewTable"))) + "` SELECT * FROM `" + COLUMN_NAME + "`;"
			_, err = db.Exec(text)
			checkErr(err)
			text = "INSERT INTO `" + r.FormValue("NewTable") + "` (Name) VALUES ('" + name + strconv.Itoa(getFirmaCode(r.FormValue("NewTable"))) + "');"
			_, err = db.Exec(text)
			checkErr(err)
		}
	}
	http.Redirect(w, r, "/"+r.FormValue("tableName"), 301)
}

////////////////////////////////////////////////////////////////////////////////

func reshuffle(values [][]string) [][]string {
	output := [][]string{}
	if len(values) != 0 {
		for i := 0; i < len(values[0]); i++ {
			temp := []string{}
			for j := 0; j < len(values); j++ {
				temp = append(temp, values[j][i])
			}
			output = append(output, temp)
		}
	}
	return output
}

func constructTable(tableName string, soort bool, order string, direction string) Table {
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/project")
	checkErr(err)
	err = db.Ping()
	checkErr(err)
	headers := []string{}
	text := "SELECT `COLUMN_NAME` FROM `INFORMATION_SCHEMA`.`COLUMNS` WHERE `TABLE_SCHEMA`='project' AND `TABLE_NAME`='" + tableName + "';"
	rows, err := db.Query(text)
	for rows.Next() {
		var COLUMN_NAME string
		err = rows.Scan(&COLUMN_NAME)
		checkErr(err)
		headers = append(headers, COLUMN_NAME)
	}
	rows.Close()
	valuess := [][]string{}
	for _, value := range headers {
		text := ""
		if order != "" {
			text = "SELECT `" + value + "` FROM `" + tableName + "` ORDER BY `" + order + "` " + direction + ";"
		} else {
			text = "SELECT `" + value + "` FROM `" + tableName + "`;"
		}
		rows, err := db.Query(text)
		foo := []string{}
		for rows.Next() {
			var COLUMN_NAME string
			err = rows.Scan(&COLUMN_NAME)
			checkErr(err)
			foo = append(foo, COLUMN_NAME)
		}
		valuess = append(valuess, foo)
	}
	valuess = reshuffle(valuess)
	TableData = Table{headers, valuess, soort, tableName}
	rows.Close()
	return TableData
}

func getTableSort(tableName string) bool {
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/project")
	result := false
	text := "SELECT `Table_Comment` FROM `INFORMATION_SCHEMA`.`tables` WHERE `TABLE_SCHEMA`='project' and `Table_Name` = '" + tableName + "' ;"
	rows, err := db.Query(text)
	for rows.Next() {
		var Table_Comment string
		err = rows.Scan(&Table_Comment)
		checkErr(err)
		if Table_Comment == "Firma" || Table_Comment == "Firmas" {
			result = true
		} else {
			result = false
		}
	}
	return result
}

func getTableComment(tableName string) string {
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/project")
	result := ""
	text := "SELECT `Table_Comment` FROM `INFORMATION_SCHEMA`.`tables` WHERE `TABLE_SCHEMA`='project' and `Table_Name` = '" + tableName + "' ;"
	rows, err := db.Query(text)
	for rows.Next() {
		var Table_Comment string
		err = rows.Scan(&Table_Comment)
		checkErr(err)
		result = Table_Comment
	}
	return result
}

func getFirmaCode(tableName string) int {
	tableCode := 0
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/project")
	checkErr(err)
	text := "Select `FirmaCode` from `firma` Where `Name` = '" + tableName + "';"
	rows, err := db.Query(text)
	for rows.Next() {
		err = rows.Scan(&tableCode)
	}
	return tableCode
}

func getTrueUrl() string {
	urlStr := ""
	if TableData.TableName == "firma" {
		urlStr = "/"
	} else {
		urlStr = "/" + TableData.TableName
	}
	return urlStr
}

func main() {

	http.HandleFunc("/", writer)
	http.HandleFunc("/addingANewColumn", addANewColumn)
	http.HandleFunc("/SaveColumn", saveColumn)
	http.HandleFunc("/NewRecord", newRecord)
	http.HandleFunc("/deleteColumn", deleteColumn)
	http.HandleFunc("/shuffleUp", shuffleUp)
	http.HandleFunc("/shuffleDown", shuffleDown)
	http.HandleFunc("/deleteRecord", deleteRecord)
	http.HandleFunc("/editRecord", editRecord)
	http.HandleFunc("/saveRecord", saveRecord)
	http.HandleFunc("/useTemplate", useTemplate)
	http.HandleFunc("/usingTemplate", usingTemplate)
	//Giving access to css
	http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("./templates/"))))
	// Listening on port 3000
	http.ListenAndServe(":3000", nil)
	fmt.Printf("We are running")
	//Starting func writer
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
