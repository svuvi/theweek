package routes

import (
	"database/sql"
	"embed"
	"net/http"

	"github.com/svuvi/theweek/layouts"
	"github.com/svuvi/theweek/models"
	"github.com/svuvi/theweek/repositories"
)

type BaseHandler struct {
	articleRepo models.ArticleRepository
	userRepo    models.UserRepository
	sessionRepo models.SessionRepository
}

func NewBaseHandler(db *sql.DB) *BaseHandler {
	return &BaseHandler{
		articleRepo: repositories.NewArticleRepo(db),
		userRepo:    repositories.NewUserRepo(db),
		sessionRepo: repositories.NewSessionRepo(db),
	}
}

//go:embed static
var static embed.FS

func (h *BaseHandler) NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", h.indexHandler)
	mux.HandleFunc("GET /{slug}", h.articleHandler)

	mux.HandleFunc("GET /login", h.loginPageHandler)
	mux.HandleFunc("POST /login", h.loginFormHandler)

	mux.HandleFunc("GET /logout", h.logoutHandler)

	mux.HandleFunc("GET /register", h.registrationPageHandler)
	mux.HandleFunc("POST /register", h.registrationFormHandler)

	mux.Handle("GET /static/", http.FileServerFS(static))

	return mux
}

func (h *BaseHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
	authorized, user := isAuthorised(r, h)

	articles, err := h.articleRepo.GetAll()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	layouts.Index(articles, authorized, user).Render(r.Context(), w)
}

func (h *BaseHandler) articleHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	article, err := h.articleRepo.GetBySlug(slug)
	if err != nil {
		http.Error(w, "404 Страница не найдена", http.StatusNotFound)
		return
	}

	authorized, user := isAuthorised(r, h)
	layouts.Article(article, authorized, user).Render(r.Context(), w)
}

func (h *BaseHandler) loginPageHandler(w http.ResponseWriter, r *http.Request) {
	authorized, user := isAuthorised(r, h)
	layouts.LoginPage(authorized, user).Render(r.Context(), w)
}

func (h *BaseHandler) registrationPageHandler(w http.ResponseWriter, r *http.Request) {
	authorized, user := isAuthorised(r, h)
	layouts.RegistrationPage(authorized, user).Render(r.Context(), w)
}
