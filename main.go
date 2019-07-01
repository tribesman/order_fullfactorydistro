package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/tealeg/xlsx"
)

var parsed []string
var productsURL []string
var pagesArr []string
var currPage string

// Config Настройки
type Config struct {
	Login    string
	Password string
	Xlsx     string
	Log      string
	LogCSV   string
	Name     string
	Mod      string
	Msrp     string
	URL      string
	Total    string
}

// Columns - сопоставление колонок с данными
type Columns struct {
	name  int
	mod   int
	msrp  int
	url   int
	total int
}

// Products массив товаров
//var Products []Product

// Product Товар
type Product struct {
	name  string
	mod   string
	msrp  string
	url   string
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

	csvFile, err := os.Create(dir + "/" + settings.LogCSV)
	checkError("Cannot create file", err)
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	line := []string{"Name", "Option", "MSRP", "URL", "Total", ""}
	_ = csvWriter.Write(line)
	defer csvWriter.Flush()

	// 1 Сопоставляем номера колонок с данными
	var columns Columns
	getColumns(dir+"/"+settings.Xlsx, &settings, &columns)

	// Заполняем массив товаров
	p := getProducts(dir+"/"+settings.Xlsx, &columns)
	for _, row := range p {
		line := []string{row.name, row.mod, row.msrp, row.url, strconv.Itoa(row.total)}
		err := csvWriter.Write(line)
		checkError("Cannot write to file", err)
	}
	fmt.Printf("len=%d cap=%d %v\n", len(p), cap(p), p)

	// test scrap product
	//c.Visit(productsURL[0])
	fulltimeEnd := time.Now()
	myLogger(logFile, fmt.Sprintf("\n---\nElapsed time: %v\n", fulltimeEnd.Sub(fulltimeStart)), debug)
}

// Сопоставляем номера колонок
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
				if strings.ToLower(text) == strings.ToLower(settings.URL) {
					columns.url = cellNum
				}
				if strings.ToLower(text) == strings.ToLower(settings.Total) && columns.total == 0 {
					columns.total = cellNum
				}
				cellNum = cellNum + 1

			}
		}
	}
	fmt.Printf("%v\n", columns)
}

// Заполнение массива товаров
func getProducts(file string, columns *Columns) (p []Product) {
	xlFile, err := xlsx.OpenFile(file)
	if err != nil {
		checkError(file+" is not open", err)
	}
	for _, sheet := range xlFile.Sheets {
		for _, row := range sheet.Rows {

			// Заполняем не более 15 строк
			//if len(p) > 15 {
			//	break
			//}
			var product Product
			product.name = row.Cells[columns.name].String()
			product.mod = row.Cells[columns.mod].String()
			product.url = row.Cells[columns.url].String()
			product.msrp = row.Cells[columns.msrp].String()
			product.total, _ = row.Cells[columns.total].Int()
			if product.name != "" && product.url != "" && product.total > 0 {
				p = append(p, product)
			}
		}
	}
	return
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
