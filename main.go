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
	"github.com/gocolly/colly"
	"github.com/tealeg/xlsx"
)

// Config Настройки
type Config struct {
	Login    string
	Password string
	LoginUrl string
	HomeUrl  string
	CartUrl  string
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

	// Добавляем товары в корзину

	putToCart(p, &settings)

	// Сохраняем товары в cvs для проверки
	for _, row := range p {
		line := []string{row.name, row.mod, row.msrp, row.url, strconv.Itoa(row.total)}
		err := csvWriter.Write(line)
		checkError("Cannot write to file", err)
	}

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

// Кладем товар в корзину
func putToCart(p []Product, settings *Config) {

	ex, _ := os.Executable()
	dir := filepath.Dir(ex)

	loginCount := 0     // кол-во попыток авторизации
	isLogin := false    // скрипт еще не залогинен
	isCartPage := false // корзина
	i := 1              // Кол-во строк для обработки

	// Instantiate default collector
	c := colly.NewCollector(
		// Visit only domains: hackerspaces.org, wiki.hackerspaces.org
		colly.AllowedDomains("fullfactorydistro.com"),
	)

	c.OnError(func(r *colly.Response, e error) {
		fmt.Println("error:", e, r.Request.URL, string(r.Body))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("\nVisiting: %s\n", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		isCartPage = false
		if strings.Contains(r.Request.URL.String(), "account") {
			log.Println("LOGGIN", r.Request.URL, r.StatusCode)
			r.Save(dir + "/login.html")
		} else if strings.Contains(r.Request.URL.String(), settings.CartUrl) {
			log.Println("Корзина", r.Request.URL, r.StatusCode)
			r.Save(dir + "/cart.html")
			isCartPage = true
		} else {
			r.Save(dir + "/body_" + strconv.Itoa(i) + ".html")
		}
	})

	//<a href="/cart/change?line=18&amp;quantity=0" class="btn">Remove</a>
	c.OnHTML("a[href].btn", func(e *colly.HTMLElement) {
		//fmt.Println("Очистка корзины", isCartPage)
		if isCartPage == true {
			// Очистика корзины
			link := e.Attr("href")
			if strings.Contains(link, "/cart/change?line=") {
				fmt.Println(link)
			}

		}
	})

	// Авторизация
	c.OnHTML("a[href='/account/login']", func(e *colly.HTMLElement) {
		if loginCount < 5 && isLogin == false {
			log.Printf("NEED LOGIN %s\n", e.Request.URL.String())

			// authenticate
			err := c.Post(settings.LoginUrl, map[string]string{"customer[email]": settings.Login, "customer[password]": settings.Password, "form_type": "customer_login", "utf8": "true"})
			if err != nil {
				log.Fatal(err)
			}
			loginCount = loginCount + 1
			isLogin = true

		} else {
			//fmt.Printf("error: can not login\n")
			//os.Exit(1)
		}
	})

	// Ищем товар и добавляем его в корзину

	//c.OnHTML("html", func(e *colly.HTMLElement) {

	//err = ioutil.WriteFile("body.html", e.Response.Body, 0644)
	//})

	c.Visit(settings.HomeUrl) // Сначала авторизируемся
	c.Visit(settings.CartUrl) // Очистка корзины
	//for _, row := range p {
	//	fmt.Printf("%v\n", row)
	//	c.Visit(row.url)
	//	i = i - 1
	//	if i == 0 {
	//		break
	//	}
	//}

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
