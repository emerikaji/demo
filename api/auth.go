package api

import (
	"demo/util"
	"errors"
	"net/http"
	"strings"

	"demo/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type registerUserRequest struct {
	Email    string          `json:"email"`
	Password string          `json:"password"`
	Role     domain.UserRole `json:"role"`
}

type loginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) handleRegisterUser(w http.ResponseWriter, r *http.Request) {
	var req registerUserRequest
	if err := util.DecodeJSON(w, r, &req); err != nil {
		util.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Password == "" || req.Role == "" {
		util.RespondError(w, http.StatusBadRequest, "email, password, and role are required")
		return
	}

	if req.Role != domain.RoleAttendee && req.Role != domain.RoleOrganizer {
		util.RespondError(w, http.StatusBadRequest, "invalid role: must be 'organizer' or 'attendee'")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, "failed to process password")
		return
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         req.Role,
	}

	if err := h.userStore.CreateUser(r.Context(), user); err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			util.RespondError(w, http.StatusConflict, "user with this email already exists")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	_ = util.EncodeJSON(w, http.StatusCreated, util.Map{
		"message": "user registered successfully",
	})
}

func (h *Handler) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	var req loginUserRequest
	if err := util.DecodeJSON(w, r, &req); err != nil {
		util.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Password == "" {
		util.RespondError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, err := h.userStore.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			util.RespondError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "failed to fetch user")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		util.RespondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := h.tokens.Generate(user)
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, "failed to generate access token")
		return
	}

	_ = util.EncodeJSON(w, http.StatusOK, util.Map{
		"token": token,
	})
}
