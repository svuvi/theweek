package routes

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/svuvi/theweek/components"
	"golang.org/x/crypto/bcrypt"
)

func (h *BaseHandler) loginFormHandler(w http.ResponseWriter, r *http.Request) {
	authorized, _ := isAuthorised(r, h)
	if authorized {
		http.Error(w, "Вы уже зашли в аккаунт", http.StatusBadRequest)
		return
	}

	r.ParseForm()
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	usernameResult := components.Empty()
	passwordResult := components.Empty()
	earlyReturn := false

	if username == "" {
		usernameResult = components.FormWarning("Логин не может быть пустым")
		earlyReturn = true
	} else if len(username) > 30 {
		usernameResult = components.FormWarning("Логин слишком длинный")
		earlyReturn = true
	} else if len(username) < 2 {
		usernameResult = components.FormWarning("Логин слишком короткий")
		earlyReturn = true
	}

	if password == "" {
		passwordResult = components.FormWarning("Пароль не может быть пустым")
		earlyReturn = true
	} else if len(password) > 72 {
		passwordResult = components.FormWarning("Пароль слишком длинный")
		earlyReturn = true
	} else if len(password) < 6 {
		passwordResult = components.FormWarning("Пароль слишком короткий")
		earlyReturn = true
	}

	if earlyReturn {
		w.WriteHeader(http.StatusUnauthorized)
		components.LoginForm(username, password, usernameResult, passwordResult).Render(r.Context(), w)
		return
	}

	user, err := h.userRepo.GetByUsername(username)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		passwordResult = components.FormWarning("Не авторизован")
		components.LoginForm(username, password, usernameResult, passwordResult).Render(r.Context(), w)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassowrd), []byte(password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		passwordResult = components.FormWarning("Не авторизован")
		components.LoginForm(username, password, usernameResult, passwordResult).Render(r.Context(), w)
		return
	}

	sessionKey := uuid.NewString()
	_, err = h.sessionRepo.Create(user.ID, sessionKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		passwordResult = components.FormWarning("Ошкбка на стороне сервера")
		components.LoginForm(username, password, usernameResult, passwordResult).Render(r.Context(), w)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_key",
		Value:    sessionKey,
		Path:     "",
		HttpOnly: true,
		Secure:   true,
	})
	components.LoggedIn().Render(r.Context(), w)
}

func (h *BaseHandler) registrationFormHandler(w http.ResponseWriter, r *http.Request) {
	authorized, _ := isAuthorised(r, h)
	if authorized {
		http.Error(w, "Вы уже зашли в аккаунт", http.StatusBadRequest)
		return
	}

	code, err := getInviteCode(r)
	if err != nil {
		http.Error(w, "Невалидный код приглашения", http.StatusBadRequest)
		return
	}

	invite, err := h.inviteRepo.GetByCode(code)
	if err != nil || !invite.IsActive {
		http.Error(w, "Невалидный код приглашения", http.StatusBadRequest)
		return
	}

	r.ParseForm()
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	passwordRepeat := r.PostFormValue("passwordRepeat")

	usernameResult := components.Empty()
	passwordResult := components.Empty()
	passwordRepeatResult := components.Empty()
	earlyReturn := false

	if username == "" {
		usernameResult = components.FormWarning("Логин не может быть пустым")
		earlyReturn = true
	} else if len(username) > 30 {
		usernameResult = components.FormWarning("Логин слишком длинный")
		earlyReturn = true
	} else if len(username) < 2 {
		usernameResult = components.FormWarning("Логин слишком короткий")
		earlyReturn = true
	} else if _, err := h.userRepo.GetByUsername(username); err != sql.ErrNoRows {
		usernameResult = components.FormWarning("Логин уже занят")
		earlyReturn = true
	}

	if password == "" {
		passwordResult = components.FormWarning("Пароль не может быть пустым")
		earlyReturn = true
	} else if len([]byte(password)) > 72 {
		passwordResult = components.FormWarning("Пароль слишком длинный")
		earlyReturn = true
	} else if len(password) < 6 {
		passwordResult = components.FormWarning("Пароль слишком короткий")
		earlyReturn = true
	}

	if passwordRepeat != password {
		passwordRepeatResult = components.FormWarning("Пароль не совпадает")
		earlyReturn = true
	}

	if earlyReturn {
		w.WriteHeader(http.StatusUnauthorized)
		components.RegistrationForm(username, password, passwordRepeat, usernameResult, passwordResult, passwordRepeatResult).Render(r.Context(), w)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Print("Unexpected error when generating hash from password: ", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	user, err := h.userRepo.Create(username, string(hashedPassword))
	if err != nil {
		log.Printf("Error when creating user.\nUsername: %s, Hashed password: %s\n%v", username, hashedPassword, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	sessionKey := uuid.NewString()
	_, err = h.sessionRepo.Create(user.ID, sessionKey)
	if err != nil {
		log.Printf("Error when creating session for the user who just registered.\nUsername: %s, Hashed password: %s\n%v", username, hashedPassword, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_key",
		Value:    sessionKey,
		Path:     "",
		HttpOnly: true,
		Secure:   true,
	})

	components.Registered().Render(r.Context(), w)

	if err = h.inviteRepo.Claim(code, user.ID); err != nil {
		log.Printf("Ошибка при попытке отметить приглашение как использованное.\n Код: %s\nОшибка%v", code, err)
	}
}

func (h *BaseHandler) logoutHandler(w http.ResponseWriter, r *http.Request) {
	authorized, _ := isAuthorised(r, h)
	if !authorized {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	sessionKey, _ := getSessionKey(r)
	err := h.sessionRepo.SetInactive(sessionKey)
	if err != nil {
		log.Printf("Error when setting session \"%s\" inactive: %v", sessionKey, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_key",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
	})

	http.Redirect(w, r, "/", http.StatusFound)
}
