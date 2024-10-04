package db

import (
	"antiplagiat/ant"
	"antiplagiat/types"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-sql-driver/mysql"
	"log"
	"os"
	"strconv"
)

var (
	DBUser     = os.Getenv("DBUser")
	DBPassword = os.Getenv("DBPassword")
	DBAddress  = os.Getenv("DBAddress")
	DBName     = os.Getenv("DBName")

	COMPARING_FILE_PATH = os.Getenv("COMPARING_FILE_PATH")
	TREES_PATH          = os.Getenv("TREES_PATH")
)

type DB struct {
	*sql.DB
}

var DBconn DB

func Connect() (DB, error) {
	if DBconn != (DB{}) {
		if err := DBconn.Ping(); err == nil {
			return DBconn, nil
		}
	}

	cfg := mysql.Config{
		User:   DBUser,
		Passwd: DBPassword,
		Net:    "tcp",
		Addr:   DBAddress,
		DBName: DBName,
	}
	conn, err := sql.Open("mysql", cfg.FormatDSN()+"&parseTime=true")
	if err != nil || conn == nil {
		return DB{}, err
	}

	// Checking if connection is valid (Open may not verify the connection)
	if err := conn.Ping(); err != nil {
		return DB{}, err
	}

	log.Println("Connected to database!")
	return DB{conn}, nil
}

func GetCourseIdByName(name string) (int, error) {
	var err error
	if DBconn, err = Connect(); err != nil {
		return -1, err
	}

	row := DBconn.QueryRow(`
		SELECT id
		FROM courses 
		WHERE name = ?
		LIMIT 1`, name)

	var id int
	if err := row.Scan(&id); err != nil {
		return -1, errors.New("No such course")
	}
	return id, nil
}

func GetCourseLongNameByName(name string) (string, error) {
	var err error
	if DBconn, err = Connect(); err != nil {
		return "", err
	}

	row := DBconn.QueryRow(`
		SELECT long_name
		FROM courses 
		WHERE name = ?
		LIMIT 1`, name)

	var long_name string
	if err := row.Scan(&long_name); err != nil {
		return "", errors.New("No such course")
	}
	return long_name, nil
}

func GetUserIdByNameSurname(name, surname string) (int, error) {
	var err error
	if DBconn, err = Connect(); err != nil {
		return -1, err
	}

	row := DBconn.QueryRow(`
		SELECT id
		FROM users 
		WHERE name = ? AND surname = ?
		LIMIT 1`, name, surname)

	var id int
	if err := row.Scan(&id); err != nil {
		return -1, errors.New("No such user")
	}
	return id, nil
}

func GetCourseById(id int) (types.Course, error) {
	var err error
	if DBconn, err = Connect(); err != nil {
		return types.Course{}, err
	}

	row := DBconn.QueryRow(`
		SELECT id, name, long_name
		FROM courses
		WHERE id = ?
		LIMIT 1`, id)

	var course types.Course
	if err := row.Scan(&course.Id, &course.Name, &course.Long_name); err != nil {
		return types.Course{}, errors.New("No such course")
	}
	return course, nil
}

func GetUserById(id int) (types.User, error) {
	var err error
	if DBconn, err = Connect(); err != nil {
		return types.User{}, err
	}

	row := DBconn.QueryRow(`
		SELECT id, name, surname
		FROM users
		WHERE id = ?
		LIMIT 1`, id)

	var user types.User
	if err := row.Scan(&user.Id, &user.Name, &user.Surname); err != nil {
		return types.User{}, errors.New("No such user")
	}
	return user, nil
}

func GetDeliveryById(id int) (types.Delivery, error) {
	var err error
	if DBconn, err = Connect(); err != nil {
		return types.Delivery{}, err
	}

	row := DBconn.QueryRow(`
		SELECT id, user_id, course_id, lab, variant, date, solution, error, language, ant_mark, ant_review
		FROM deliveries
		WHERE id = ?
		LIMIT 1`, id)

	var delivery types.Delivery
	if err := row.Scan(&delivery.Id, &delivery.User_id, &delivery.Course_id, &delivery.Lab, &delivery.Variant, &delivery.Date, &delivery.Solution, &delivery.Error, &delivery.Language, &delivery.Ant_mark, &delivery.Ant_review); err != nil {
		return types.Delivery{}, errors.New("No such delivery")
	}
	return delivery, nil
}

