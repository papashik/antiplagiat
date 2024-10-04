package main

import (
	//"io"
	"antiplagiat/db"
	"antiplagiat/types"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	HTTPS_SERVER = os.Getenv("HTTPS_SERVER")
	PORT         = os.Getenv("PORT")
	CERT_PATH    = os.Getenv("CERT_PATH")
	KEY_PATH     = os.Getenv("KEY_PATH")
)

// struct for template.Execute
type RootPage struct {
	Courses       []CourseRow
	Error_message string
}

type CourseRow struct {
	Name       string
	Long_name  string
	Lab_amount int
}

// struct for template.Execute
type CoursePage struct {
	Labs             []LabRow
	Сourse_name      string
	Course_long_name string
	Error_message    string
}

type LabRow struct {
	Number         int
	Student_amount int
}

// struct for template.Execute
type TablePage struct {
	Deliveries       []DeliveryRow
	Сourse_name      string
	Course_long_name string
	Lab              int
	Error_message    string
}

type DeliveryRow struct {
	Id             int
	Surname        string
	Name           string
	Variant        int
	Date           string
	Language       string
	Ant_mark       int
	Ant_mark_color string
	Ant_suspicion  template.HTML
}

// struct for template.Execute
type DeliveryPage struct {
	Delivery       types.Delivery
	Date           string
	Ant_mark_color string
	Ant_suspicion  template.HTML
	Ant_review     template.HTML // full review, type HTML for formatting with colors
	Solution       template.HTML // type HTML for showing line wraps
	Parsing_error  template.HTML

	Course        types.Course
	User          types.User
	Error_message string
}

func GenerateRootPage() (root RootPage) {
	rows, err := db.GetCourseRows()
	if err != nil {
		log.Println(err)
		root.Error_message = err.Error()
		return
	}

	var res_rows []CourseRow
	for rows.Next() {
		var row CourseRow
		if err := rows.Scan(&row.Name, &row.Long_name, &row.Lab_amount); err != nil {
			log.Println(err)
			root.Error_message = err.Error()
			return
		}

		res_rows = append(res_rows, row)
	}
	rows.Close()
	root.Courses = res_rows

	return
}

func GenerateCoursePage(course_name string) (course CoursePage) {
	var err error
	course.Сourse_name = course_name
	course.Course_long_name, err = db.GetCourseLongNameByName(course_name)
	if err != nil {
		log.Println(err)
		course.Error_message = err.Error()
		return
	}

	rows, err := db.GetLabRows(course_name)
	if err != nil {
		log.Println(err)
		course.Error_message = err.Error()
		return
	}

	var res_rows []LabRow
	for rows.Next() {
		var row LabRow
		if err := rows.Scan(&row.Number, &row.Student_amount); err != nil {
			log.Println(err)
			course.Error_message = err.Error()
			return
		}
		res_rows = append(res_rows, row)
	}
	rows.Close()
	course.Labs = res_rows

	return
}

func GenerateTablePage(course_name string, lab int) (table TablePage) {
	var err error
	db.UpdateAntMarks()
	table.Lab = lab
	table.Сourse_name = course_name
	table.Course_long_name, err = db.GetCourseLongNameByName(course_name)
	if err != nil {
		log.Println(err)
		table.Error_message = err.Error()
		return
	}

	rows, err := db.GetDeliveryRows(course_name, lab, "date DESC")
	if err != nil {
		log.Println(err)
		table.Error_message = err.Error()
		return
	}

	var res_rows []DeliveryRow
	for rows.Next() {
		var row DeliveryRow
		var t time.Time                  // for formatting time
		var ant_suspicion sql.NullString // for receiving ant_review
		if err := rows.Scan(&row.Id, &row.Surname, &row.Name, &row.Variant, &t, &row.Language, &row.Ant_mark, &ant_suspicion); err != nil {
			log.Println(err)
			table.Error_message = err.Error()
			return
		}
		row.Date = t.Format("02.01.2006 15:04:05")
		row.Ant_mark_color = GetMarkColor(row.Ant_mark)
		if ant_suspicion.Valid {
			row.Ant_suspicion = template.HTML(ant_suspicion.String)
		}
		row.Ant_suspicion = GetAntSuspicion(string(row.Ant_suspicion))
		res_rows = append(res_rows, row)
	}
	rows.Close()
	table.Deliveries = res_rows

	return
}

func GenerateDeliveryPage(id int) (page DeliveryPage) {
	var err error
	page.Delivery, err = db.GetDeliveryById(id)
	if err != nil {
		log.Println(err)
		page.Error_message = err.Error()
		return
	}

	page.User, err = db.GetUserById(page.Delivery.User_id)
	if err != nil {
		log.Println(err)
		page.Error_message = err.Error()
		return
	}

	page.Course, err = db.GetCourseById(page.Delivery.Course_id)
	if err != nil {
		log.Println(err)
		page.Error_message = err.Error()
		return
	}

	page.Date = page.Delivery.Date.Format("02.01.2006 15:04:05")
	page.Ant_mark_color = GetMarkColor(page.Delivery.Ant_mark)
	page.Ant_suspicion = GetAntSuspicion(page.Delivery.Ant_review.String)
	page.Ant_review = GetAntReview(page.Delivery.Ant_review.String)
	page.Solution = template.HTML(`<pre style="display: inline-block; padding: 0px 20px; border:3px #42aaff solid; border-radius: 20px">` + strings.Replace(strings.Replace(template.HTMLEscapeString(page.Delivery.Solution.String), "\n", "<br>", -1), "\t", "&emsp;&emsp;&emsp;&emsp;", -1) + "</pre>")
	page.Parsing_error = template.HTML(`<pre style="display: inline-block; padding: 0px 20px">` + strings.Replace(strings.Replace(template.HTMLEscapeString(page.Delivery.Error.String), "\n", "<br>", -1), "\t", "&emsp;&emsp;&emsp;&emsp;", -1) + "</pre>")
	return
}

