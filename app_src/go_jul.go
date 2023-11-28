package main

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
var key = []byte("The#Rime#of#the#Ancient#Mariner")
var store = sessions.NewCookieStore(key)

type predrawPageData struct {
	// Capital first letter to expose it so that i can be called with .
	Selected   []string
	Unselected []string
}

type drawPageData struct {
	TransformY string
	Duration   string
	Delay      string
	Con        []string
	Won        []string
	Res        int
}

type countdownPageData struct {
	Won []string
}

func isElementExist(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func verifyUser(users [][]string, user string, pass string) bool {
	for _, u := range users {
		if u[0] == user && u[1] == pass {
			return true
		}
	}
	return false
}

func readCsv(csvFile string) ([][]string, error) {
	f, err := os.Open(csvFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func writeCsv(csvFile string, csvArray [][]string) error {
	f, err := os.OpenFile(csvFile, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	writer := csv.NewWriter(f)
	if err != nil {
		return err
	}
	writer.WriteAll(csvArray)
	writer.Flush()
	return nil
}

func login(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-cookie")
	auth, _ := session.Values["authenticated"].(bool)
	logger := log.WithFields(logrus.Fields{
		"remote_ip":    r.RemoteAddr,
		"method":       r.Method,
		"uri":          r.RequestURI,
		"autenticated": auth,
		"user":         session.Values["user"],
	})
	logger.Info("Request")
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/login.html")
		if err != nil {
			logger.Error(err)
			errCode := http.StatusInternalServerError
			http.Error(w, http.StatusText(errCode), errCode)
		} else {
			t.Execute(w, nil)
		}
	} else if r.Method == "POST" {
		user := r.PostFormValue("user")
		pass := r.PostFormValue("pass")
		usersArray, err := readCsv("config/users.csv")
		if err != nil {
			logger.Error(err)
			errCode := http.StatusInternalServerError
			http.Error(w, http.StatusText(errCode), errCode)
		} else if verifyUser(usersArray, user, pass) {
			logger.Info("Autentication successful")
			session.Values["user"] = user
			session.Values["authenticated"] = true
			session.Save(r, w)
			http.Redirect(w, r, "/select", http.StatusSeeOther)
		} else {
			logger.Warn("Autentication failed")
			t, err := template.ParseFiles("templates/login.html")
			if err != nil {
				logger.Error(err)
				errCode := http.StatusInternalServerError
				http.Error(w, http.StatusText(errCode), errCode)
			} else {
				t.Execute(w, nil)
			}
		}
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-cookie")
	auth, _ := session.Values["authenticated"].(bool)
	logger := log.WithFields(logrus.Fields{
		"remote_ip":    r.RemoteAddr,
		"method":       r.Method,
		"uri":          r.RequestURI,
		"autenticated": auth,
		"user":         session.Values["user"],
	})
	logger.Info("Request")
	session.Values["authenticated"] = false
	session.Save(r, w)
	http.Redirect(w, r, "/countdown", http.StatusSeeOther)
}

func draw(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-cookie")
	auth, auth_ok := session.Values["authenticated"].(bool)
	logger := log.WithFields(logrus.Fields{
		"remote_ip":    r.RemoteAddr,
		"method":       r.Method,
		"uri":          r.RequestURI,
		"autenticated": auth,
		"user":         session.Values["user"],
	})
	logger.Info("Request")
	if !auth_ok || !auth {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	user := session.Values["user"]
	con := []string{}
	conFile := fmt.Sprintf("config/%s_con.csv", user)
	conArray, err := readCsv(conFile)
	if err != nil {
		if err != nil {
			logger.Error(err)
			errCode := http.StatusInternalServerError
			http.Error(w, http.StatusText(errCode), errCode)
			return
		}
		errCode := http.StatusInternalServerError
		http.Error(w, http.StatusText(errCode), errCode)
		return
	}
	won := []string{}
	wonFile := fmt.Sprintf("config/%s_won.csv", user)
	wonArray, err := readCsv(wonFile)
	if err != nil {
		logger.Error(err)
		errCode := http.StatusInternalServerError
		http.Error(w, http.StatusText(errCode), errCode)
		return
	}
	action := mux.Vars(r)["action"]
	if action == "draw" {
		err := r.ParseForm()
		if err != nil {
			logger.Error(err)
			errCode := http.StatusInternalServerError
			http.Error(w, http.StatusText(errCode), errCode)
			return
		}
		session.Values["selected"] = r.Form["selected"]
		session.Save(r, w)
		con = r.Form["selected"]
	} else if action == "redraw" {
		if cs, ok := session.Values["selected"].([]string); ok && len(cs) > 0 {
			con = cs
		} else {
			http.Redirect(w, r, "/select", http.StatusSeeOther)
			return
		}
	} else if action == "add" {
		err := r.ParseForm()
		if err != nil {
			logger.Error(err)
			errCode := http.StatusInternalServerError
			http.Error(w, http.StatusText(errCode), errCode)
			return
		}
		if cs, ok := session.Values["selected"].([]string); ok && len(cs) > 0 {
			con = cs
		} else {
			logger.Error(ok)
			errCode := http.StatusInternalServerError
			http.Error(w, http.StatusText(errCode), errCode)
			return
		}

		add, _ := strconv.Atoi(r.Form["add"][0])
		// addArray := []string{conArray[add]}
		wonArray = append(wonArray, []string{con[add]})
		err = writeCsv(wonFile, wonArray)
		if err != nil {
			logger.Error(err)
			errCode := http.StatusInternalServerError
			http.Error(w, http.StatusText(errCode), errCode)
			return
		} else {
			logger.Error(con[add] + " added to won list")
		}
		var addi int
		for i, v := range conArray {
			if v[0] == con[add] {
				addi = i
				break
			}
		}
		conArray[addi][1] = "0"
		err = writeCsv(conFile, conArray)
		if err != nil {
			logger.Error(err)
			errCode := http.StatusInternalServerError
			http.Error(w, http.StatusText(errCode), errCode)
			return
		}
		http.Redirect(w, r, "/select", http.StatusSeeOther)
		return
	} else {
		logger.Error("Unknown action")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	for _, v := range wonArray {
		won = append(won, v[0])
	}
	rand.Seed(time.Now().UnixNano())
	randIndex := rand.Intn(len(con))
	cr := make([]string, len(con))
	copy(cr, con)
	rand.Shuffle(len(cr), func(i, j int) { cr[i], cr[j] = cr[j], cr[i] })
	conTotal := []string{}
	for i := 1; i <= 6; i++ {
		conTotal = append(conTotal, cr...)
	}
	// uncomment to cheat
	// override := "override user"
	// // check if override is in the con list
	// if isElementExist(conTotal, override) && !isElementExist(won, override) {
	// 	// find override in the con list and overwrite the randomly picked user
	// 	for i, v := range con {
	// 		if v == override {
	// 			randIndex = i
	// 		}
	// 	}
	// }
	// avoid the same name twice in a row
	if conTotal[len(conTotal)-1] != con[randIndex] {
		conTotal = append(conTotal, con[randIndex])
	}
	// fmt.Println(conTotal)
	// fmt.Println(len(conTotal))
	transformY := (100 / float64(len(conTotal))) * (float64(len(conTotal)) - 1)
	data := drawPageData{
		Delay:      fmt.Sprintf("-%ds", len(conTotal)/21),
		Duration:   fmt.Sprintf("%ds", len(conTotal)/6),
		TransformY: fmt.Sprintf("-%f%%", transformY),
		Con:        conTotal,
		Won:        won,
		Res:        randIndex,
	}
	t, err := template.ParseFiles("templates/draw.html")
	if err != nil {
		logger.Error(err)
		errCode := http.StatusInternalServerError
		http.Error(w, http.StatusText(errCode), errCode)
	} else {
		t.Execute(w, data)
	}
}

func predraw(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-cookie")
	auth, auth_ok := session.Values["authenticated"].(bool)
	logger := log.WithFields(logrus.Fields{
		"remote_ip":    r.RemoteAddr,
		"method":       r.Method,
		"uri":          r.RequestURI,
		"autenticated": auth,
		"user":         session.Values["user"],
	})
	logger.Info("Request")
	if !auth_ok || !auth {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	user := session.Values["user"]
	conFile := fmt.Sprintf("config/%s_con.csv", user)
	conArray, err := readCsv(conFile)
	if err != nil {
		logger.Error(err)
		errCode := http.StatusInternalServerError
		http.Error(w, http.StatusText(errCode), errCode)
		return
	}
	selected := []string{}
	unselected := []string{}
	// won := []string{}
	// wonMap := map[string]bool{}
	// for _, v := range wonArray {
	//  wonMap[v[0]] = true
	//  won = append(won, v[0])
	// }
	// con := []string{}
	// for _, v := range conArray {
	//  if _, ok := wonMap[v[0]]; !ok {
	//      con = append(con, v[0])
	//  }
	// }

	for _, v := range conArray {
		if v[1] == "1" {
			selected = append(selected, v[0])
		} else {
			unselected = append(unselected, v[0])
		}
	}

	data := predrawPageData{
		Selected:   selected,
		Unselected: unselected,
	}
	t, err := template.ParseFiles("templates/predraw.html")
	if err != nil {
		logger.Error(err)
		errCode := http.StatusInternalServerError
		http.Error(w, http.StatusText(errCode), errCode)
	} else {
		t.Execute(w, data)
	}
}

func countdown(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-cookie")
	auth, _ := session.Values["authenticated"].(bool)
	logger := log.WithFields(logrus.Fields{
		"remote_ip":    r.RemoteAddr,
		"method":       r.Method,
		"uri":          r.RequestURI,
		"autenticated": auth,
		"user":         session.Values["user"],
	})
	logger.Info("Request")
	won := []string{}
	user := session.Values["user"]
	if user != nil {
		wonFile := fmt.Sprintf("config/%s_won.csv", user)
		wonArray, err := readCsv(wonFile)
		if err != nil {
			logger.Error(err)
			errCode := http.StatusInternalServerError
			http.Error(w, http.StatusText(errCode), errCode)
			return
		}
		for _, v := range wonArray {
			won = append(won, v[0])
		}
	}
	data := countdownPageData{
		Won: won,
	}
	t, err := template.ParseFiles("templates/countdown.html")
	if err != nil {
		logger.Error(err)
		errCode := http.StatusInternalServerError
		http.Error(w, http.StatusText(errCode), errCode)
	} else {
		t.Execute(w, data)
	}
}

func main() {
	log.SetFormatter(&logrus.JSONFormatter{})
	r := mux.NewRouter()
	r.HandleFunc("/", countdown).Methods("GET")
	r.HandleFunc("/login", login).Methods("GET", "POST")
	r.HandleFunc("/logout", logout).Methods("GET")
	r.HandleFunc("/select", predraw).Methods("GET")
	r.HandleFunc("/d/{action}", draw).Methods("POST")
	r.HandleFunc("/countdown", countdown).Methods("GET")
	s := http.StripPrefix("/static", http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/static").Handler(s)
	http.ListenAndServe(":9000", r)
}
