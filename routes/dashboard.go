package routes

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/svuvi/theweek/components"
	"github.com/svuvi/theweek/layouts"
)

func (h *BaseHandler) DasboardPageHandler(w http.ResponseWriter, r *http.Request) {
	authorized, user := isAuthorised(r, h)
	if !authorized || !user.IsAdmin {
		http.Error(w, "Отказано в доступе", http.StatusUnauthorized)
		return
	}

	layouts.DashboardHome().Render(r.Context(), w)
}

func (h *BaseHandler) DashboardUsersHandler(w http.ResponseWriter, r *http.Request) {
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

func (h *BaseHandler) DashboardInvitesHandler(w http.ResponseWriter, r *http.Request) {
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

func (h *BaseHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
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

func (h *BaseHandler) DeleteInvite(w http.ResponseWriter, r *http.Request) {
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
