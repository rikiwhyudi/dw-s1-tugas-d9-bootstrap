package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
	"tugas9/connection"

	"github.com/gorilla/mux"
)

var Data = map[string]interface{} {
	"Title": "Personal Web",
}

type Blog struct {
	Id 			int
	Title 		string
	Images		string
	Start_date 	time.Time
	End_date	time.Time
	Duration	string
	// Post_date 	time.Time
	Format_date	string
	Author		string
	Content 	string
}

var Blogs = []Blog{
	
}

func main() {

	route := mux.NewRouter()

	connection.DatabaseConnect()

	// route path folder untuk public
	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer((http.Dir("./public")))))

	//routing
	route.HandleFunc("/", home).Methods("GET")
	route.HandleFunc("/contact", contact).Methods("GET")
	route.HandleFunc("/blog-detail/{index}", blogDetail).Methods("GET")
	route.HandleFunc("/blog", form).Methods("GET")
	route.HandleFunc("/process", process).Methods("POST")
	route.HandleFunc("/delete/{index}", deleted).Methods("GET")


	fmt.Println("Server running on port 5000");
	http.ListenAndServe("localhost:5000", route)
}

func home(w http.ResponseWriter, r *http.Request) {
	//mengatur header
	w.Header().Set("Content-Type", "text/html; charset=utf-8")


	//membuat variabel memparsing template halaman index
	var tmpl, err  = template.ParseFiles("views/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	rows, _ := connection.Conn.Query(context.Background(), "SELECT id, name, start_date, end_date, description, image FROM tb_projects")

	var result []Blog 

	for rows.Next() {
		var each = Blog{} //memanggil struct
		err := rows.Scan(&each.Id, &each.Title, &each.Start_date, &each.End_date, &each.Content, &each.Images)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		each.Author ="Riki Wahyudi"
		each.Duration ="3 Month"
		// each.Format_date = each.Start_date.Format("2 January 2006")
		result = append(result, each)
	}

	respData := map[string]interface{}{
		"Blogs": result,
	}	

	w.WriteHeader(http.StatusOK) 
	tmpl.Execute(w, respData)
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

func blogDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/blog-detail.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" +err.Error()))
		return
	}
	var BlogDetail = Blog{}

	index, _ := strconv.Atoi(mux.Vars(r)["index"]) 
	for i, data := range Blogs {
		if index == i {
			BlogDetail = Blog{
				Title: data.Title,
				// Post_date: data.Post_date,
				Author: data.Author,
				Content: data.Content,
			}
		}
	}

	data := map[string]interface{}{
		"Blog": BlogDetail,
	}
	

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, data)
}


func form(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/blog.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message :" +err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, nil)

}

func process(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Title :" + r.PostForm.Get("inputTitle"))
	fmt.Println("Content :" + r.PostForm.Get("inputContent"))
	fmt.Println("StartDate :" + r.PostForm.Get("inputStart"))
	fmt.Println("EndDate :" + r.PostForm.Get("inputEnd"))

	var title = r.PostForm.Get("inputTitle")
	var content = r.PostForm.Get("inputContent")

	var newBlog = Blog{
		Title: title,
		Content: content,
		Author: "Riki Wahyudi",
	}

	Blogs = append(Blogs, newBlog)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func deleted(w http.ResponseWriter, r *http.Request) {
	index, _ := strconv.Atoi(mux.Vars(r)["index"])
	Blogs = append(Blogs[:index], Blogs[index+1:]...)
	fmt.Println(Blogs)
	http.Redirect(w, r, "/", http.StatusFound)
}