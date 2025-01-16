package routes

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/svuvi/theweek/components"
	"github.com/svuvi/theweek/layouts"
	"github.com/svuvi/theweek/models"
)

func (h *BaseHandler) dasboardPageHandler(w http.ResponseWriter, r *http.Request) {
	authorized, user := isAuthorised(r, h)
	if !authorized || !user.IsAdmin {
		http.Error(w, "Отказано в доступе", http.StatusUnauthorized)
		return
	}

	layouts.DashboardHome().Render(r.Context(), w)
}

func (h *BaseHandler) dashboardUsersHandler(w http.ResponseWriter, r *http.Request) {
	authorized, user := isAuthorised(r, h)
	if !authorized || !user.IsAdmin {
		http.Error(w, "Отказано в доступе", http.StatusUnauthorized)
		return
	}

	users, err := h.userRepo.GetAll()
	if err != nil {
		http.Error(w, "Ошибка при попытке загрузить пользователей", http.StatusInternalServerError)
		return
	}

	layouts.DashboardUsers(users).Render(r.Context(), w)
}

func (h *BaseHandler) dashboardInvitesHandler(w http.ResponseWriter, r *http.Request) {
	authorized, user := isAuthorised(r, h)
	if !authorized || !user.IsAdmin {
		http.Error(w, "Отказано в доступе", http.StatusUnauthorized)
		return
	}

	invites, err := h.inviteRepo.GetAll()
	if err != nil {
		http.Error(w, "Ошибка при попытке загрузить приглашения", http.StatusInternalServerError)
		return
	}
	layouts.DashboardInvites(invites).Render(r.Context(), w)
}

func (h *BaseHandler) createInvite(w http.ResponseWriter, r *http.Request) {
	authorized, user := isAuthorised(r, h)
	if !authorized || !user.IsAdmin {
		http.Error(w, "Отказано в доступе", http.StatusUnauthorized)
		return
	}

	_, err := h.inviteRepo.Create()
	if err != nil {
		http.Error(w, "Ошибка при попытке создать приглашение", http.StatusInternalServerError)
		return
	}

	invites, err := h.inviteRepo.GetAll()
	if err != nil {
		http.Error(w, "Ошибка при попытке загрузить приглашения", http.StatusInternalServerError)
		return
	}

	components.InviteTable(invites).Render(r.Context(), w)
}

func (h *BaseHandler) deleteInvite(w http.ResponseWriter, r *http.Request) {
	authorized, user := isAuthorised(r, h)
	if !authorized || !user.IsAdmin {
		http.Error(w, "Отказано в доступе", http.StatusUnauthorized)
		return
	}

	if err := uuid.Validate(r.PathValue("code")); err != nil {
		http.Error(w, "Невалидный формат кода", http.StatusBadRequest)
		return
	}

	if err := h.inviteRepo.Delete(r.PathValue("code")); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BaseHandler) dashboardPublishing(w http.ResponseWriter, r *http.Request) {
	authorized, user := isAuthorised(r, h)
	if !user.IsAdmin {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	idString := r.PathValue("articleID")

	if idString == "" || idString == "1" {
		layouts.PublishingPage(authorized, user, &models.Article{ID: 0}).Render(r.Context(), w)
		return
	}

	articleID, err := strconv.Atoi(idString)
	if err != nil || articleID < 1 {
		http.NotFound(w, r)
		return
	}

	article, err := h.articleRepo.GetByID(articleID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	layouts.PublishingPage(authorized, user, article).Render(r.Context(), w)
}

func (h *BaseHandler) publishingFormHandler(w http.ResponseWriter, r *http.Request) {
	authorized, user := isAuthorised(r, h)
	if !authorized || !user.IsAdmin {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	a := models.Article{ID: 0}

	idString := r.PathValue("articleID")
	a.ID, _ = strconv.Atoi(idString)

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Print(err)
		http.Error(w, "Невозможно обработать данные формы", http.StatusBadRequest)
		return
	}

	a.Slug = r.PostFormValue("slug")
	a.Title = r.PostFormValue("title")
	a.Description = r.PostFormValue("description")
	a.TextMD = r.PostFormValue("textMD")

	re := regexp.MustCompile(`^[a-z0-9-]+$`)
	match := re.MatchString(a.Slug)
	if !match {
		slugResult := components.FormWarning("Ссылка может содержать только маленькие латинские буквы, цифры и знак \"-\"")
		components.PublishingForm(slugResult, templ.NopComponent, &a).Render(r.Context(), w)
		return
	}

	art, err := h.articleRepo.GetBySlug(a.Slug)
	if err == nil && art.ID != a.ID {
		slugResult := components.FormWarning("Эта ссылка уже занята")
		components.PublishingForm(slugResult, templ.NopComponent, &a).Render(r.Context(), w)
		return
	}

	// Обработка файла обложки
	file, fileHeader, err := r.FormFile("coverImage")
	if err != nil {
		if err != http.ErrMissingFile {
			http.Error(w, "Ошибка при чтении картинки обложки", http.StatusInternalServerError)
			return
		}
		file = nil
	}
	defer func() {
		if file != nil {
			file.Close()
		}
	}()

	var coverImageID int
	if file != nil {
		if fileHeader.Size > 1<<20 {
			coverResult := components.FormWarning("Файл слишком большой. Максимальный размер: 1МБ.")
			components.PublishingForm(templ.NopComponent, coverResult, &a).Render(r.Context(), w)
			return
		}
		content, err := io.ReadAll(file)
		if err != nil {
			coverResult := components.FormWarning("Ошибка при чтении файла картинки обложки")
			components.PublishingForm(templ.NopComponent, coverResult, &a).Render(r.Context(), w)
			return
		}

		coverImageID, err = h.imageRepo.Create(fileHeader.Filename, user.ID, content)
		if err != nil {
			log.Print(err)
			coverResult := components.FormWarning("Ошибка при сохранении файла картинки обложки в базу данных")
			components.PublishingForm(templ.NopComponent, coverResult, &a).Render(r.Context(), w)
			return
		}
	} else {
		coverImageID = 0
	}
	a.CoverImageID = coverImageID

	if a.ID == 0 {
		err = h.articleRepo.Create(a.Slug, a.Title, a.TextMD, a.Description, coverImageID)
	} else {
		err = h.articleRepo.Update(&a)
	}

	if err != nil {
		log.Print(err)
		slugResult := components.FormWarning("Внутренняя ошибка сервера")
		components.PublishingForm(slugResult, templ.NopComponent, &a).Render(r.Context(), w)
		return
	}

	components.PublishingSuccessful(a.Slug).Render(r.Context(), w)
}
