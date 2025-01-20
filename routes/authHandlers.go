package routes

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/svuvi/theweek/components"
	"github.com/svuvi/theweek/layouts"
	"golang.org/x/crypto/bcrypt"
)

func (h *BaseHandler) loginFormHandler(w http.ResponseWriter, r *http.Request) {
	authorized, _ := isAuthorised(r, h)
	if authorized {
		http.Error(w, "Вы уже зашли в аккаунт", http.StatusBadRequest)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
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

	if !acceptablePassword(password) {
		passwordResult = components.FormWarning("Неприемлимый пароль. Пароль не должен быть короче 6 символов или длиннее 72")
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
		usernameResult = components.FormWarning("Такой пользователь не существует")
		components.LoginForm(username, password, usernameResult, passwordResult).Render(r.Context(), w)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassowrd), []byte(password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		passwordResult = components.FormWarning("Неверный пароль")
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

	err = r.ParseForm()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
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

	if !acceptablePassword(password) {
		passwordResult = components.FormWarning("Неприемлимый пароль. Пароль не должен быть короче 6 символов или длиннее 72")
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

func (h *BaseHandler) changePasswordForm(w http.ResponseWriter, r *http.Request) {
	authorised, user := isAuthorised(r, h)
	if !authorised {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	passwordCurrent := r.PostFormValue("passwordCurrent")
	passwordNew := r.PostFormValue("passwordNew")
	passwordNewRepeat := r.PostFormValue("passwordNewRepeat")

	passwordResult := components.Empty()
	passwordNewResult := components.Empty()
	repeatResult := components.Empty()
	earlyReturn := false

	if !acceptablePassword(passwordCurrent) {
		passwordResult = components.FormWarning("Неприемлимый пароль. Пароль не должен быть короче 6 символов или длиннее 72")
		earlyReturn = true
	} else {
		if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassowrd), []byte(passwordCurrent)); err != nil {
			passwordResult = components.FormWarning("Неверный пароль")
			earlyReturn = true
		}
	}
	if !acceptablePassword(passwordNew) {
		passwordNewResult = components.FormWarning("Неприемлимый пароль. Пароль не должен быть короче 6 символов или длиннее 72")
		earlyReturn = true
	}
	if passwordNew != passwordNewRepeat {
		repeatResult = components.FormWarning("Повтор не совпадает")
		earlyReturn = true
	}

	if earlyReturn {
		components.PasswordChangeForm(passwordResult, passwordNewResult, repeatResult, passwordCurrent, passwordNew, "").Render(r.Context(), w)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordNew), 14)
	if err != nil {
		repeatResult = components.FormWarning("Внутрянняя ошибка сервера. Новый пароль не вступает в силу")
		components.PasswordChangeForm(passwordResult, passwordNewResult, repeatResult, passwordCurrent, passwordNew, "").Render(r.Context(), w)
		return
	}

	err = h.userRepo.ChangePassword(user.ID, string(hashedPassword))
	if err != nil {
		repeatResult = components.FormWarning("Внутрянняя ошибка сервера. Новый пароль не вступает в силу")
		components.PasswordChangeForm(passwordResult, passwordNewResult, repeatResult, passwordCurrent, passwordNew, "").Render(r.Context(), w)
		return
	}

	components.PasswordChanged().Render(r.Context(), w)
	log.Printf("Изменен пароль пользователя %s", user.Username)
}

func (h *BaseHandler) restorePasswordPage(w http.ResponseWriter, r *http.Request) {
	isAuthorised(r, h) // Здесь только для отметки последнего использования сессии

	q := r.URL.Query()
	code := q.Get("code")

	if code == "" {
		layouts.RestorePasswordRequestPage(false).Render(r.Context(), w)
		return
	}

	if err := uuid.Validate(code); err != nil {
		layouts.RestorePasswordRequestPage(true).Render(r.Context(), w)
		return
	}

	rCode, err := h.recoveryCodeRepo.Get(code)
	if err != nil || !rCode.IsActive() {
		layouts.RestorePasswordRequestPage(true).Render(r.Context(), w)
		return
	}

	layouts.RestorePasswordPage(code).Render(r.Context(), w)
}

func (h *BaseHandler) restorePasswordForm(w http.ResponseWriter, r *http.Request) {
	authorised, _ := isAuthorised(r, h) // отметка

	q := r.URL.Query()
	code := q.Get("code")

	if err := uuid.Validate(code); err != nil {
		log.Print("invalid uuid")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	rCode, err := h.recoveryCodeRepo.Get(code)
	if err != nil || !rCode.IsActive() {
		log.Print("err when retreiving the code from db: \n", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	userame := r.PostFormValue("username")
	password := r.PostFormValue("password")
	passwordRepeat := r.PostFormValue("passwordRepeat")

	if password != passwordRepeat {
		result := components.FormWarning("Пароли не совпадают")
		components.PasswordRestoreForm(code, result).Render(r.Context(), w)
		return
	}

	if !acceptablePassword(password) {
		result := components.FormWarning("Неприемлимый пароль. Пароль не должен быть короче 6 символов или длиннее 72")
		components.PasswordRestoreForm(code, result).Render(r.Context(), w)
		return
	}

	usrFromForm, err := h.userRepo.GetByUsername(userame)
	if err == sql.ErrNoRows {
		result := components.FormWarning("Неверный логин")
		components.PasswordRestoreForm(code, result).Render(r.Context(), w)
		log.Print(err)
		return
	} else if err != nil {
		result := components.FormWarning("Внутренняя ошибка сервера. Сообщи администратору и попробуй позже")
		components.PasswordRestoreForm(code, result).Render(r.Context(), w)
		log.Print(err)
		return
	}

	if usrFromForm.ID != rCode.UserID {
		result := components.FormWarning("Неверное имя пользователя или код")
		components.PasswordRestoreForm(code, result).Render(r.Context(), w)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		result := components.FormWarning("Внутренняя ошибка сервера. Сообщи администратору и попробуй позже")
		components.PasswordRestoreForm(code, result).Render(r.Context(), w)
		return
	}

	err = h.userRepo.ChangePassword(rCode.UserID, string(hash))
	if err != nil {
		result := components.FormWarning("Внутренняя ошибка сервера. Сообщи администратору и попробуй позже")
		components.PasswordRestoreForm(code, result).Render(r.Context(), w)
		return
	}

	err = h.recoveryCodeRepo.SetUsed(rCode.RecoveryCode)
	if err != nil {
		log.Printf("Ошибка при попытке отметить код восстановления для пользователя как использованный.\nКод:%s\nID пользователя:%d", rCode.RecoveryCode, rCode.UserID)
	}

	if !authorised {
		sessionKey := uuid.NewString()
		_, err = h.sessionRepo.Create(rCode.ID, sessionKey)
		if err != nil {
			result := components.FormWarning("Внутренняя ошибка сервера. Сообщи администратору и попробуй позже")
			components.PasswordRestoreForm(code, result).Render(r.Context(), w)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_key",
			Value:    sessionKey,
			Path:     "",
			HttpOnly: true,
			Secure:   true,
		})
	}

	components.PasswordRestored().Render(r.Context(), w)
}
