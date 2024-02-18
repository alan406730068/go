package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
)

func setCookie(w http.ResponseWriter, r *http.Request) {
	c := http.Cookie{
		Name:     "username",
		Value:    "Alan",
		HttpOnly: true,
	}
	http.SetCookie(w, &c)
}

func getCookie(w http.ResponseWriter, r *http.Request) {

	c, err := r.Cookie("username")
	if err != nil {
		fmt.Fprintln(w, "Cannot get cookie")
	}
	fmt.Fprintln(w, c)
}
func out(w http.ResponseWriter, r *http.Request) {
	// t1, err := template.ParseFiles("view/register.html") //讀取檔案
	// if err != nil {
	// 	panic(err)
	// }
	// t1.Execute(w, nil)
	fmt.Println("aaa")
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["auth"] = false
	_ = session.Save(r, w)
	http.Redirect(w, r, "/success", 301)
}
func login(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if session.IsNew {
		t1, err := template.ParseFiles("view/login.html") //讀取檔案
		if err != nil {
			panic(err)
		}
		add := " "
		c, err := r.Cookie("adress")
		if err == nil {
			add = c.Value
		}
		t1.Execute(w, struct {
			Add string
		}{
			add,
		})
	} else {
		http.Redirect(w, r, "/success", 301)
	}
}
func register(w http.ResponseWriter, r *http.Request) {
	t1, err := template.ParseFiles("view/register.html") //讀取檔案
	if err != nil {
		panic(err)
	}
	t1.Execute(w, nil)
}
func registerAccount(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //得到login的資料，若無回傳資料則這個網頁無法正常運作  會獲得一個dic
	var adress []string
	var password []string
	var name []string
	for k, v := range r.Form {
		if k == "adress" {
			adress = v
		} else if k == "password" {
			password = v
		} else if k == "name" {
			name = v
		}
	}

	rows, _ := db.Query("SELECT `adress` FROM `login` WHERE `adress` = ?", adress[0])
	if !rows.Next() { //bool值  裡面沒東西為false
		db.Exec("INSERT INTO `login`(`name`, `adress`,`password`) VALUES (? , ? , ? )", name[0], adress[0], password[0]) //執行sql語法
		http.Redirect(w, r, "/login", 301)
	} else {
		http.Redirect(w, r, "/register", 301)
	}
	defer rows.Close()
}

func processlogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //得到login的資料，若無回傳資料則這個網頁無法正常運作  會獲得一個dic
	var adress []string
	var password []string
	for k, v := range r.Form {
		if k == "adress" {
			adress = v
		} else if k == "password" {
			password = v
		}
	}
	rows, err := db.Query("SELECT `name`,`adress`,`password` FROM `login` WHERE `adress` = ?", adress[0])
	var Tname string
	var Tadress string
	var Tpasswd string
	for rows.Next() {
		rows.Scan(&Tname, &Tadress, &Tpasswd)
		if err != nil {
			log.Fatalln(err)
		}
	}
	defer rows.Close()
	if adress[0] == Tadress && password[0] == Tpasswd {
		session, _ := store.Get(r, "session-name")
		session.Values["auth"] = true
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/success", 301)
	} else {
		c := http.Cookie{
			Name:     "adress",
			Value:    Tadress,
			HttpOnly: true,
		}
		http.SetCookie(w, &c)
		http.Redirect(w, r, "/login", 301)
	}
}
func success(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	if session.IsNew {
		session.Options.MaxAge = -1
		_ = session.Save(r, w)
		http.Redirect(w, r, "/login", 301)
	} else {
		auth := session.Values["auth"]
		if auth != nil {
			isAuth, ok := auth.(bool)
			if ok && isAuth {
				t1, err := template.ParseFiles("view/index.html") //讀取檔案
				if err != nil {
					panic(err)
				}
				t1.Execute(w, nil)
			} else {
				session.Options.MaxAge = -1
				_ = session.Save(r, w)
				http.Redirect(w, r, "/login", 301)
				return
			}
		} else {
			session.Options.MaxAge = -1
			_ = session.Save(r, w)
			http.Redirect(w, r, "/login", 301)
			return
		}
	}
}
func fail(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "bad")
}
func main() {
	http.HandleFunc("/login", login)                     //登入介面
	http.HandleFunc("/out", out)                         //登出介面
	http.HandleFunc("/register", register)               //註冊介面
	http.HandleFunc("/registerAccount", registerAccount) //註冊介面
	http.HandleFunc("/processlogin", processlogin)       //處理登入資訊
	http.HandleFunc("/success", success)                 //登入成功
	http.HandleFunc("/fail", fail)                       //登入失敗
	http.HandleFunc("/setCookie", setCookie)             //設定cookie
	http.HandleFunc("/getCookie", getCookie)             //取得cookie
	http.ListenAndServe(":8000", nil)
}

var db *sql.DB

var store *sessions.CookieStore

// 與DB連線。 init() 初始化，時間點比 main() 更早。
func init() {
	store = sessions.NewCookieStore([]byte("secret-key"))
	dbConnect, err := sql.Open(
		"mysql",
		"root:az66886688@tcp(127.0.0.1:3306)/golang",
	)

	if err != nil {
		log.Fatalln(err) //建立資料檔
	}

	err = dbConnect.Ping()
	if err != nil {
		log.Fatalln(err)
	}

	db = dbConnect // 用全域變數接

	db.SetMaxOpenConns(10) // 可設置最大DB連線數，設<=0則無上限（連線分成 in-Use正在執行任務 及 idle執行完成後的閒置 兩種）
	db.SetMaxIdleConns(10) // 設置最大idle閒置連線數。
}
