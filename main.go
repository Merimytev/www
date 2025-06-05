package main

import ("fmt"; "net/http"; "html/template"; "log")

type User struct {
	Name string
	Age uint16
	Money int32
	Hobbies [] string
}

func (u *User) getAllInfo() string {
	return fmt.Sprintf("User name is: %s. He is %d years old, and he has money " +
	 "equal: %d", u.Name, u.Age, u.Money)
}

func (u *User) setNewName(newName string) {
	u.Name = newName
}

func home_page(w http.ResponseWriter, r *http.Request) {
	bob := User{Name: "Bob", Age: 28, Money: -50, Hobbies: []string{"Bicycle", "Basketball", "Running"}}
	// bob.setNewName("Alex")
	// fmt.Fprintf(w, bob.getAllInfo())
	template, err := template.ParseFiles(
		"templates/index.html", 
		"templates/header.html", 
		"templates/footer.html")
	if err != nil {
		log.Println("Ошибка загрузки шаблона:", err)
		http.Error(w, "Ошибка парсинга html шаблона: ", http.StatusInternalServerError)
		return
	}
	template.ExecuteTemplate(w, "index", bob)
}

func sign_page(w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles(
		"templates/sign_page.html", 
		"templates/header.html", 
		"templates/footer.html")
	if err != nil {
		log.Println("Ошибка загрузки шаблона:", err)
		http.Error(w, "Ошибка парсинга html шаблона: ", http.StatusInternalServerError)
		return
	}
	template.ExecuteTemplate(w, "signin", nil)
}

func product_page(w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles(
		"templates/product_page.html", 
		"templates/header.html", 
		"templates/footer.html")
	if err != nil {
		log.Println("Ошибка загрузки шаблона:", err)
		http.Error(w, "Ошибка парсинга html шаблона: ", http.StatusInternalServerError)
		return
	}
	template.ExecuteTemplate(w, "product", nil)
}

func handleRequest() {
	http.HandleFunc("/", home_page)
	http.HandleFunc("/product/", product_page)
	http.HandleFunc("/signin/", sign_page)
	err := http.ListenAndServe(":8088", nil)
	if err != nil {
		log.Fatal("Сервер не запустился:", err)
	}
}

func main() {
	fmt.Println("Сервер запущен...")
	handleRequest()
}