package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/tealeg/xlsx"
)

type Product struct {
	name        string
	msrp        string
	description string
	options     []Option
}
type Option struct {
	name  string
	price string
	stock string
}

var Products []Product

var parsed []string
var productsURL []string
var pagesArr []string
var currPage string

// Настройки
type Config struct {
	Login    string
	Password string
	Xlsx     string
	Log      string
	Name     string
	Mod      string
	Msrp     string
	Url      string
	Total    string
}

type Columns struct {
	name  int
	mod   int
	msrp  int
	url   int
	total int
}

func main() {

	fulltimeStart := time.Now()

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	dir := filepath.Dir(ex)
	fmt.Println(dir)

	// 0 Без логов
	// 1 Только в файл
	// 2 В файл и консоль
	debug := 2
	//isProductPage := false

	// Load settings
	var settings Config

	_, err = toml.DecodeFile(dir+"/config.toml", &settings)
	checkError("Config", err)
	fmt.Println(settings)

	logFile, err := os.Create(dir + "/" + settings.Log)
	checkError("Cannot create file", err)
	defer logFile.Close()

	// 1 Сопоставляем номера колонок с данными
	var columns Columns
	getColumns(dir+"/"+settings.Xlsx, &settings, &columns)

	// test scrap product
	//c.Visit(productsURL[0])
	fulltimeEnd := time.Now()
	myLogger(logFile, fmt.Sprintf("\n---\nElapsed time: %v\n", fulltimeEnd.Sub(fulltimeStart)), debug)
}
func getColumns(file string, settings *Config, columns *Columns) {
	xlFile, err := xlsx.OpenFile(file)
	if err != nil {
		checkError(file+" is not open", err)
	}
	rows := 0
	for _, sheet := range xlFile.Sheets {
		for _, row := range sheet.Rows {

			// Считываем не более 15 строк
			rows = rows + 1
			if rows > 15 {
				break
			}
			cellNum := 0
			for _, cell := range row.Cells {
				text := cell.String()
				if strings.ToLower(text) == strings.ToLower(settings.Name) {
					columns.name = cellNum
				}
				if strings.ToLower(text) == strings.ToLower(settings.Mod) {
					columns.mod = cellNum
				}
				if strings.ToLower(text) == strings.ToLower(settings.Msrp) {
					columns.msrp = cellNum
				}
				if strings.ToLower(text) == strings.ToLower(settings.Url) {
					columns.url = cellNum
				}
				if strings.ToLower(text) == strings.ToLower(settings.Total) {
					columns.total = cellNum
				}
				cellNum = cellNum + 1

			}
		}
	}
	fmt.Printf("%v\n", columns)

}
func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func myLogger(f *os.File, s string, debug int) {
	if debug == 1 {
		f.WriteString(s)
	}
	if debug == 2 {
		fmt.Print(s)
		f.WriteString(s)
	}
}
