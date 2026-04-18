package handler

// Содержит маршрутизацию, аутентификацию и обработчики
// для взаимодействия с API вишлистов и пользователей.

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"cdek/internal/domain"
	"cdek/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
)

type Handler struct {
	svc       *service.Service
	jwtSecret string
}

func NewHandler(svc *service.Service, secret string) *Handler {
	return &Handler{svc: svc, jwtSecret: secret}
}

func (h *Handler) InitRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Post("/register", h.register)
		r.Post("/login", h.login)

		r.Get("/wishlists/public/{token}", h.getPublicWishlist)
		r.Post("/wishlists/public/{token}/items/{itemID}/reserve", h.reserveItem)

		r.Group(func(r chi.Router) {
			r.Use(h.authMiddleware)
			r.Post("/wishlists", h.createWishlist)
			r.Get("/wishlists", h.getUserWishlists)
			r.Put("/wishlists/{wishlistID}", h.updateWishlist)
			r.Delete("/wishlists/{wishlistID}", h.deleteWishlist)

			r.Post("/wishlists/{wishlistID}/items", h.addItem)
			r.Put("/items/{itemID}", h.updateItem)
			r.Delete("/items/{itemID}", h.deleteItem)
		})
	})

	fs := http.FileServer(http.Dir("./static"))

	r.Handle("/*", http.StripPrefix("/", fs))

	return r
}

func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "missing auth header", http.StatusUnauthorized)

			return
		}

		parts := strings.Split(authHeader, " ")

		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid auth header", http.StatusUnauthorized)

			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			return []byte(h.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)

			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok {
			http.Error(w, "invalid claims", http.StatusUnauthorized)

			return
		}

		userID := int(claims["user_id"].(float64))

		ctx := context.WithValue(r.Context(), "userID", userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch err {
	case domain.ErrForbidden:
		http.Error(w, err.Error(), http.StatusForbidden)
	case domain.ErrNotFound:
		http.Error(w, err.Error(), http.StatusNotFound)
	case domain.ErrAlreadyReserved:
		http.Error(w, err.Error(), http.StatusConflict)
	case domain.ErrInvalidCredentials:
		http.Error(w, err.Error(), http.StatusUnauthorized)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)

		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)

		return
	}

	if err := h.svc.Register(r.Context(), req); err != nil {
		h.handleError(w, err)

		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"message": "user created"})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)

		return
	}

	token, err := h.svc.Login(r.Context(), req)

	if err != nil {
		h.handleError(w, err)

		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *Handler) createWishlist(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)

	var req domain.CreateWishlistReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)

		return
	}

	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)

		return
	}

	wishlist, err := h.svc.CreateWishlist(r.Context(), userID, req)

	if err != nil {
		h.handleError(w, err)

		return
	}

	respondJSON(w, http.StatusCreated, wishlist)
}

func (h *Handler) getUserWishlists(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	lists, err := h.svc.GetUserWishlists(r.Context(), userID)

	if err != nil {
		h.handleError(w, err)

		return
	}

	respondJSON(w, http.StatusOK, lists)
}

func (h *Handler) addItem(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	wishlistID, _ := strconv.Atoi(chi.URLParam(r, "wishlistID"))

	var req domain.CreateItemReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)

		return
	}

	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)

		return
	}

	item, err := h.svc.AddItem(r.Context(), userID, wishlistID, req)

	if err != nil {
		h.handleError(w, err)

		return
	}

	respondJSON(w, http.StatusCreated, item)
}

func (h *Handler) getPublicWishlist(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	wishlist, err := h.svc.GetPublicWishlist(r.Context(), token)

	if err != nil {
		h.handleError(w, err)

		return
	}

	respondJSON(w, http.StatusOK, wishlist)
}

func (h *Handler) reserveItem(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	itemID, _ := strconv.Atoi(chi.URLParam(r, "itemID"))

	err := h.svc.ReserveItem(r.Context(), token, itemID)

	if err != nil {
		h.handleError(w, err)

		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "item reserved successfully"})
}

func (h *Handler) updateWishlist(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	wishlistID, _ := strconv.Atoi(chi.URLParam(r, "wishlistID"))

	var req domain.CreateWishlistReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)

		return
	}

	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)

		return
	}

	if err := h.svc.UpdateWishlist(r.Context(), userID, wishlistID, req); err != nil {
		h.handleError(w, err)

		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "wishlist updated"})
}

func (h *Handler) deleteWishlist(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	wishlistID, _ := strconv.Atoi(chi.URLParam(r, "wishlistID"))

	if err := h.svc.DeleteWishlist(r.Context(), userID, wishlistID); err != nil {
		h.handleError(w, err)

		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "wishlist deleted"})
}

func (h *Handler) updateItem(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	itemID, _ := strconv.Atoi(chi.URLParam(r, "itemID"))

	var req domain.CreateItemReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)

		return
	}

	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)

		return
	}

	if err := h.svc.UpdateItem(r.Context(), userID, itemID, req); err != nil {
		h.handleError(w, err)

		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "item updated"})
}

func (h *Handler) deleteItem(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	itemID, _ := strconv.Atoi(chi.URLParam(r, "itemID"))

	if err := h.svc.DeleteItem(r.Context(), userID, itemID); err != nil {
		h.handleError(w, err)

		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "item deleted"})
}
