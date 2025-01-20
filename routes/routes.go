package routes

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/svuvi/theweek/components"
	"github.com/svuvi/theweek/layouts"
	"github.com/svuvi/theweek/models"
	"github.com/svuvi/theweek/repositories"
)

type BaseHandler struct {
	articleRepo      models.ArticleRepository
	userRepo         models.UserRepository
	sessionRepo      models.SessionRepository
	inviteRepo       models.InviteRepository
	imageRepo        models.ImageRepository
	recoveryCodeRepo models.RecoveryCodeRepository
}

func NewBaseHandler(db *sql.DB) *BaseHandler {
	return &BaseHandler{
		articleRepo:      repositories.NewArticleRepo(db),
		userRepo:         repositories.NewUserRepo(db),
		sessionRepo:      repositories.NewSessionRepo(db),
		inviteRepo:       repositories.NewInviteRepo(db),
		imageRepo:        repositories.NewImageRepo(db),
		recoveryCodeRepo: repositories.NewRecoveryCodeRepo(db),
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

	mux.HandleFunc("GET /account/", h.accountPage)
	mux.HandleFunc("GET /account/change-password", h.changePasswordPage)
	mux.HandleFunc("POST /account/change-password", h.changePasswordForm)
	mux.HandleFunc("GET /account/restore-password", h.restorePasswordPage)
	mux.HandleFunc("POST /account/restore-password", h.restorePasswordForm)

	mux.HandleFunc("GET /dashboard/", h.dasboardPageHandler)
	mux.HandleFunc("GET /dashboard/users/", h.dashboardUsersHandler)
	mux.HandleFunc("GET /dashboard/invites/", h.dashboardInvitesHandler)
	mux.HandleFunc("POST /dashboard/invites/create", h.createInvite)
	mux.HandleFunc("DELETE /dashboard/invites/delete/{code}", h.deleteInvite)
	mux.HandleFunc("POST /dashboard/reocvery-codes/create", h.createRecoveryCodeForm)
	mux.HandleFunc("DELETE /dashboard/reocvery-codes/delete/{rCodeID}", h.deleteRecoveryCode)
	mux.HandleFunc("GET /dashboard/publishing/", h.dashboardPublishing)
	mux.HandleFunc("GET /dashboard/publishing/{articleID}", h.dashboardPublishing)
	mux.HandleFunc("POST /dashboard/publishing/", h.publishingFormHandler)
	mux.HandleFunc("POST /dashboard/publishing/{articleID}", h.publishingFormHandler)

	mux.HandleFunc("/delete/{type}/{id}", h.deleteResourceHandler)

	mux.HandleFunc("GET /images/{imageID}", h.imageHandler)
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
		http.NotFound(w, r)
		return
	}

	/* coverImageName, err := h.imageRepo.GetName(article.CoverImageID)
	if err != nil {
		log.Print("Ошибка при попытке получить имя картинки из БД:\n", err)
	}

	var coverImagePath string
	if coverImageName != "" {
		coverImagePath = fmt.Sprintf("/images/%d/%s", article.CoverImageID, coverImageName)
	} else {
		coverImagePath = ""
	} */

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

func (h *BaseHandler) imageHandler(w http.ResponseWriter, r *http.Request) {
	i, err := strconv.ParseInt(r.PathValue("imageID"), 10, 64)
	id := int(i)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	/* filename := r.PathValue("filename")
	if filename == "" {
		http.NotFound(w, r)
		return
	} */

	img, err := h.imageRepo.Get(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	/* if img.Filename != filename {
		http.NotFound(w, r)
		return
	} */

	w.Header().Set("Content-Type", http.DetectContentType(img.Content))
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, img.Filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(img.Content)))

	_, err = w.Write(img.Content)
	if err != nil {
		http.Error(w, "Ошибка при отправке файла", http.StatusInternalServerError)
	}
}

func (h *BaseHandler) deleteResourceHandler(w http.ResponseWriter, r *http.Request) {
	_, user := isAuthorised(r, h)
	if !user.IsAdmin {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	idValue := r.PathValue("id")
	typeString := r.PathValue("type")
	log.Printf("Администратор %s запросил удаление %s с id=%s", user.Username, typeString, idValue)

	id, err := strconv.Atoi(idValue)
	if err != nil || id < 1 || typeString != "article" { // todo: добавить картинки и другие типы ресурсов по надобности
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if typeString == "article" {
		err = h.articleRepo.Delete(id)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		components.ArticleDeleted().Render(r.Context(), w)
		return
	}
}

func (h *BaseHandler) accountPage(w http.ResponseWriter, r *http.Request) {
	authorised, user := isAuthorised(r, h)
	if !authorised {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	layouts.AccountPage(user).Render(r.Context(), w)
}

func (h *BaseHandler) changePasswordPage(w http.ResponseWriter, r *http.Request) {
	authorised, user := isAuthorised(r, h)
	if !authorised {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	layouts.ChangePasswordPage(user).Render(r.Context(), w)
}
