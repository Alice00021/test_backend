package entity

import "errors"

var (
	ErrAccessDenied = errors.New("access denied")
	ErrUserNotFound = errors.New("user not found")

	ErrEmailNotVerified    = errors.New("email not verified")
	ErrEmailAlreadyUsed    = errors.New("email already used")
	ErrGenerateVerifyToken = errors.New("failed to generate verify token")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidToken        = errors.New("invalid token")
	ErrExpiredToken        = errors.New("token expired")

	ErrAuthorNotFound = errors.New("author not found")
	ErrBookNotFound   = errors.New("book not found")

	ErrPasswordMismatch = errors.New("newPassword and confirmPassword must be the same")

	ErrOpenFile   = errors.New("failed to open file")
	ErrCreateFile = errors.New("failed to create file")
	ErrSaveFile   = errors.New("failed to save file")

	ErrCommandNotFound         = errors.New("command not found")
	ErrCommandDuplicateAddress = errors.New("address is used by multiple commands")
	ErrCommandVolumeExceeded   = errors.New("volume exceeded")
	ErrCommandNameNotFound     = errors.New("command name not found")
	ErrOperationNotFound       = errors.New("operation not found")
)