// Mark constants for choosing color and messages
const (
	AWFUL   = 50
	BAD     = 60
	AVERAGE = 70
	NORMAL  = 100
)

func GetMarkColor(ant_mark int) string {
	switch {
	case ant_mark <= AWFUL:
		return "#FF2400" // red
	case ant_mark <= BAD:
		return "#FF4F00" // orange
	case ant_mark <= AVERAGE:
		return "#FFD700" // yellow gold
	case ant_mark <= NORMAL:
		return "#00A550" // green
	default:
		return "#000000"
	}
}

func GetAntSuspicion(ant_review_json string) template.HTML {
	var review map[int]int
	err := json.Unmarshal([]byte(ant_review_json), &review)
	if err != nil {
		return template.HTML(err.Error())
	}

	var min_mark, min_mark_id int = 100, 0
	for id, mark := range review {
		if mark < min_mark || (mark == min_mark && id < min_mark_id) {
			min_mark = mark
			min_mark_id = id
		}
	}

	if min_mark_id == 0 {
		return template.HTML("")
	}

	var res string
	switch {
	case min_mark <= BAD:
		res = fmt.Sprintf(`<a href="/delivery/%d">Посылка № %d</a>`, min_mark_id, min_mark_id)
	default:
		res = fmt.Sprintf(`OK (<a href="/delivery/%d">№ %d</a>)`, min_mark_id, min_mark_id)
	}
	return template.HTML(res)
}

func GetAntReview(ant_review_json string) template.HTML {
	var review map[int]int
	err := json.Unmarshal([]byte(ant_review_json), &review)
	if err != nil {
		return template.HTML(err.Error())
	}

	if len(review) == 0 {
		return ""
	}

	var result string = "<ul>"
	for id, mark := range review {
		result += fmt.Sprintf(`<li><a href="/delivery/%d">Посылка № %d</a>. Результат: <span style="color:%s">%d</span></li>`, id, id, GetMarkColor(mark), mark)
	}
	return template.HTML(result + "</ul>")
}

func mainHandler(w http.ResponseWriter, req *http.Request) {
	path_array := strings.FieldsFunc(req.URL.Path, func(c rune) bool { return c == '/' })
	switch len(path_array) {
	// root page
	case 0:
		root := GenerateRootPage()
		tmpl, err := template.ParseFiles("./html/root.html")
		if err != nil {
			errorHandler(w, req, http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, root)

	// course page
	case 1:
		course_name := path_array[0]
		course := GenerateCoursePage(course_name)
		tmpl, err := template.ParseFiles("./html/course.html")
		if err != nil {
			errorHandler(w, req, http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, course)

	// table page (report of single lab results)
	case 2:
		course_name := path_array[0]
		lab_str := path_array[1]
		lab, err := strconv.Atoi(lab_str)
		if err != nil {
			errorHandler(w, req, http.StatusNotFound)
			return
		}

		table := GenerateTablePage(course_name, lab)
		tmpl, err := template.ParseFiles("./html/table.html")
		if err != nil {
			errorHandler(w, req, http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, table)

	// 404 error
	default:
		errorHandler(w, req, http.StatusNotFound)
	}
}

func deliveryHandler(w http.ResponseWriter, req *http.Request) {
	path_array := strings.FieldsFunc(req.URL.Path, func(c rune) bool { return c == '/' })
	switch len(path_array) {
	// delivery page (report of single delivery)
	case 2:
		delivery_id_str := path_array[1]
		id, err := strconv.Atoi(delivery_id_str)
		if err != nil {
			errorHandler(w, req, http.StatusNotFound)
			return
		}

		page := GenerateDeliveryPage(id)
		tmpl, err := template.ParseFiles("./html/delivery.html")
		if err != nil {
			errorHandler(w, req, http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, page)

	// 404 error
	default:
		errorHandler(w, req, http.StatusNotFound)
	}
}

func errorHandler(w http.ResponseWriter, req *http.Request, status int) {
	w.WriteHeader(status)
	switch status {
	case http.StatusNotFound:
		fmt.Fprint(w, "404 Not found")

	case http.StatusInternalServerError:
		fmt.Fprint(w, "500 HTML parsing error")

	default:
		fmt.Fprint(w, "Error occured")
	}

}

func main() {
	// logging settings
	log_file, err := os.OpenFile("ant_server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	log.SetOutput(log_file)

	http.HandleFunc("/delivery/", deliveryHandler)
	http.Handle("/html/", http.StripPrefix("/html", http.FileServer(http.Dir("./html"))))
	http.HandleFunc("/", mainHandler)

	if HTTPS_SERVER == "true" {
		err = http.ListenAndServeTLS(PORT, CERT_PATH, KEY_PATH, nil)
	} else {
		err = http.ListenAndServe(PORT, nil)
	}
	log.Fatal(err)
}
