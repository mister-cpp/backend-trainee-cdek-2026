package service

// Реализованы методы для создания, редактирования, удаления вишлистов
// и управления подарками (добавление, обновление, резервирование).

import (
	"cdek/internal/domain"
	"context"
)

func (s *Service) CreateWishlist(ctx context.Context, userID int, req domain.CreateWishlistReq) (domain.Wishlist, error) {
	w := domain.Wishlist{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		EventDate:   req.EventDate,
	}

	id, token, err := s.repo.CreateWishlist(ctx, w)

	if err != nil {
		return w, err
	}

	w.ID = id
	w.Token = token

	return w, nil
}

func (s *Service) GetUserWishlists(ctx context.Context, userID int) ([]domain.Wishlist, error) {
	return s.repo.GetWishlistsByUserID(ctx, userID)
}

func (s *Service) AddItem(ctx context.Context, userID int, wishlistID int, req domain.CreateItemReq) (domain.Item, error) {
	if req.Priority < 1 || req.Priority > 10 {
		return domain.Item{}, domain.ErrInvalidData
	}

	wl, err := s.repo.GetWishlistByID(ctx, wishlistID)

	if err != nil {
		return domain.Item{}, domain.ErrNotFound
	}

	if wl.UserID != userID {
		return domain.Item{}, domain.ErrForbidden
	}

	item := domain.Item{
		WishlistID:  wishlistID,
		Title:       req.Title,
		Description: req.Description,
		URL:         req.URL,
		Priority:    req.Priority,
	}

	id, err := s.repo.CreateItem(ctx, item)

	item.ID = id

	return item, err
}

func (s *Service) GetPublicWishlist(ctx context.Context, token string) (domain.Wishlist, error) {
	w, err := s.repo.GetWishlistByToken(ctx, token)

	if err != nil {
		return w, domain.ErrNotFound
	}

	items, err := s.repo.GetItemsByWishlistID(ctx, w.ID)

	if err != nil {
		return w, err
	}

	w.Items = items

	return w, nil
}

func (s *Service) ReserveItem(ctx context.Context, token string, itemID int) error {
	return s.repo.ReserveItem(ctx, token, itemID)
}

func (s *Service) UpdateWishlist(ctx context.Context, userID int, id int, req domain.CreateWishlistReq) error {
	wl, err := s.repo.GetWishlistByID(ctx, id)

	if err != nil {
		return domain.ErrNotFound
	}

	if wl.UserID != userID {
		return domain.ErrForbidden
	}

	w := domain.Wishlist{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		EventDate:   req.EventDate,
	}

	return s.repo.UpdateWishlist(ctx, w)
}

func (s *Service) DeleteWishlist(ctx context.Context, userID int, id int) error {
	wl, err := s.repo.GetWishlistByID(ctx, id)

	if err != nil {
		return domain.ErrNotFound
	}

	if wl.UserID != userID {
		return domain.ErrForbidden
	}

	return s.repo.DeleteWishlist(ctx, id)
}

func (s *Service) UpdateItem(ctx context.Context, userID int, id int, req domain.CreateItemReq) error {
	if req.Priority < 1 || req.Priority > 10 {
		return domain.ErrInvalidData
	}

	itemOld, err := s.repo.GetItemByID(ctx, id)

	if err != nil {
		return domain.ErrNotFound
	}

	if itemOld.IsReserved {
		return domain.ErrAlreadyReserved
	}

	wl, err := s.repo.GetWishlistByID(ctx, itemOld.WishlistID)

	if err != nil {
		return err
	}

	if wl.UserID != userID {
		return domain.ErrForbidden
	}

	item := domain.Item{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		URL:         req.URL,
		Priority:    req.Priority,
	}

	return s.repo.UpdateItem(ctx, item)
}

func (s *Service) DeleteItem(ctx context.Context, userID int, id int) error {
	item, err := s.repo.GetItemByID(ctx, id)

	if err != nil {
		return domain.ErrNotFound
	}

	wl, err := s.repo.GetWishlistByID(ctx, item.WishlistID)

	if err != nil {
		return err
	}

	if wl.UserID != userID {
		return domain.ErrForbidden
	}

	return s.repo.DeleteItem(ctx, id)
}
