package routes

import (
	"database/sql"
	"embed"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/svuvi/theweek/layouts"
	"github.com/svuvi/theweek/models"
	"github.com/svuvi/theweek/repositories"
)

type BaseHandler struct {
	articleRepo models.ArticleRepository
	userRepo    models.UserRepository
	sessionRepo models.SessionRepository
	inviteRepo  models.InviteRepository
	imageRepo   models.ImageRepository
}

func NewBaseHandler(db *sql.DB) *BaseHandler {
	return &BaseHandler{
		articleRepo: repositories.NewArticleRepo(db),
		userRepo:    repositories.NewUserRepo(db),
		sessionRepo: repositories.NewSessionRepo(db),
		inviteRepo:  repositories.NewInviteRepo(db),
		imageRepo:   repositories.NewImageRepo(db),
	}
}

//go:embed static
var static embed.FS

func (h *BaseHandler) NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", h.indexHandler)
	mux.HandleFunc("GET /{slug}", h.articleHandler)

	mux.HandleFunc("GET /login", h.loginPageHandler)
	mux.HandleFunc("POST /login", h.loginFormHandler)

	mux.HandleFunc("GET /logout", h.logoutHandler)

	mux.HandleFunc("GET /invite/{code}", h.claimInvite)

	mux.HandleFunc("GET /register", h.registrationPageHandler)
	mux.HandleFunc("POST /register", h.registrationFormHandler)

	mux.HandleFunc("GET /dashboard/", h.dasboardPageHandler)
	mux.HandleFunc("GET /dashboard/users/", h.dashboardUsersHandler)
	mux.HandleFunc("GET /dashboard/invites/", h.dashboardInvitesHandler)
	mux.HandleFunc("POST /dashboard/invites/create", h.createInvite)
	mux.HandleFunc("DELETE /dashboard/invites/delete/{code}", h.deleteInvite)
	mux.HandleFunc("GET /dashboard/publishing/", h.dashboardPublishing)
	mux.HandleFunc("POST /dashboard/publishing/", h.publishingFormHandler)

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

func (h *BaseHandler) claimInvite(w http.ResponseWriter, r *http.Request) {
	authorized, _ := isAuthorised(r, h)
	if authorized {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	code := r.PathValue("code")

	if err := uuid.Validate(code); err != nil {
		http.Error(w, "Невалидный формат кода", http.StatusBadRequest)
		return
	}

	invite, err := h.inviteRepo.GetByCode(code)
	if err != nil {
		http.Error(w, "Вас не приглашали", http.StatusNotFound)
		return
	}

	if !invite.IsActive {
		http.Error(w, "Приглашение уже использовано", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "registration_invite",
		Value:    code,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   900,
	})

	http.Redirect(w, r, "/register", http.StatusFound)
}

func (h *BaseHandler) registrationPageHandler(w http.ResponseWriter, r *http.Request) {
	authorized, user := isAuthorised(r, h)
	if authorized {
		layouts.AlreadyRegisteredPage(user).Render(r.Context(), w)
		return
	}

	code, err := getInviteCode(r)
	log.Println(code)
	if err != nil {
		layouts.RegistrationNoInvite(false).Render(r.Context(), w)
		return
	}

	invite, err := h.inviteRepo.GetByCode(code)
	if !invite.IsActive || err != nil {
		layouts.RegistrationNoInvite(true).Render(r.Context(), w)
		return
	}

	layouts.RegistrationPage().Render(r.Context(), w)
}
