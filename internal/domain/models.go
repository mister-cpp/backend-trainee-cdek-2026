package domain

// Определяем основные структуры данных, DTO для запросов и ответов API,
// а также общие ошибки.

import (
	"errors"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type Wishlist struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	EventDate   time.Time `json:"event_date"`
	Token       string    `json:"token"`
	Items       []Item    `json:"items,omitempty"`
}

type Item struct {
	ID          int    `json:"id"`
	WishlistID  int    `json:"wishlist_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Priority    int    `json:"priority"`
	IsReserved  bool   `json:"is_reserved"`
}

type RegisterReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateWishlistReq struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	EventDate   time.Time `json:"event_date"`
}

type CreateItemReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Priority    int    `json:"priority"`
}

var (
	ErrForbidden          = errors.New("access denied")
	ErrUnchanged          = errors.New("item not found in this wishlist or already reserved")
	ErrAlreadyReserved    = errors.New("cannot modify reserved item")
	ErrNotFound           = errors.New("resource not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidData        = errors.New("invalid data")
)
