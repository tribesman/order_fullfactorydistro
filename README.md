# Заказ на fullfactorydistro.com
Скрипт набивает товары в корзину из order.xlsx файла

Для генерации файла с остатками и ценами необходимо использовать https://github.com/tribesman/fullfactorydistro.com

Колонки:
* Name - Наименование (как на сайте fullfactorydistro.com)	
* Option - Модификация товара (как на сайте fullfactorydistro.com)
* MSRP - Цена товара (необходима для проверки товара)
* Total -  кол-во товара к заказу

# Перед началом работы
Не обходимо создать файл ./config.toml в тойже папке в которой лежит программа
```
Login = "login"
Password = "pwd"
LoginUrl    = "https://fullfactorydistro.com/account/login"
HomeUrl     = "https://fullfactorydistro.com"
CartUrl     = "https://fullfactorydistro.com/cart"
Xlsx        = "order.xlsx" # файл с заказом
Log         = "log.txt" # лог
LogCSV      = "output.csv"  # файл с выводом товаров и признаком добавления в корзину
Name        = "Name"  # Наименование товара
mod         = "Option" # Модификация
msrp        = "MSRP"  # РРЦ
url         = "Photo"  # Ссылка на страницу товара
total       = "Total"  # Кол-во товара для добавления в корзину

checkoutNote        = "Is a test automatic order, please cancel the order."
checkoutFirstName   = "___"
checkoutLastName    = "___"
checkoutAddress     = "___"
checkoutCity        = "___"
checkoutCountry     = "___"
checkoutProvince    = "___"
checkoutZip         = "___"
checkoutShippingMethood         = "shopify-Freight%20Forwarder%20(International%20Only)-0.00"
checkoutPaymentGateway          = "___"
checkoutDifferentBillingAddress = "false"
```

# Build
```
GOOS=darwin GOARCH=amd64  go build -o build/fullfactorydistro_mac
GOOS=windows GOARCH=amd64  go build -o build/fullfactorydistro_win.exe
```
