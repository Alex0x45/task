package task

import (
	"time"
    "net/http"
    "html/template"
    "appengine"
    "appengine/datastore"
    "appengine/user"
)

type Task struct {
	User string
	Desc string
	Created	time.Time
}  

func init() {
    http.HandleFunc("/", viewHandler)
    http.HandleFunc("/add/", addHandler)
    http.HandleFunc("/save/", saveHandler)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "add.html", nil); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    usr := user.Current(c)
	if desc := r.FormValue("desc"); desc != "" && usr != nil {
    	t := Task{usr.String(), r.FormValue("desc"), time.Now()}
		key := datastore.NewIncompleteKey(c, "Task", taskUserKey(c, usr.String()))
        _, err := datastore.Put(c, key, &t)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func taskUserKey(c appengine.Context, userName string) *datastore.Key {
        return datastore.NewKey(c, "Task", userName, 0, nil)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    usr := user.Current(c)
    if usr == nil {
        url, err := user.LoginURL(c, r.URL.String())
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Location", url)
        w.WriteHeader(http.StatusFound)
        return
    }
    q := datastore.NewQuery("Task").Ancestor(taskUserKey(c, usr.String())).Order("-Created").Limit(20)
    tasks := make([]Task, 0, 20)
    if _, err := q.GetAll(c, &tasks); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if err := templates.ExecuteTemplate(w, "view.html", tasks); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

var templates = template.Must(template.ParseFiles("add.html", "view.html"))
