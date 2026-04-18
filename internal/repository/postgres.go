package repository

// Содержит интерфейс для выполнения SQL с postgresql,
// обеспечивая сохранение и получение информации о пользователях, списках желаний и подарках.

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"

	"cdek/internal/domain"

	_ "github.com/lib/pq"
)

type PostgresRepository interface {
	CreateUser(ctx context.Context, user domain.User) (int, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)

	CreateWishlist(ctx context.Context, w domain.Wishlist) (int, string, error)
	GetWishlistsByUserID(ctx context.Context, userID int) ([]domain.Wishlist, error)
	GetWishlistByID(ctx context.Context, id int) (domain.Wishlist, error)
	GetWishlistByToken(ctx context.Context, token string) (domain.Wishlist, error)
	UpdateWishlist(ctx context.Context, w domain.Wishlist) error
	DeleteWishlist(ctx context.Context, id int) error

	CreateItem(ctx context.Context, item domain.Item) (int, error)
	GetItemByID(ctx context.Context, id int) (domain.Item, error)
	GetItemsByWishlistID(ctx context.Context, wishlistID int) ([]domain.Item, error)
	UpdateItem(ctx context.Context, item domain.Item) error
	DeleteItem(ctx context.Context, id int) error
	ReserveItem(ctx context.Context, token string, itemID int) error
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) PostgresRepository {
	return &postgresRepository{db: db}
}

func generateToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (r *postgresRepository) CreateUser(ctx context.Context, user domain.User) (int, error) {
	var id int

	query := `
		INSERT INTO users (email, password) 
		VALUES ($1, $2) RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query, user.Email, user.Password).Scan(&id)

	return id, err
}

func (r *postgresRepository) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	var user domain.User

	query := `
		SELECT id, email, password, created_at 
		FROM users 
		WHERE email = $1
	`

	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt)

	return user, err
}

func (r *postgresRepository) CreateWishlist(ctx context.Context, w domain.Wishlist) (int, string, error) {
	var id int

	token := generateToken()

	query := `
		INSERT INTO wishlists (user_id, title, description, event_date, token) 
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query, w.UserID, w.Title, w.Description, w.EventDate, token).Scan(&id)

	return id, token, err
}

func (r *postgresRepository) GetWishlistsByUserID(ctx context.Context, userID int) ([]domain.Wishlist, error) {
	query := `
		SELECT id, user_id, title, description, event_date, token 
		FROM wishlists
		WHERE user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var wishlists []domain.Wishlist

	for rows.Next() {
		var w domain.Wishlist

		if err := rows.Scan(&w.ID, &w.UserID, &w.Title, &w.Description, &w.EventDate, &w.Token); err != nil {
			return nil, err
		}

		wishlists = append(wishlists, w)
	}

	return wishlists, nil
}

func (r *postgresRepository) GetWishlistByToken(ctx context.Context, token string) (domain.Wishlist, error) {
	var w domain.Wishlist

	query := `
		SELECT id, user_id, title, description, event_date, token 
		FROM wishlists 
		WHERE token = $1
	`

	err := r.db.QueryRowContext(ctx, query, token).Scan(&w.ID, &w.UserID, &w.Title, &w.Description, &w.EventDate, &w.Token)

	return w, err
}

func (r *postgresRepository) CreateItem(ctx context.Context, item domain.Item) (int, error) {
	var id int

	query := `
		INSERT INTO items (wishlist_id, title, description, url, priority) 
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query, item.WishlistID, item.Title, item.Description, item.URL, item.Priority).Scan(&id)

	return id, err
}

func (r *postgresRepository) GetItemsByWishlistID(ctx context.Context, wishlistID int) ([]domain.Item, error) {
	query := `
		SELECT id, wishlist_id, title, description, url, priority, is_reserved 
		FROM items 
		WHERE wishlist_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, wishlistID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var items []domain.Item

	for rows.Next() {
		var i domain.Item

		if err := rows.Scan(&i.ID, &i.WishlistID, &i.Title, &i.Description, &i.URL, &i.Priority, &i.IsReserved); err != nil {
			return nil, err
		}

		items = append(items, i)
	}

	return items, nil
}

func (r *postgresRepository) ReserveItem(ctx context.Context, token string, itemID int) error {
	query := `
		UPDATE items 
		SET is_reserved = true 
		WHERE id = $1 
			AND is_reserved = false 
			AND wishlist_id = (SELECT id 
			                   FROM wishlists 
			                   WHERE token = $2)
	`

	res, err := r.db.ExecContext(ctx, query, itemID, token)

	if err != nil {
		return err
	}

	rowsAffected, _ := res.RowsAffected()

	if rowsAffected == 0 {
		return domain.ErrUnchanged
	}

	return nil
}

func (r *postgresRepository) UpdateWishlist(ctx context.Context, w domain.Wishlist) error {
	query := `
		UPDATE wishlists 
		SET title = $1, description = $2, event_date = $3 
		WHERE id = $4
	`

	_, err := r.db.ExecContext(ctx, query, w.Title, w.Description, w.EventDate, w.ID)

	return err
}

func (r *postgresRepository) DeleteWishlist(ctx context.Context, id int) error {
	query := `
		DELETE FROM wishlists 
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)

	return err
}

func (r *postgresRepository) GetWishlistByID(ctx context.Context, id int) (domain.Wishlist, error) {
	var w domain.Wishlist

	query := `
		SELECT id, user_id, title, description, event_date, token 
		FROM wishlists 
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(&w.ID, &w.UserID, &w.Title, &w.Description, &w.EventDate, &w.Token)

	return w, err
}

func (r *postgresRepository) GetItemByID(ctx context.Context, id int) (domain.Item, error) {
	var i domain.Item

	query := `
		SELECT id, wishlist_id, title, description, url, priority, is_reserved 
		FROM items 
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(&i.ID, &i.WishlistID, &i.Title, &i.Description, &i.URL, &i.Priority, &i.IsReserved)

	return i, err
}

func (r *postgresRepository) UpdateItem(ctx context.Context, item domain.Item) error {
	query := `
		UPDATE items 
		SET title = $1, description = $2, url = $3, priority = $4 
		WHERE id = $5
	`

	_, err := r.db.ExecContext(ctx, query, item.Title, item.Description, item.URL, item.Priority, item.ID)

	return err
}

func (r *postgresRepository) DeleteItem(ctx context.Context, id int) error {
	query := `
		DELETE FROM items
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)

	return err
}
