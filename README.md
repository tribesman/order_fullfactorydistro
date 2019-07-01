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
```

# Build
```
GOOS=darwin GOARCH=amd64  go build -o build/fullfactorydistro_mac
GOOS=windows GOARCH=amd64  go build -o build/fullfactorydistro_win.exe
```
