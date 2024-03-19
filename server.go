package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Gender string

var myKey = []byte("SqwozBab")

const (
	Male   Gender = "male"
	Female Gender = "female"
)

type Actor struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Sex         Gender `json:"sex"`
	DateOfBirth string `json:"date_of_birth"`
}

type Film struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DateOfIssue string `json:"date_of_issue"`
	Rating      int    `json:"rating"`
	Actor_list  []int  `json:"actor_list"`
}

type ReceiptFilm struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DateOfIssue string `json:"date_of_issue"`
	Rating      int    `json:"rating"`
}

type SearchFilm struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DateOfIssue string `json:"date_of_issue"`
	Rating      int    `json:"rating"`
}

type ReceivingActor struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Sex         Gender `json:"sex"`
	DateOfBirth string `json:"date_of_birth"`
	Film        []SearchFilm
}

type Users struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Mail     string `json:"mail"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func newFilmWithDefaultRating() Film {
	return Film{Rating: -1}
}

func handler() {
	conf := newConf()
	connect := connect(conf)
	mx := http.NewServeMux()
	srv := &http.Server{
		Addr:    conf.Server,
		Handler: mx,
	}

	mx.HandleFunc("/user", user(connect))
	mx.HandleFunc("/actor", actorHandler(connect))
	mx.HandleFunc("/film", filmHandler(connect))
	mx.HandleFunc("/health", healthCheckHandler)
	mx.HandleFunc("/user/login", userLogin(connect))

	log.Printf("Сервер %s работает.", conf.Server)
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func actorHandler(connect *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			deleteActor(connect, w, r)
		case http.MethodPut:
			updateActor(connect, w, r)
		case http.MethodPost:
			createActor(connect, w, r)
		case http.MethodGet:
			receivingActor(connect, w, r)
		default:
			http.Error(w, "Не правильный http метод", http.StatusMethodNotAllowed)
		}
	}
}

func createActor(connect *sql.DB, w http.ResponseWriter, r *http.Request) {
	var newActor Actor
	err := json.NewDecoder(r.Body).Decode(&newActor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	//valid
	validName := []rune("0123456789")
	validDate := "2006-01-02"

	for _, val := range validName {
		if strings.Contains(newActor.Name, string(val)) {
			fmt.Println(errors.New("Некорректно введено имя. Введите имя без чисел."))
			return
		}
	}
	fmt.Println(newActor.DateOfBirth)
	dateOfBirth, err := time.Parse(validDate, newActor.DateOfBirth)
	if err != nil {
		fmt.Println(dateOfBirth)
		fmt.Println(errors.New("Введите дату в формате 'yyyy-mm-dd'"), err)
		return
	} else {
		newActor.DateOfBirth = dateOfBirth.Format(validDate)
	}

	if newActor.Sex != Male && newActor.Sex != Female {
		fmt.Println(errors.New("Не корректный пол. Введите male или female"))
		return
	}

	id, err := createNew(connect, newActor)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	response := map[string]int{
		"id": id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func createNew(connect *sql.DB, newActor Actor) (int, error) {
	rows := connect.QueryRow(fmt.Sprintf(`INSERT INTO "actor" ("name","sex","date_of_birth") VALUES('%s','%s','%s') RETURNING "id" `, newActor.Name, newActor.Sex, newActor.DateOfBirth))
	var id int

	if err := rows.Scan(&id); err != nil {
		return 0, err
	}

	fmt.Println(fmt.Sprintf(`Успершный запрос: INSERT INTO "actor" ("name","sex","date_of_birth") VALUES('%s','%s','%s') RETURNING "id" `, newActor.Name, newActor.Sex, newActor.DateOfBirth))

	return id, nil
}

func updateActor(connect *sql.DB, w http.ResponseWriter, r *http.Request) {
	var updateActor Actor

	err := json.NewDecoder(r.Body).Decode(&updateActor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	//valid
	validName := []rune("0123456789")
	validDate := "2006-01-02"

	if updateActor.Name != "" {
		for _, val := range validName {
			if strings.Contains(updateActor.Name, string(val)) {
				fmt.Println(errors.New("Некорректно введено имя. Введите имя без чисел."))
				return
			}
		}
	}
	if updateActor.DateOfBirth != "" {
		dateOfBirth, err := time.Parse(validDate, updateActor.DateOfBirth)
		if err != nil {
			fmt.Println(errors.New("Введите дату в формате 'yyyy-mm-dd'"))
			return
		} else {
			updateActor.DateOfBirth = dateOfBirth.Format(validDate)
		}
	}
	if updateActor.Sex != "" {
		if updateActor.Sex != Male && updateActor.Sex != Female {
			fmt.Println(errors.New("Не корректный пол. Введите male или female"))
			return
		}
	}

	updateNew(connect, updateActor)

	w.Header().Set("Content-Type", "application/json")
}

func updateNew(connect *sql.DB, updateActor Actor) {
	//конкотинация
	var whereAfter string
	var setAfter string
	if updateActor.Id != 0 {
		whereAfter = whereAfter + fmt.Sprintf(`id = %d`, updateActor.Id)
	}
	if updateActor.Name != "" {
		setAfter = setAfter + fmt.Sprintf(`name = '%s',`, updateActor.Name)
	}
	if updateActor.Sex != "" {
		setAfter = setAfter + fmt.Sprintf(`sex = '%s',`, updateActor.Sex)
	}
	if updateActor.DateOfBirth != "" {
		setAfter = setAfter + fmt.Sprintf(`date_of_birth = '%s',`, updateActor.DateOfBirth)
	}
	setAfter = strings.TrimRight(setAfter, ",")

	_, er := connect.Exec(fmt.Sprintf(`UPDATE actor SET %s WHERE %s`, setAfter, whereAfter))
	if er != nil {
		fmt.Println(er.Error())
	}

	fmt.Println(fmt.Sprintf(`UPDATE actor SET %s WHERE %s`, setAfter, whereAfter))
}

func deleteActor(connect *sql.DB, w http.ResponseWriter, r *http.Request) {
	var deleteActor Actor

	err := json.NewDecoder(r.Body).Decode(&deleteActor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	deleteNew(connect, deleteActor)

	w.Header().Set("Content-Type", "application/json")
}

func deleteNew(connect *sql.DB, deleteActor Actor) error {
	tx, err := connect.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = connect.Exec(fmt.Sprintf(`DELETE FROM actor_film_participations WHERE actor_id = %v`, deleteActor.Id))
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	_, er := connect.Exec(fmt.Sprintf(`DELETE FROM actor WHERE id = %v`, deleteActor.Id))
	if er != nil {
		fmt.Println(er.Error())
		return err
	}

	err = tx.Commit()
	return err
}

func receivingActor(connect *sql.DB, w http.ResponseWriter, r *http.Request) {
	actor, err := receivingNewActor(connect)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actor)
}

func receivingNewActor(connect *sql.DB) ([]ReceivingActor, error) {
	var slice []ReceivingActor
	var err error

	actor, err := connect.Query(fmt.Sprintf(`SELECT * FROM actor`))
	if err != nil {
		fmt.Println(err.Error())
	}
	for actor.Next() {
		var receivingActor ReceivingActor

		err = actor.Scan(&receivingActor.Id, &receivingActor.Name, &receivingActor.Sex, &receivingActor.DateOfBirth)
		if err != nil {
			fmt.Println(err.Error())
		}
		film_of_actor, err := connect.Query(fmt.Sprintf(`SELECT DISTINCT film.id, film.name, film.description, film.date_of_issue, film.rating
		FROM actor JOIN actor_film_participations ON actor_film_participations.actor_id = actor.id 	JOIN film ON actor_film_participations.film_id= film.id
		WHERE actor.id = %v`, receivingActor.Id))
		if err != nil {
			fmt.Println(err.Error())
		}

		for film_of_actor.Next() {
			var struct_film SearchFilm
			err = film_of_actor.Scan(&struct_film.Id, &struct_film.Name, &struct_film.Description, &struct_film.DateOfIssue, &struct_film.Rating)
			if err != nil {
				fmt.Println(err.Error())
			}
			struct_film.DateOfIssue = strings.TrimRight(struct_film.DateOfIssue, "T00:00:00Z")
			receivingActor.Film = append(receivingActor.Film, struct_film)
		}

		receivingActor.DateOfBirth = strings.TrimRight(receivingActor.DateOfBirth, "T00:00:00Z")
		slice = append(slice, receivingActor)
	}
	return slice, err
}

///////////////////////////////////////////////////////////////

func filmHandler(connect *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			receiptFilm(connect, w, r)
		case http.MethodPost:
			createFilm(connect, w, r)
		case http.MethodPut:
			updateFilm(connect, w, r)
		case http.MethodDelete:
			deleteFilm(connect, w, r)
		default:
			http.Error(w, "Не правильный http метод", http.StatusMethodNotAllowed)
		}
	}
}

func createFilm(connect *sql.DB, w http.ResponseWriter, r *http.Request) {
	var newFilm Film
	err := json.NewDecoder(r.Body).Decode(&newFilm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	//valid
	validDate := "2006-01-02"
	validName := len([]rune(newFilm.Name))
	validDescription := len([]rune(newFilm.Description))

	if 150 < validName || validName < 1 {
		fmt.Println(errors.New("Название фильма не должно быть меньше 1 и больше 150 символов"))
		return
	}
	if 1000 < validDescription {
		fmt.Println(errors.New("Описание фильма не должно быть больше 1000 символов"))
		return
	}
	DateOfIssue, err := time.Parse(validDate, newFilm.DateOfIssue)
	if err != nil {
		fmt.Println(errors.New("Введите дату в формате 'yyyy-mm-dd'"))
		return
	} else {
		newFilm.DateOfIssue = DateOfIssue.Format(validDate)
	}
	if 10 < newFilm.Rating || newFilm.Rating < 0 {
		fmt.Println(errors.New("Рейтинг фильма не должно быть меньше 1 и больше 10"))
		return
	}

	id, err := createNewFilm(connect, newFilm)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	response := map[string]int{
		"id": id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func createNewFilm(connect *sql.DB, newFilm Film) (int, error) {
	tx, err := connect.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	rows := connect.QueryRow(fmt.Sprintf(`INSERT INTO "film" ("name","description","date_of_issue", "rating") VALUES('%s','%s','%s', '%v') RETURNING "id" `, newFilm.Name, newFilm.Description, newFilm.DateOfIssue, newFilm.Rating))

	var id int

	if err := rows.Scan(&id); err != nil {
		return 0, err
	}

	for _, actor := range newFilm.Actor_list {
		_, err := connect.Exec(fmt.Sprintf(`INSERT INTO "actor_film_participations" ("film_id","actor_id") VALUES('%v','%v')`, id, actor))
		if err != nil {
			return 0, err
		}
		fmt.Println(fmt.Sprintf(`INSERT INTO "actor_film_participations" ("film_id","actor_id") VALUES('%v','%v')`, id, actor))
	}

	fmt.Println(fmt.Sprintf(`INSERT INTO "film" ("name","description","date_of_issue", "rating") VALUES('%s','%s','%s', '%v') RETURNING "id" `, newFilm.Name, newFilm.Description, newFilm.DateOfIssue, newFilm.Rating))
	err = tx.Commit()
	return id, err
}

func updateFilm(connect *sql.DB, w http.ResponseWriter, r *http.Request) {

	updateFilm := newFilmWithDefaultRating()

	err := json.NewDecoder(r.Body).Decode(&updateFilm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	//valid
	validDate := "2006-01-02"
	validName := len([]rune(updateFilm.Name))
	validDescription := len([]rune(updateFilm.Description))

	if updateFilm.Name != "" {
		if 150 < validName || validName < 1 {
			fmt.Println(errors.New("Название фильма не должно быть меньше 1 и больше 150 символов"))
			return
		}
	}
	if updateFilm.Description != "" {
		if 1000 < validDescription {
			fmt.Println(errors.New("Описание фильма не должно быть больше 1000 символов"))
			return
		}
	}
	if updateFilm.DateOfIssue != "" {
		DateOfIssue, err := time.Parse(validDate, updateFilm.DateOfIssue)
		if err != nil {
			fmt.Println(errors.New("Введите дату в формате 'yyyy-mm-dd'"))
			return
		} else {
			updateFilm.DateOfIssue = DateOfIssue.Format(validDate)
		}
	}
	if updateFilm.Rating != -1 {
		if 10 < updateFilm.Rating || updateFilm.Rating < 0 {
			fmt.Println(errors.New("Рейтинг фильма не должно быть меньше 0 и больше 10"))
			return
		}
	}
	updateNewFilm(connect, updateFilm)

	w.Header().Set("Content-Type", "application/json")
}

func updateNewFilm(connect *sql.DB, updateFilm Film) {
	//конкотинация
	var whereAfter string
	var setAfter string

	fmt.Println(updateFilm.Rating)
	if updateFilm.Id != 0 {
		whereAfter = whereAfter + fmt.Sprintf(`id = %d`, updateFilm.Id)
	}
	if updateFilm.Name != "" {
		setAfter = setAfter + fmt.Sprintf(`name = '%s',`, updateFilm.Name)
	}
	if updateFilm.Description != "" {
		setAfter = setAfter + fmt.Sprintf(`description = '%s',`, updateFilm.Description)
	}
	if updateFilm.DateOfIssue != "" {
		setAfter = setAfter + fmt.Sprintf(`date_of_issue = '%s',`, updateFilm.DateOfIssue)
	}
	if updateFilm.Rating != -1 {
		setAfter = setAfter + fmt.Sprintf(`rating = '%v',`, updateFilm.Rating)
	}
	setAfter = strings.TrimRight(setAfter, ",")

	_, er := connect.Exec(fmt.Sprintf(`UPDATE film SET %s WHERE %s`, setAfter, whereAfter))
	if er != nil {
		fmt.Println(er.Error())
	}

}

func deleteFilm(connect *sql.DB, w http.ResponseWriter, r *http.Request) {
	var deleteFilm Film

	err := json.NewDecoder(r.Body).Decode(&deleteFilm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	deleteNewFilm(connect, deleteFilm)

	w.Header().Set("Content-Type", "application/json")
}

func deleteNewFilm(connect *sql.DB, deleteFilm Film) error {

	tx, err := connect.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = connect.Exec(fmt.Sprintf(`DELETE FROM actor_film_participations WHERE film_id = %v`, deleteFilm.Id))
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	_, err = connect.Exec(fmt.Sprintf(`DELETE FROM film WHERE id = %v`, deleteFilm.Id))
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	err = tx.Commit()
	return err
}

func receiptFilm(connect *sql.DB, w http.ResponseWriter, r *http.Request) {
	film, err := receiptNewFilm(connect, r)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(film)
}

func receiptNewFilm(connect *sql.DB, r *http.Request) ([]ReceiptFilm, error) {
	sort := r.URL.Query().Get("order_by")
	if sort == "" {
		sort = "rating DESC"
	}

	fmt.Println(fmt.Sprintf(`SELECT * FROM "film" ORDER BY %s`, sort))
	film, err := connect.Query(fmt.Sprintf(`SELECT * FROM "film" ORDER BY %s`, sort))
	if err != nil {
		fmt.Println(err.Error())
	}
	var slice []ReceiptFilm
	for film.Next() {
		var receipt ReceiptFilm
		err = film.Scan(&receipt.Id, &receipt.Name, &receipt.Description, &receipt.DateOfIssue, &receipt.Rating)
		if err != nil {
			fmt.Println(err.Error())
		}
		receipt.DateOfIssue = strings.TrimRight(receipt.DateOfIssue, "T00:00:00Z")
		slice = append(slice, receipt)
	}
	return slice, err
}

func filmHandlerSearches(connect *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			searchFilm(connect, w, r)
		default:
			http.Error(w, "Не правильный http метод", http.StatusMethodNotAllowed)
		}
	}
}

func searchFilm(connect *sql.DB, w http.ResponseWriter, r *http.Request) {
	film, err := searchNewFilm(connect, r)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(film)
}

func searchNewFilm(connect *sql.DB, r *http.Request) ([]SearchFilm, error) {
	searchFilm := r.URL.Query().Get("film")
	searchActor := r.URL.Query().Get("actor")
	prosent := "%"
	var slice []SearchFilm
	var err error
	var request string
	if searchFilm != "" {
		request = fmt.Sprintf(`SELECT * FROM "film" WHERE "name" LIKE '%s%s%s';`, prosent, searchFilm, prosent)
		fmt.Println(request)
	}
	if searchActor != "" {
		request = fmt.Sprintf(
			`SELECT DISTINCT film.id, film.name, film.description, film.date_of_issue, film.rating
		FROM film
		JOIN actor_film_participations ON actor_film_participations.film_id= film.id
		JOIN actor ON actor_film_participations.actor_id = actor.id
		WHERE actor.name like '%s%s%s';`, prosent, searchActor, prosent)
		fmt.Println(request)
	}

	film, err := connect.Query(request)
	if err != nil {
		fmt.Println(err.Error())
	}
	for film.Next() {
		var searchFilm SearchFilm
		err = film.Scan(&searchFilm.Id, &searchFilm.Name, &searchFilm.Description, &searchFilm.DateOfIssue, &searchFilm.Rating)
		if err != nil {
			fmt.Println(err.Error())
		}
		searchFilm.DateOfIssue = strings.TrimRight(searchFilm.DateOfIssue, "T00:00:00Z")
		slice = append(slice, searchFilm)
	}
	return slice, err
}

//////////////////////////////////////////////////

func user(connect *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			authorizationUsers(connect, w, r)
		case http.MethodPost:
			createUser(connect, w, r)
		default:
			http.Error(w, "Не правильный http метод", http.StatusMethodNotAllowed)
		}
	}
}

func createUser(connect *sql.DB, w http.ResponseWriter, r *http.Request) {
	var newUser Users
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	//valid
	validName := []rune("0123456789")

	for _, val := range validName {
		if strings.Contains(newUser.Name, string(val)) {
			fmt.Println(errors.New("Некорректно введено имя. Введите имя без чисел."))
			return
		}
	}

	if strings.Contains(newUser.Mail, "@") == false {
		fmt.Println(errors.New("Некорректно введен mail"))
		return
	}

	id, err := createNewUser(connect, newUser)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	response := map[string]int{
		"id": id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func createNewUser(connect *sql.DB, newUser Users) (int, error) {
	hashPassword := hashPassword(newUser.Password)
	user := connect.QueryRow(fmt.Sprintf(`INSERT INTO "users" ("name","mail","password", "role") VALUES('%s','%s','%s','%s') RETURNING "id" `, newUser.Name, newUser.Mail, hashPassword, newUser.Role))
	var id int

	if err := user.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func hashPassword(password string) []byte {
	pass := []byte(password)
	cost := 10
	hashpassword, err := bcrypt.GenerateFromPassword(pass, cost)
	if err != nil {
		fmt.Println(err.Error())
	}

	return hashpassword
}

func userLogin(connect *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			authorizationUsers(connect, w, r)
		default:
			http.Error(w, "Не правильный http метод", http.StatusMethodNotAllowed)
		}
	}
}

func authorizationUsers(connect *sql.DB, w http.ResponseWriter, r *http.Request) {
	var u Users
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
	}

	id, err := isThereAUser(connect, u)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(id)
}

func isThereAUser(connect *sql.DB, u Users) (int, error) {
	var id int
	var password string
	user := connect.QueryRow(fmt.Sprintf(`SELECT "id","password" FROM "users" WHERE "mail" = '%s' `, u.Mail))
	err := user.Scan(&id, &password)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(u.Password))
	if err != nil {
		fmt.Println(err.Error())
		return 0, err
	}
	return id, err
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Сервер работает корректно")
}
