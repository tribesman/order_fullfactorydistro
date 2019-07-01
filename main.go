package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gocolly/colly"
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
	isProductPage := false

	// Load settings
	var settings Config
	_, err = toml.DecodeFile(dir+"/config.toml", &settings)
	checkError("Config", err)
	fmt.Println(settings)

	csvFile, err := os.Create(dir + "/products.csv")
	checkError("Cannot create file", err)
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	line := []string{"Name", "Option", "MSRP", "PRICE", "Stock", "URL"}
	_ = csvWriter.Write(line)
	defer csvWriter.Flush()

	logFile, err := os.Create(dir + "/log.txt")
	checkError("Cannot create file", err)
	defer logFile.Close()

	// Make a Regex to say we only want
	regNumbers, err := regexp.Compile("[^\\d\\.\\,]+")
	if err != nil {
		log.Fatal(err)
	}

	regNoWhiteSpaces, err := regexp.Compile(`[\t\n\r]+|[\s]{2,}`)
	if err != nil {
		log.Fatal(err)
	}

	// index pages
	pagesArr = append(pagesArr, "https://fullfactorydistro.com/collections/odyssey?sort_by=title-ascending")
	pagesArr = append(pagesArr, "https://fullfactorydistro.com/collections/gsport?sort_by=title-ascending")
	pagesArr = append(pagesArr, "https://fullfactorydistro.com/collections/sunday?sort_by=title-ascending")
	pagesArr = append(pagesArr, "https://fullfactorydistro.com/collections/fairdale?sort_by=title-ascending")

	loginPage := "https://fullfactorydistro.com/account/login"
	login := false

	// Instantiate default collector
	c := colly.NewCollector(
		// Visit only domains: hackerspaces.org, wiki.hackerspaces.org
		colly.AllowedDomains("fullfactorydistro.com"),
	)

	// Collect links to products
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := strings.TrimSpace(e.Attr("href"))
		link = strings.Trim(link, " ")

		if strings.Contains(link, "products") && strings.Contains(link, "collections") {
			productsURL = append(productsURL, e.Request.AbsoluteURL(link))
		}
	})

	// Next page
	c.OnHTML("a[href].next", func(e *colly.HTMLElement) {
		link := strings.Trim(e.Attr("href"), " ")
		inArray, _ := in_array(e.Request.AbsoluteURL(link), pagesArr)
		if !inArray {
			pagesArr = append(pagesArr, e.Request.AbsoluteURL(link))
			log.Println("ADD PAGE", e.Request.AbsoluteURL(link))
		}

	})

	// Need to login
	c.OnHTML("a[href='/account/login']", func(e *colly.HTMLElement) {
		if login == false {
			log.Printf("NEED LOGIN\n")

			// authenticate
			err := c.Post(loginPage, map[string]string{"customer[email]": settings.Login, "customer[password]": settings.Password, "form_type": "customer_login", "utf8": "true"})
			if err != nil {
				log.Fatal(err)
			}
			login = true
		}
	})

	c.OnHTML("html", func(e *colly.HTMLElement) {
		if isProductPage {
			var product Product
			product.name = regNoWhiteSpaces.ReplaceAllString(e.ChildText("span[itemprop='name']"), "")
			product.msrp = regNumbers.ReplaceAllString(e.ChildText("span.compare"), "")
			var option Option
			e.ForEach("#inventory>div.price-P4", func(_ int, el *colly.HTMLElement) {
				option.name = regNoWhiteSpaces.ReplaceAllString(el.ChildText("span.inventory-title"), "")
				option.price = regNumbers.ReplaceAllString(el.ChildText("span.inventory-price"), "")
				option.stock = regNumbers.ReplaceAllString(el.ChildText("span.inventory-qty"), "")
				product.options = append(product.options, option)
			})

			myLogger(logFile, "IS PRODUCT", debug)
			myLogger(logFile, fmt.Sprintln(product), debug)
			fmt.Printf(" Count Options: %d\n", cap(product.options))
			for _, v := range product.options {
				//line := []string{product.name, product.msrp, product.description, option.name, option.price, option.stock}
				line := []string{product.name, v.name, product.msrp, v.price, v.stock, currPage}
				err := csvWriter.Write(line)
				checkError("Cannot write to file", err)
			}
		}

	})

	// attach callbacks after login
	c.OnResponse(func(r *colly.Response) {
		currPage = r.Request.URL.String()
		isProductPage = false
		if strings.Contains(r.Request.URL.String(), "account") {
			log.Println("LOGGIN", r.Request.URL, r.StatusCode)
		}
		if strings.Contains(r.Request.URL.String(), "collections") && strings.Contains(r.Request.URL.String(), "products") {
			isProductPage = true
		}
	})

	c.OnError(func(r *colly.Response, e error) {
		fmt.Println("error:", e, r.Request.URL, string(r.Body))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		myLogger(logFile, fmt.Sprintf("Visiting: %s\n", r.URL.String()), debug)
	})

	// Start scraping
	parse := true
	for ok := true; ok; ok = parse {
		parse = false
		for _, v := range pagesArr {
			inArray, _ := in_array(v, parsed)
			if !inArray {
				parse = true
				parsed = append(parsed, v)
				c.Visit(v)
			}

		}
		if parse == false {
			myLogger(logFile, fmt.Sprintf("Total pages parced: %d \n", len(parsed)), debug)
			for _, v := range parsed {
				myLogger(logFile, fmt.Sprintf("PAGE: %s \n", v), debug)
			}
		}
	}

	// Log all product pages
	for i, v := range productsURL {
		myLogger(logFile, fmt.Sprintf("PRODUCT: %d - %s\n", i, v), debug)
		c.Visit(v)
	}

	// test scrap product
	//c.Visit(productsURL[0])
	fulltimeEnd := time.Now()
	myLogger(logFile, fmt.Sprintf("\n---\nElapsed time: %v\n", fulltimeEnd.Sub(fulltimeStart)), debug)
}

func in_array(val string, array []string) (exists bool, index int) {
	exists = false
	index = -1

	for i, v := range array {
		if val == v {
			index = i
			exists = true
			return
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
