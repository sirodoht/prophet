package internal

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"

	"go.uber.org/zap"
)

type Page struct {
	store  *SQLStore
	logger *zap.Logger
}

func NewHandlers(store *SQLStore) *Page {
	return &Page{
		store: store,
	}
}

func (page *Page) RenderDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(
		"internal/templates/layout.html",
		"internal/templates/dashboard.html",
	)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("cannot compile dashboard template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		panic(err)
	}
}

func (page *Page) RenderOnePost(w http.ResponseWriter, r *http.Request) {
	// parse url post id
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("invalid id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// get post from database
	post, err := page.store.GetOnePost(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		page.logger.With(
			zap.Error(err),
		).Error("failed to get post")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// compile markdown to html
	unsafeHTML := blackfriday.Run([]byte(post.Body))
	bodyHTML := bluemonday.UGCPolicy().SanitizeBytes(unsafeHTML)

	// respond
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(
		"internal/templates/layout.html",
		"internal/templates/post.html",
	)
	if err != nil {
		panic(err)
	}
	err = t.Execute(w, map[string]interface{}{
		"Post":     post,
		"BodyHTML": template.HTML(bodyHTML),
	})
	if err != nil {
		panic(err)
	}
}

func (page *Page) RenderAllPost(w http.ResponseWriter, r *http.Request) {
	posts, err := page.store.GetAllPost(r.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		page.logger.With(
			zap.Error(err),
		).Error("failed to get all posts")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(
		"internal/templates/layout.html",
		"internal/templates/post_list.html",
	)
	if err != nil {
		panic(err)
	}
	err = t.Execute(w, map[string]interface{}{
		"PostList": posts,
	})
	if err != nil {
		panic(err)
	}
}

func (page *Page) RenderNewPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(
		"internal/templates/layout.html",
		"internal/templates/post_new.html",
	)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("cannot compile post new template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		panic(err)
	}
}

func (page *Page) SaveNewPost(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	body := r.FormValue("body")

	type ReqBody struct {
		Title string
		Body  string
	}
	rb := &ReqBody{
		Title: title,
		Body:  body,
	}
	fmt.Printf("%+v", rb)

	if rb.Title == "" || rb.Body == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now()
	d := &Post{
		Title:     rb.Title,
		Body:      rb.Body,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := page.store.InsertPost(r.Context(), d)
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (page *Page) RenderEditPost(w http.ResponseWriter, r *http.Request) {
	// parse url post id
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("invalid id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// get post based on url id
	post, err := page.store.GetOnePost(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		page.logger.With(
			zap.Error(err),
		).Error("failed to get post")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// render
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(
		"internal/templates/layout.html",
		"internal/templates/post_edit.html",
	)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("cannot compile post edit template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, map[string]interface{}{
		"Post": post,
	})
	if err != nil {
		panic(err)
	}
}

func (page *Page) SaveEditPost(w http.ResponseWriter, r *http.Request) {
	// parse post id from url
	idAsString := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idAsString, 10, 64)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("invalid id")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// gather post form data
	var data struct {
		Title string
		Body  string
	}
	data.Title = r.FormValue("title")
	data.Body = r.FormValue("body")

	// validate data
	if data.Title == "" || data.Body == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// write updated post on database
	err = page.store.UpdatePost(r.Context(), id, "title", data.Title)
	if err != nil {
		panic(err)
	}
	err = page.store.UpdatePost(r.Context(), id, "body", data.Body)
	if err != nil {
		panic(err)
	}

	// respond
	http.Redirect(w, r, "/posts/"+idAsString, http.StatusFound)
}
