package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"tugas12/connection"
	"tugas12/middleware"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

/*
struct seperti blueprint/cetakan.
sebagai tipe data penampung hasil query
*/
type MetaData struct {
	Title		string
	IsLogin		bool
	UserName	string
	FlashData	string
}

var Data =  MetaData{
	Title: "Personal Web",
}

type Blog struct {
	Id			int
	Title 			string
	Images			string
	Start_date 		time.Time
	End_date		time.Time
	Duration		string
	Post_date 		time.Time
	SFormat_date		string
	EFormat_date		string
	Author			string
	Technologies 		[]string
	Content 		string
	IsLogin			bool
}

type User struct {
	Id		int
	Name 		string
	Email		string
	Password	string
}

func main() {

	route := mux.NewRouter()

	connection.DatabaseConnect()

	// route path folder untuk public
	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer((http.Dir("./public")))))
	route.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads/"))))


	//routing. parameter pertama adalah rute dan parameter ke-2 adalah handlernya dengan method get dan post dll
	route.HandleFunc("/", home).Methods("GET")
	route.HandleFunc("/contact", contact).Methods("GET")

	route.HandleFunc("/blog-detail/{id}", blogDetail).Methods("GET")
	route.HandleFunc("/blog", form).Methods("GET")
	route.HandleFunc("/process", middleware.UploadFile(process)).Methods("POST")

	route.HandleFunc("/edit/{id}", editForm).Methods("GET")
	route.HandleFunc("/updated/{id}", middleware.UploadFile(edit)).Methods("POST")

	route.HandleFunc("/delete/{id}", deleted).Methods("GET")

	route.HandleFunc("/form-register", formRegister).Methods("GET")
	route.HandleFunc("/register", register).Methods("POST")

	route.HandleFunc("/form-login", formLogin).Methods("GET")
	route.HandleFunc("/login", login).Methods("POST")

	route.HandleFunc("/logout", logout).Methods("GET")


	fmt.Println("Server running on port 5000");
	//membuat sekaligus start server baru
	http.ListenAndServe("localhost:5000", route)
}

	/* untuk keperluan penanganan request ke rute yang ditentukan
	 Parameter ke-1 merupakan objek untuk keperluan http response
	 parameter ke-2 yang bertipe pointer dereff *request, berisikan informasi-informasi yang berhubungan dengan http request 
	 untuk rute yang bersangkutan.
	 */
func home(w http.ResponseWriter, r *http.Request) {
	//mengatur header
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	//membuat variabel memparsing template halaman index
	var tmpl, err  = template.ParseFiles("views/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		//meng-output-kan nilai balik data. Argumen method adalah data yang ingin dijadikan output
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}

	fm := session.Flashes("message")

	var flashes []string
	if len(fm) > 0 {
		session.Save(r, w)
		for _, fl := range fm {
			flashes = append(flashes, fl.(string))
		}
	}

	Data.FlashData = strings.Join(flashes, "")

	//mengambil semua data yg di select dari database di tb_project untuk kemudian di render ke halaman depan (index).
	rows, _ := connection.Conn.Query(context.Background(), "SELECT tb_projects.id, title, start_date, end_date, description, image, post_date, technologies, name FROM tb_projects LEFT JOIN tb_user ON tb_projects.authorid = tb_user.id ORDER BY id DESC")

	var result []Blog //data slice of array di gunakan untuk menampung hasil query

	for rows.Next() {
		var each = Blog{} //memanggil struct
		//scan mengambil nilai record yang sedang diiterasi, untuk disimpan pada variabel pointer
		err := rows.Scan(
			&each.Id, 
			&each.Title, 
			&each.Start_date, 
			&each.End_date, 
			&each.Content, 
			&each.Images, 
			&each.Post_date,
			&each.Technologies,
			&each.Author)

		if err != nil {
			fmt.Println(err.Error())
			return
		}
		each.Duration = getDuration(each.Start_date, each.End_date)
		each.SFormat_date = each.Post_date.Format("2 January 2006")

		if session.Values["IsLogin"] != true {
			Data.IsLogin = false
		} else {
			Data.IsLogin = session.Values["IsLogin"].(bool)
		}

		result = append(result, each)
	}

	respData := map[string]interface{}{
		"Data":	Data,
		"Blogs": result,
	}	

	w.WriteHeader(http.StatusOK) 
	tmpl.Execute(w, respData)
}

func blogDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	//konversi string ke int
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	var tmpl, err = template.ParseFiles("views/blog-detail.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message :" +err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")
	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}

	var BlogDetail = Blog{}
	/*mengambil data berdasarkan dari id di database didalam tb_project untuk kemudian di render ke halaman details blog
	*/
	err = connection.Conn.QueryRow(context.Background(), 
	"SELECT id, title, start_date, end_date, description, image, technologies FROM tb_projects WHERE id=$1", id).Scan(
		&BlogDetail.Id, 
		&BlogDetail.Title, 
		&BlogDetail.Start_date, 
		&BlogDetail.End_date, 
		&BlogDetail.Content, 
		&BlogDetail.Images,
		&BlogDetail.Technologies,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	BlogDetail.Duration = getDuration(BlogDetail.Start_date, BlogDetail.End_date)
	BlogDetail.SFormat_date = BlogDetail.Start_date.Format("2 January 2006")
	BlogDetail.EFormat_date = BlogDetail.End_date.Format("2 January 2006")

	data := map[string]interface{}{
		"Data": Data,
		"Blog": BlogDetail,
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, data)
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var  tmpl, err = template.ParseFiles("views/form.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" +err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, nil)
}

func form(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/blog.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" +err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")
	
	fm := session.Flashes("message")

	var flashes []string
	if len(fm) > 0 {
		session.Save(r, w)
		for _, fl := range fm {
			flashes = append(flashes, fl.(string))
		}
	}

		Data.FlashData = strings.Join(flashes, "")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false		
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, Data)

}

func process(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var title = r.PostForm.Get("inputTitle")
	var content = r.PostForm.Get("inputContent")
	var start = r.PostForm.Get("inputStart")
	var end = r.PostForm.Get("inputEnd")
	var nodejs = r.PostForm.Get("nodejs")
	var nextjs = r.PostForm.Get("nextjs")
	var reactjs = r.PostForm.Get("reactjs")
	var typescript = r.PostForm.Get("typescript")
	
	var technologies = []string{
		nodejs,
		nextjs,
		reactjs,
		typescript,
	} 

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")
	author := session.Values["Id"].(int)

	dataContex := r.Context().Value("dataFile")
	image := dataContex.(string)

	_, err = connection.Conn.Exec(context.Background(), 
	"INSERT INTO tb_projects(title, start_date, end_date, description, image, technologies, authorid) VALUES ($1, $2, $3, $4, $5, $6, $7)", 
	title, start, end, content, image, technologies, author)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " +err.Error()))
		return
	}

	session.AddFlash("Project berhasil di tambahkan... ", "message")
	session.Save(r, w)

	http.Redirect(w, r, "/blog", http.StatusMovedPermanently)
}

func editForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/form-edit.html")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " +err.Error()))
		return
	}
	
	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	fm := session.Flashes("message")

	var flashes []string
	if len(fm) > 0 {
		session.Save(r, w)
		for _, fl := range fm {
			flashes = append(flashes, fl.(string))
		}
	}
		Data.FlashData = strings.Join(flashes, "")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Name"].(string)
	}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	var edit = Blog{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT id, title, start_date, end_date, description, image, technologies FROM tb_projects WHERE id=$1", id).Scan(
		&edit.Id, 
		&edit.Title, 
		&edit.Start_date,
		&edit.End_date,
		&edit.Content, 
		&edit.Images, 
		&edit.Technologies,
	)
	
	edit.SFormat_date = edit.Start_date.Format("2006-01-02")
	edit.EFormat_date = edit.End_date.Format("2006-01-02")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " +err.Error()))
		return
	}

	data := map[string]interface{}{
		"Data": Data,
		"Blogs": edit,
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, data)

}

