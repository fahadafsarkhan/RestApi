package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type myimages struct {
	Hidpi  string `json:"hidpi"`
	Normal string `json:"normal"`
	Teaser string `json:"teaser"`
}

type imgDataItem struct {
	ID          int64     `json:"id"`
	Images      *myimages `json:"images"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}

func saveImages(w http.ResponseWriter, r *http.Request) {
	var data []imgDataItem
	data = getcontent()
	res := getImageAndSaveInImagesFolder(data)
	json.NewEncoder(w).Encode(res)
}

func getImages(w http.ResponseWriter, r *http.Request) {
	res := getdatafromSQLite()
	json.NewEncoder(w).Encode(res)
}

func getImagesbytitle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	res := getdatafromSQLiteusingtitle(params["id"])
	json.NewEncoder(w).Encode(res)
}

// func recordInSQLite()
// {

// }

func getcontent() []imgDataItem {
	// json data
	url := "https://api.dribbble.com/v2/user/shots?access_token=b737dfdc876c78747ff5cda57e4f5cce3bc144327a8f7d2e2e8fd516223c3e9d"

	res, err := http.Get(url)
	checkErr(err)
	//defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	checkErr(err)
	var data []imgDataItem
	err1 := json.Unmarshal(body, &data)
	checkErr(err1)

	return data
}

func getImageAndSaveInImagesFolder(data []imgDataItem) string {
	res := ""
	for i := 0; i < len(data); i++ {
		url, title := data[i].Images.Normal, data[i].Title
		response, e := http.Get(url)
		checkErr(e)
		defer response.Body.Close()
		s := "./tmp/" + title + ".jpg"
		file, err := os.Create(s)
		checkErr(err)
		// Use io.Copy to just dump the response body to the file. This supports huge files
		_, err = io.Copy(file, response.Body)
		checkErr(err)
		file.Close()
		saveinSQLite(data[i].Title, data[i].Description, s, data[i].Images.Normal)
		fmt.Println("image title (" + title + ") saved on location" + s)
		res += " Image title (" + title + ") saved on location " + s
	}
	return res
}

func saveinSQLite(tit, des, filename, origlink string) {
	db, err := sql.Open("sqlite3", "./sqlite.db")
	checkErr(err)
	stmt, err := db.Prepare("INSERT INTO imgdtls(imtitle, imdescription, imorigionallink, imfilename) values(?,?,?,?)")
	checkErr(err)
	res, err := stmt.Exec(tit, des, origlink, filename)
	id, err := res.LastInsertId()
	checkErr(err)
	fmt.Println(id)
	db.Close()
}

func getdatafromSQLite() string {
	db, err := sql.Open("sqlite3", "./sqlite.db")
	checkErr(err)
	rows, err := db.Query("SELECT * FROM imgdtls")
	checkErr(err)
	var imid int
	var imtitle string
	var imdescription string
	var imorigionallink string
	var imfilename string
	s := ""
	for rows.Next() {
		err = rows.Scan(&imid, &imtitle, &imdescription, &imorigionallink, &imfilename)
		checkErr(err)
		s = s + "{ imtitle= " + imtitle + ", imdescription= " + imdescription + ", imorigionallink= " + imorigionallink + ", imfilename= " + imfilename + "} ,         "
	}
	rows.Close()
	db.Close()
	return s
}

func getdatafromSQLiteusingtitle(title string) string {
	db, err := sql.Open("sqlite3", "./sqlite.db")
	checkErr(err)
	rows, err := db.Query("SELECT * FROM imgdtls WHERE imtitle = '" + title + "'")
	checkErr(err)
	var imid int
	var imtitle string
	var imdescription string
	var imorigionallink string
	var imfilename string
	s := ""
	for rows.Next() {
		err = rows.Scan(&imid, &imtitle, &imdescription, &imorigionallink, &imfilename)
		checkErr(err)
		s = s + "{ imtitle= " + imtitle + ", imdescription= " + imdescription + ", imorigionallink= " + imorigionallink + ", imfilename= " + imfilename + "} ,         "
	}
	rows.Close()
	db.Close()
	return s
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/saveimagesincomputer", saveImages).Methods("GET")
	router.HandleFunc("/getdatafromsqlite", getImages).Methods("GET")
	router.HandleFunc("/getdatafromsqlite/{id}", getImagesbytitle).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}