func GetCourseRows() (*sql.Rows, error) {
	var err error
	if DBconn, err = Connect(); err != nil {
		return nil, err
	}

	rows, err := DBconn.Query(`
		SELECT name, long_name, IFNULL(lab_amount, 0) AS lab_amount
		FROM courses
		LEFT JOIN (SELECT course_id, COUNT(DISTINCT lab) AS lab_amount
		FROM deliveries
		GROUP BY course_id) AS amounts ON courses.id = course_id`)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func GetLabRows(course_name string) (*sql.Rows, error) {
	var err error
	if DBconn, err = Connect(); err != nil {
		return nil, err
	}

	course_id, err := GetCourseIdByName(course_name)
	if err != nil {
		return nil, err
	}

	rows, err := DBconn.Query(`
		SELECT lab, IFNULL(COUNT(DISTINCT user_id), 0) AS student_amount
		FROM deliveries
		WHERE course_id = ?
		GROUP BY lab`, course_id)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// sort_by = "surname"/"date" - сортировка полей в таблице
func GetDeliveryRows(course_name string, lab int, sort_by string) (*sql.Rows, error) {
	var err error
	if DBconn, err = Connect(); err != nil {
		return nil, err
	}

	course_id, err := GetCourseIdByName(course_name)
	if err != nil {
		return nil, err
	}

	rows, err := DBconn.Query(`
		SELECT deliveries.id, surname, name, variant, date, language, ant_mark, ant_review
		FROM deliveries
		LEFT JOIN users ON user_id = users.id
		WHERE deliveries.course_id = ? AND lab = ?
		ORDER BY `+sort_by, course_id, lab)

	if err != nil {
		return nil, err
	}
	return rows, nil
}

func AddDelivery(surname, name, course_name string, lab, variant int, solution, language string) error {
	var err error
	if DBconn, err = Connect(); err != nil {
		return err
	}

	user_id, err := GetUserIdByNameSurname(name, surname)
	if err != nil {
		if user_id, err = AddUser(name, surname); err != nil {
			return err
		}
	}

	course_id, err := GetCourseIdByName(course_name)
	if err != nil {
		return err
	}

	result, err := DBconn.Exec(`
		INSERT INTO deliveries (user_id, course_id, lab, variant, solution, language) values (?, ?, ?, ?, ?, ?)`,
		user_id, course_id, lab, variant, solution, language)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	err = ant.CreateTree("create.log", solution, language, TREES_PATH, strconv.Itoa(int(id))+".tree")
	if err != nil {
		UpdateError(int(id), err.Error())
		return err
	}

	return nil
}

func UpdateError(id int, err_str string) error {
	var err error
	if DBconn, err = Connect(); err != nil {
		return err
	}

	_, err = DBconn.Exec(`
		UPDATE deliveries
		SET error = ?
		WHERE id = ?`, err_str, id)
	if err != nil {
		return err
	}

	return nil
}

func AddUser(name, surname string) (int, error) {
	var err error
	if DBconn, err = Connect(); err != nil {
		return 0, err
	}

	result, err := DBconn.Exec(`
		INSERT INTO users (name, surname) values (?, ?)`, name, surname)
	if err != nil {
		return 0, err
	}

	var id int64
	if id, err = result.LastInsertId(); err != nil {
		return 0, err
	}

	return int(id), nil
}

func UpdateAntMarks() error {
	var err error
	if DBconn, err = Connect(); err != nil {
		return err
	}

	// Getting info about new deliveries
	rows, err := DBconn.Query(`
		SELECT id, course_id, lab, variant, solution, error, language
		FROM deliveries 
		WHERE ant_mark = -1
		ORDER BY id`)

	if err != nil {
		return err
	}

	// log.Println("Select (ant_mark = -1):")

	unchecked_deliveries := make([]types.Delivery, 0)
	for rows.Next() {
		var row types.Delivery
		if err := rows.Scan(&row.Id, &row.Course_id, &row.Lab, &row.Variant, &row.Solution, &row.Error, &row.Language); err != nil {
			log.Println(err)
		}
		unchecked_deliveries = append(unchecked_deliveries, row)
		// log.Printf("id: %d, course_id: %d, lab: %d, variant: %d", row.Id, row.Course_id, row.Lab, row.Variant)
	}
	rows.Close()

	// Going through unchecked deliveries
	for i := range unchecked_deliveries {
		unchecked_id_str := strconv.Itoa(unchecked_deliveries[i].Id)

		// Checking if tree is made
		_, err := os.Open(TREES_PATH + unchecked_id_str + ".tree")
		if err != nil && unchecked_deliveries[i].Error.String == "" {
			log.Println(err)
			err = ant.CreateTree("create_log", unchecked_deliveries[i].Solution.String, unchecked_deliveries[i].Language, TREES_PATH, unchecked_id_str+".tree")
			if err != nil {
				log.Println(err)
				err = UpdateError(unchecked_deliveries[i].Id, err.Error())
				if err != nil {
					log.Println(err)
				}
			}
		}

		// Getting info about checked (mark != -1) deliveries with same course, lab and variant
		rows, err := DBconn.Query(`
			SELECT id, user_id, solution, error, language
			FROM deliveries 
			WHERE ant_mark != -1 && ant_mark != -100
			AND course_id = ? AND lab = ? AND variant = ?`,
			unchecked_deliveries[i].Course_id, unchecked_deliveries[i].Lab, unchecked_deliveries[i].Variant)

		if err != nil {
			return err
		}

		// log.Printf("Select (ant_mark != -1 AND course_id = %d AND lab = %d AND variant = %d):", unchecked_deliveries[i].Course_id, unchecked_deliveries[i].Lab, unchecked_deliveries[i].Variant)

		checked_deliveries := make([]types.Delivery, 0)
		for rows.Next() {
			var row types.Delivery
			if err := rows.Scan(&row.Id, &row.User_id, &row.Solution, &row.Error, &row.Language); err != nil {
				log.Println(err)
			}
			// log.Printf("id: %d, user_id: %d", row.Id, row.User_id)
			checked_deliveries = append(checked_deliveries, row)
		}
		rows.Close()

		// And comparing them
		res_map := make(map[int]int) // map: id->ant_mark

		for j := range checked_deliveries {
			checked_id_str := strconv.Itoa(checked_deliveries[j].Id)

			// Checking if tree is made
			_, err = os.Open(TREES_PATH + checked_id_str + ".tree")
			if err != nil && checked_deliveries[j].Error.String == "" {
				err = ant.CreateTree("create.log", checked_deliveries[j].Solution.String, checked_deliveries[j].Language, TREES_PATH, checked_id_str+".tree")
				if err != nil {
					log.Println(err)
					err = UpdateError(checked_deliveries[j].Id, err.Error())
					if err != nil {
						log.Fatal(err)
					}
				}
			}

			res, err := ant.CompareTrees("compare.log", COMPARING_FILE_PATH, TREES_PATH+unchecked_id_str+".tree", TREES_PATH+checked_id_str+".tree")
			if err != nil {
				log.Printf("ERROR comparing %s and %s", unchecked_id_str, checked_id_str)
				res = -1
			}
			res_map[checked_deliveries[j].Id] = int(res * 100)
		}

		// log.Println("Map (id->ant_mark):", res_map)

		var min_ant_mark int
		if len(res_map) == 0 {
			min_ant_mark = 100
		} else {
			min_ant_mark = findMinValue(res_map)
		}

		json_map, err := json.Marshal(res_map)
		if err != nil {
			return err
		}

		_, err = DBconn.Exec(`
			UPDATE deliveries 
			SET ant_mark = ?, ant_review = ?
			WHERE id = ?`,
			min_ant_mark, json_map, unchecked_deliveries[i].Id)

		if err != nil {
			return err
		}
		// log.Printf("Update: ant_mark = %d, ant_review = %v, where id = %d", min_ant_mark, res_map, unchecked_deliveries[i].Id)
	}

	return nil
}

func findMinValue(m map[int]int) int {
	minValue := int(^uint(0) >> 1) // initialize with the highest possible int value
	for _, value := range m {
		if value < minValue {
			minValue = value
		}
	}
	return minValue
}