func edit(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Title :" + r.PostForm.Get("inputTitle"))
	fmt.Println("Content :" + r.PostForm.Get("inputContent"))
	fmt.Println("Start :" + r.PostForm.Get("inputStart"))
	fmt.Println("End :" + r.PostForm.Get("inputEnd"))
	
	var title = r.PostForm.Get("inputTitle")
	var content = r.PostForm.Get("inputContent")
	var start = r.PostForm.Get("inputStart")
	var end = r.PostForm.Get("inputEnd")
	var nodejs = r.PostForm.Get("nodejs")
	var nextjs = r.PostForm.Get("nextjs")
	var reactjs = r.PostForm.Get("reactjs")
	var typescript = r.PostForm.Get("typescript")
	
	var technologies = []string{
		nodejs,
		nextjs,
		reactjs,
		typescript,
	} 

	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	dataContex := r.Context().Value("dataFile")
	image := dataContex.(string)

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")
	// author := session.Values["Id"].(int)

	_, err = connection.Conn.Exec(context.Background(), "UPDATE tb_projects SET title=$1, start_date=$2, end_date=$3, description=$4, image=$5, technologies=$6 WHERE id=$7" , title, start, end, content, image, technologies, id)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " +err.Error()))
		return
	}

	session.AddFlash("Project berhasil di edit... ", "message")
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
	
func deleted(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_projects WHERE id=$1", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " +err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}


func getDuration(Start_date time.Time, End_date time.Time) string {
	
	timeFormat := "2006-01-02"
	start, _ := time.Parse(timeFormat, Start_date.Format(timeFormat))
	end, _ := time.Parse(timeFormat, End_date.Format(timeFormat))

	distance := end.Sub(start).Hours() / 24
	var duration string

	if distance > 30 {
		if (distance / 30) <= 1 {
			duration = "1 Month"
		}
	duration = strconv.Itoa(int(distance)/30) + " Month"
	} else {
		if distance <= 1 {
			duration = "1 Days"
		} 
	duration = strconv.Itoa(int(distance)) + " Days"
	}

	return duration
}

func formRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/form-register.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " +err.Error()))
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	fm := session.Flashes("message")

	var flashes []string
	if len(fm) > 0 {
		session.Save(r, w)
		for _, fl := range fm {
			flashes = append(flashes, fl.(string))
		}
	}

	Data.FlashData = strings.Join(flashes, "")

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, Data)
}

func register(w http.ResponseWriter, r *http.Request) {
	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var name = r.PostForm.Get("inputName")
	var email = r.PostForm.Get("inputEmail")
	var password = r.PostForm.Get("inputPassword")

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	fmt.Println(passwordHash)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_user(name, email, password) VALUES($1, $2, $3)", name, email, passwordHash)
	
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " +err.Error()))
		return	
	}

	session.AddFlash("Succesfully register... ", "message")
	session.Save(r, w)

	http.Redirect(w, r, "/form-register", http.StatusMovedPermanently)
}

func formLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/form-login.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	fm := session.Flashes("message")

	var flashes []string
	if len(fm) > 0 {
		session.Save(r, w)
		for _, fl := range fm {
			flashes = append(flashes, fl.(string))
		}
	}

	Data.FlashData = strings.Join(flashes, "")
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, Data)
}

func login(w http.ResponseWriter, r *http.Request) {
	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")

	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	email := r.PostForm.Get("inputEmail")
	password := r.PostForm.Get("inputPassword")

	user := User{}
	var messages string 

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_user WHERE email=$1", email).Scan(
		&user.Id, 
		&user.Name, 
		&user.Email, 
		&user.Password,
	)

	if err != nil {
		messages = "Email not registered!"
		session.AddFlash(messages, "message")
		session.Save(r, w)
		fmt.Println(messages)
		http.Redirect(w, r, "/form-login", http.StatusMovedPermanently)
		 return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		messages = "Wrong password!"		
		session.AddFlash(messages, "message")
		session.Save(r, w)
		fmt.Println(messages)
		http.Redirect(w, r, "/form-login", http.StatusMovedPermanently)
		return 
	}

	session.Values["IsLogin"] = true
	session.Values["Name"] = user.Name
	session.Values["Id"] = user.Id
	session.Options.MaxAge = 10800 // 3 hours
	session.AddFlash("Successfully login!", "message")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func logout(w http.ResponseWriter, r *http.Request) {
	var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
	session, _ := store.Get(r, "SESSION_KEY")
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}