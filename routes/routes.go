package routes

import (
	"database/sql"
	"embed"
	"log"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/svuvi/theweek/components"
	"github.com/svuvi/theweek/layouts"
	"github.com/svuvi/theweek/models"
	"github.com/svuvi/theweek/repositories"
)

type BaseHandler struct {
	articleRepo models.ArticleRepository
	userRepo    models.UserRepository
	sessionRepo models.SessionRepository
	inviteRepo  models.InviteRepository
}

func NewBaseHandler(db *sql.DB) *BaseHandler {
	return &BaseHandler{
		articleRepo: repositories.NewArticleRepo(db),
		userRepo:    repositories.NewUserRepo(db),
		sessionRepo: repositories.NewSessionRepo(db),
		inviteRepo:  repositories.NewInviteRepo(db),
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

	mux.HandleFunc("GET /dashboard/", h.DasboardPageHandler)
	mux.HandleFunc("GET /dashboard/users/", h.DashboardUsersHandler)
	mux.HandleFunc("GET /dashboard/invites/", h.DashboardInvitesHandler)
	mux.HandleFunc("POST /dashboard/invites/create", h.CreateInvite)
	mux.HandleFunc("DELETE /dashboard/invites/delete/{code}", h.DeleteInvite)

	mux.HandleFunc("GET /write", h.writingPageHandler)
	mux.HandleFunc("POST /write", h.writingFormHandler)

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

func (h *BaseHandler) writingPageHandler(w http.ResponseWriter, r *http.Request) {
	authorized, user := isAuthorised(r, h)
	layouts.WritingPage(authorized, user).Render(r.Context(), w)
}

func (h *BaseHandler) writingFormHandler(w http.ResponseWriter, r *http.Request) {
	authorized, user := isAuthorised(r, h)
	if !authorized || !user.IsAdmin {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	r.ParseForm()

	slug := r.PostFormValue("slug")
	title := r.PostFormValue("title")
	textMD := r.PostFormValue("textMD")

	re := regexp.MustCompile(`^[a-z0-9-]+$`)
	match := re.MatchString(slug)
	if !match {
		slugResult := components.FormWarning("Ссылка может содержать только маленькие латинские буквы, цифры и знак \"-\"")
		components.WritingForm(slugResult, slug, title, textMD).Render(r.Context(), w)
		return
	}

	_, err := h.articleRepo.GetBySlug(slug)
	if err == nil {
		slugResult := components.FormWarning("Эта ссылка уже занята")
		components.WritingForm(slugResult, slug, title, textMD).Render(r.Context(), w)
		return
	}

	err = h.articleRepo.Create(slug, title, textMD)
	if err != nil {
		slugResult := components.FormWarning("Внутренняя ошибка сервера")
		components.WritingForm(slugResult, slug, title, textMD).Render(r.Context(), w)
		return
	}

	components.WritingSuccessful(slug).Render(r.Context(), w)
}
