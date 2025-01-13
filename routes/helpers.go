package routes

import (
	"log"
	"net/http"

	"github.com/google/uuid"
)

// Также обновляет last_use
func isAuthorised(r *http.Request, h *BaseHandler) bool {
	value, err := getSessionKey(r)
	if err != nil {
		return false
	}

	session, err := h.sessionRepo.GetSessionByKey(value)
	if err != nil || !session.IsActive {
		return false
	}

	if err = h.sessionRepo.UpdateLastUsedByID(session.ID); err != nil {
		log.Printf("Ошибка при попытке обновить поле last_use для сессии с ID=%d:\n%v", session.ID, err)
	}

	return true
}

// getSessionKey читает и валидирует "session-key" куки из запроса.
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
