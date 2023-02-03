package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/microcosm-cc/bluemonday"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
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

func (page *Page) RenderIndex(w http.ResponseWriter, r *http.Request) {
	var sessionNpub string
	var posts []*Post
	var err error
	if r.Context().Value(KeySessionNpub) != nil {
		sessionNpub = r.Context().Value(KeySessionNpub).(string)
		posts, err = page.store.GetAllPostByNpub(r.Context(), sessionNpub)
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
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(
		"internal/templates/layout.html",
		"internal/templates/index.html",
	)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("cannot compile index template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, map[string]interface{}{
		"SessionNpub": r.Context().Value(KeySessionNpub),
		"PostList":    posts,
	})
	if err != nil {
		panic(err)
	}
}

func (page *Page) Login(w http.ResponseWriter, r *http.Request) {
	// set npub cookie, serving the role of a session token
	cookie := http.Cookie{
		Name:     "npub",
		Value:    r.FormValue("npub"),
		Path:     "/",
		HttpOnly: false,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (page *Page) Logout(w http.ResponseWriter, r *http.Request) {
	// delete cookie by setting a new one with same name and max age < 0
	cookie := http.Cookie{
		Name:     "npub",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: false,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
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
	err = t.Execute(w, map[string]interface{}{
		"SessionNpub": r.Context().Value(KeySessionNpub),
	})
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
	postBody := strings.Replace(post.Body, "\r\n", "\n", -1) // fix line breaks
	unsafeHTML := blackfriday.Run([]byte(postBody))
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
		"SessionNpub": r.Context().Value(KeySessionNpub),
		"Post":        post,
		"BodyHTML":    template.HTML(bodyHTML),
	})
	if err != nil {
		panic(err)
	}
}

func (page *Page) RenderAllPost(w http.ResponseWriter, r *http.Request) {
	sessionNpub := r.Context().Value(KeySessionNpub).(string)
	posts, err := page.store.GetAllPostByNpub(r.Context(), sessionNpub)
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
		"SessionNpub": sessionNpub,
		"PostList":    posts,
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
	defaultRelays := `wss://nostr01.opencult.com`
	err = t.Execute(w, map[string]interface{}{
		"SessionNpub":      r.Context().Value(KeySessionNpub),
		"DefaultRelayList": defaultRelays,
	})
	if err != nil {
		panic(err)
	}
}

func (page *Page) SaveNewPost(w http.ResponseWriter, r *http.Request) {
	// build data
	var data struct {
		Body      string
		Nsec      string
		RelayList string
	}
	data.Body = r.FormValue("body")
	data.Nsec = r.FormValue("nsec")
	data.RelayList = r.FormValue("relaylist")
	if data.Body == "" || data.Nsec == "" || data.RelayList == "" {
		http.Error(w, "all fields are required", 400)
		return
	}

	// decode nsec
	prefix, sk, err := nip19.Decode(data.Nsec)
	if err != nil {
		http.Error(w, "cannot decode key", 400)
		return
	}
	if prefix != "nsec" {
		http.Error(w, "key is not nsec", 400)
		return
	}
	pk, err := nostr.GetPublicKey(sk.(string))
	if err != nil {
		http.Error(w, "cannot get public key", 400)
		return
	}
	npub, err := nip19.EncodePublicKey(pk)
	if err != nil {
		http.Error(w, "cannot encode public key", 400)
		return
	}
	if npub != r.Context().Value(KeySessionNpub).(string) {
		http.Error(w, "nsec does not match to logged in npub", 400)
		return
	}

	// create on database
	now := time.Now()
	post := &Post{
		Npub:      npub,
		Body:      data.Body,
		RelayList: data.RelayList,
		CreatedAt: now,
	}
	_, err = page.store.InsertPost(r.Context(), post)
	if err != nil {
		panic(err)
	}

	// sign nostr event
	ev := nostr.Event{
		PubKey:    pk,
		CreatedAt: time.Now(),
		Kind:      1,
		Tags:      nil,
		Content:   data.Body,
	}
	err = ev.Sign(sk.(string)) // sign sets the event ID field and the event Sig field
	if err != nil {
		http.Error(w, "cannot sign nostr event", 500)
		return
	}

	// publish event to relays
	relayArr := strings.Split(data.RelayList, "\n")
	for _, url := range relayArr {
		trimmedUrl := strings.Trim(url, "\r\n\t")
		relay, err := nostr.RelayConnect(context.Background(), trimmedUrl)
		if err != nil {
			message := "cannot connect to " + trimmedUrl
			http.Error(w, message, 400)
			return
		}
		fmt.Println("published to ", url, relay.Publish(context.Background(), ev))
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (page *Page) RenderAbout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles(
		"internal/templates/layout.html",
		"internal/templates/about.html",
	)
	if err != nil {
		page.logger.With(
			zap.Error(err),
		).Error("cannot compile about template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, map[string]interface{}{
		"SessionNpub": r.Context().Value(KeySessionNpub),
	})
	if err != nil {
		panic(err)
	}
}
