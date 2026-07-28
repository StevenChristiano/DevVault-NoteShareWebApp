package usecase

import (
	"errors"
	"time"
	"unicode"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/repository"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/pkg/token"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyUsed = errors.New("email is already registered")
	ErrInvalidCredentials = errors.New("email or password is incorrect")
	ErrPasswordTooWeak = errors.New("password is too weak, must be at least 8 characters, contain at least an uppercase letter, a lowercase letter, and a number")
)

type AuthUsecase interface {
	Register(username, email, password string) error
	Login(email, password string) (jwtToken string, err error)
}

type authUsecase struct {
	userRepo  repository.UserRepository
	jwtSecret string
	jwtTTL    time.Duration
}

func NewAuthUsecase(userRepo repository.UserRepository, jwtSecret string, jwtTTL time.Duration) AuthUsecase {
	return &authUsecase{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtTTL:    jwtTTL,
	}
}

func isPasswordStrong(password string) bool {
    if len(password) < 8 {
        return false
    }
    var hasUpper, hasLower, hasNumber bool
    for _, ch := range password {
        switch {
        case unicode.IsUpper(ch):
            hasUpper = true
        case unicode.IsLower(ch):
            hasLower = true
        case unicode.IsNumber(ch):
            hasNumber = true
        }
    }
    return hasUpper && hasLower && hasNumber
}

func (u *authUsecase) Register(username, email, password string) error {
	existing, err := u.userRepo.FindByEmail(email)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrEmailAlreadyUsed
	}
	if !isPasswordStrong(password) {
		return ErrPasswordTooWeak
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	user := &entity.User {
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}
	return u.userRepo.Create(user)
}

func (u *authUsecase) Login(email, password string) (jwtToken string, err error) {
	user, err := u.userRepo.FindByEmail(email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	return token.Generate(user.ID, u.jwtSecret, u.jwtTTL)
}