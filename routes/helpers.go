package routes

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/svuvi/theweek/models"
)

// Также обновляет last_use
func isAuthorised(r *http.Request, h *BaseHandler) (bool, *models.User) {
	user := new(models.User)
	value, err := getSessionKey(r)
	if err != nil {
		return false, user
	}

	session, err := h.sessionRepo.GetSessionByKey(value)
	if err != nil || !session.IsActive {
		return false, user
	}

	user, err = h.userRepo.GetByID(session.UserID)
	if err != nil {
		// невозможный сценарий
		return false, user
	}

	if err = h.sessionRepo.UpdateLastUsedByID(session.ID); err != nil {
		log.Printf("Ошибка при попытке обновить поле last_use для сессии с ID=%d:\n%v", session.ID, err)
	}

	return true, user
}

// getSessionKey читает и валидирует "session_key" куки из запроса.
// Возвращает куки если это валидная uuid-строка и nil. Иначе, возвращает пустую строку и ошибку.
func getSessionKey(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_key")
	if err == http.ErrNoCookie {
		return "", err
	}
	if err = uuid.Validate(cookie.Value); err != nil {
		return "", err
	}

	return cookie.Value, nil
}

// getInviteCode читает и валидирует "registration_invite" куки из запроса.
// Возвращает куки если это валидная uuid-строка и nil. Иначе, возвращает пустую строку и ошибку.
func getInviteCode(r *http.Request) (string, error) {
	cookie, err := r.Cookie("registration_invite")
	if err == http.ErrNoCookie {
		return "", err
	}
	if err = uuid.Validate(cookie.Value); err != nil {
		return "", err
	}

	return cookie.Value, nil
}
