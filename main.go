package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"golang.org/x/crypto/bcrypt"
    _ "github.com/mattn/go-sqlite3" // драйвер для SQLite, _ — чтобы только инициализировать
)

type User struct {
	ID       int
	Username string
	Password string
}

type Product struct {
	ID          int
	Name        string
	Description string
	Price       float64
	Image       string
}

type TemplateData struct {
    Username string
    Products []Product
    Error    string
    Success  string
}

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./db/store.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL,
		password TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		description TEXT,
		price REAL,
		image TEXT
	);
	`)
	if err != nil {
		log.Fatal("Ошибка при создании таблиц:", err)
	}

	// Добаление базовых товаров для примера вывода
		_, _ = db.Exec(`
		    INSERT INTO products (name, description, price, image)
		    SELECT ?, ?, ?, ?
		    WHERE NOT EXISTS (SELECT 1 FROM products WHERE name = ?)`,
		    "Смартфон iPhone 15",
		    "Новый флагман от Apple с процессором A17 и потрясающей камерой",
		    499990,
		    "iphone.jpg",
		    "Смартфон iPhone 15")

		_, _ = db.Exec(`
		    INSERT INTO products (name, description, price, image)
		    SELECT ?, ?, ?, ?
		    WHERE NOT EXISTS (SELECT 1 FROM products WHERE name = ?)`,
		    "Смартфон POCO X5 Pro",
		    "Производительный смартфон от Xiaomi с AMOLED экраном и Snapdragon",
		    149990,
		    "poco.jpg",
		    "Смартфон POCO X5 Pro")

		_, _ = db.Exec(`
			INSERT INTO products (name, description, price, image)
			SELECT ?, ?, ?, ?
			WHERE NOT EXISTS (SELECT 1 FROM products WHERE name = ?)`,
			"Ноутбук Huawei MateBook",
			"Процессор Intel i5-12450H и 16 ГБ ОЗУ для работы и игр. 512 ГБ SSD - быстрая загрузка и хранение файлов. Яркий 16.1 дюймов экран с тонкими рамками. Лёгкий металлический корпус с Windows 11. Идеален для работы, учёбы и мультимедиа.",
			369990,
			"HuaweiLaptop.jpg",
			"Ноутбук Huawei MateBook")

		_, _ = db.Exec(`
			INSERT INTO products (name, description, price, image)
			SELECT ?, ?, ?, ?
			WHERE NOT EXISTS (SELECT 1 FROM products WHERE name = ?)`,
			"Клавиатура Bloody S98 Navy",
			"Механические свитчи Bloody BLMS. Специальная настройка точки и силы срабатывания обеспечивает впечатляющие скорость и ощущения во время игры! Быстрая замена свитчей. Тихий набор текста. Прослойка из силикона уменьшает эффект эха при наборе текста.",
			27990,
			"BloodyKeyboard.jpg",
			"Клавиатура Bloody S98 Navy")
}

func homePage(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, name, description, price, image FROM products")
    if err != nil {
        http.Error(w, "Ошибка запроса продуктов", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var products []Product
    for rows.Next() {
        var p Product
        rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Image)
        products = append(products, p)
    }

    tmpl, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")
    if err != nil {
        http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
        return
    }

    username := ""
    if c, err := r.Cookie("username"); err == nil {
        username = c.Value
    }

    data := TemplateData{
        Username: username,
        Products: products,
    }

    tmpl.ExecuteTemplate(w, "index", data)
}


func registerPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		_, err := db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", name, email, string(hashedPassword))
		if err != nil {
			http.Error(w, "Ошибка при регистрации", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/register?success=1", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("templates/register.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		return
	}

	success := r.URL.Query().Get("success")
	tmpl.ExecuteTemplate(w, "register", struct{ Success string }{Success: success})
}


func signPage(w http.ResponseWriter, r *http.Request) {
    var errorMessage string
    username := ""

    if r.Method == http.MethodPost {
        username = r.FormValue("username")
        password := r.FormValue("password")

        var dbPassword string
        err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&dbPassword)
        if err != nil {
            errorMessage = "Пользователь не найден"
        } else {
            err = bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(password))
            if err != nil {
                errorMessage = "Неверный пароль"
            } else {
                http.SetCookie(w, &http.Cookie{
                    Name:     "username",
                    Value:    username,
                    Path:     "/",
                    HttpOnly: true,
                    MaxAge:   3600 * 24,
                })
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
            }
        }
    } else {
        // Проверка cookie, если пользователь уже авторизован
        if c, err := r.Cookie("username"); err == nil {
            username = c.Value
        }
    }

    tmpl, err := template.ParseFiles("templates/signin.html", "templates/header.html", "templates/footer.html")
    if err != nil {
        http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
        return
    }

    data := TemplateData{
        Username: username,
        Error:    errorMessage,
    }
    tmpl.ExecuteTemplate(w, "signin", data)
}


func logoutPage(w http.ResponseWriter, r *http.Request) {
    http.SetCookie(w, &http.Cookie{
        Name:   "username",
        Value:  "",
        Path:   "/",
        MaxAge: -1, // удалить куки
    })
    http.Redirect(w, r, "/", http.StatusSeeOther)
}


func productPage(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/product/"):]
	var p Product
	err := db.QueryRow("SELECT id, name, description, price, image FROM products WHERE id = ?", idStr).
		Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Image)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles("templates/product.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
		return
	}

	username := ""
	if c, err := r.Cookie("username"); err == nil {
		username = c.Value
	}

	data := TemplateData{
		Username: username,
		Products: []Product{p},
	}
	tmpl.ExecuteTemplate(w, "product", data)
}


func handleRequests() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", homePage)
	http.HandleFunc("/register/", registerPage)
	http.HandleFunc("/signin/", signPage)
	http.HandleFunc("/logout/", logoutPage)
	http.HandleFunc("/product/", productPage)
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir("media")))) // Media (Картинки и т.д.)

	fmt.Println("Сервер запущен")
	fmt.Println("Сервер слушает порт :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	initDB()
	handleRequests()
}
