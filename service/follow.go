package service

import (
	"context"
	"errors"
	"simple-userfollow-graphql/graph/model"
	"simple-userfollow-graphql/tools"
	"time"

	"gorm.io/gorm"
)

func (s *Service) UserGetByID(ctx context.Context, id int) (*model.User, error) {
	var user model.User

	err := s.DB.WithContext(ctx).
		Model(&model.User{}).
		Scopes(tools.IsDeletedAtNull).
		Where("id = ?", id).
		First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Service) UserGetFollowersByIDs(
	ctx context.Context,
	userIDs []int,
) (map[int][]*model.User, error) {
	type row struct {
		OwnerID   int        `gorm:"column:owner_id"`
		ID        int        `gorm:"column:id"`
		Name      string     `gorm:"column:name"`
		Email     string     `gorm:"column:email"`
		CreatedAt time.Time  `gorm:"column:created_at"`
		UpdatedAt *time.Time `gorm:"column:updated_at"`
	}

	var rows []row
	err := s.DB.WithContext(ctx).
		Table("user AS u").
		Select("uf.following_id AS owner_id, u.id, u.name, u.email, u.created_at, u.updated_at").
		Joins("JOIN user_follow AS uf ON uf.follower_id = u.id").
		Where("uf.following_id IN ?", userIDs).
		Where("u.deleted_at IS NULL").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int][]*model.User)
	for _, row := range rows {
		result[row.OwnerID] = append(result[row.OwnerID], &model.User{
			ID:        row.ID,
			Name:      row.Name,
			Email:     row.Email,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}

	return result, nil
}

func (s *Service) UserGetFollowingByIDs(
	ctx context.Context,
	userIDs []int,
) (map[int][]*model.User, error) {
	type row struct {
		OwnerID   int        `gorm:"column:owner_id"`
		ID        int        `gorm:"column:id"`
		Name      string     `gorm:"column:name"`
		Email     string     `gorm:"column:email"`
		CreatedAt time.Time  `gorm:"column:created_at"`
		UpdatedAt *time.Time `gorm:"column:updated_at"`
	}

	var rows []row
	err := s.DB.WithContext(ctx).
		Table("user AS u").
		Select("uf.follower_id AS owner_id, u.id, u.name, u.email, u.created_at, u.updated_at").
		Joins("JOIN user_follow AS uf ON uf.following_id = u.id").
		Where("uf.follower_id IN ?", userIDs).
		Where("u.deleted_at IS NULL").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int][]*model.User)
	for _, row := range rows {
		result[row.OwnerID] = append(result[row.OwnerID], &model.User{
			ID:        row.ID,
			Name:      row.Name,
			Email:     row.Email,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}

	return result, nil
}

func (s *Service) UserGetFollowers(ctx context.Context, userID int) ([]*model.User, error) {
	users := make([]*model.User, 0)

	err := s.DB.WithContext(ctx).
		Table("user AS u").
		Select("u.id, u.name, u.email, u.created_at, u.updated_at").
		Joins("JOIN user_follow AS uf ON uf.follower_id = u.id").
		Where("uf.following_id = ?", userID).
		Where("u.deleted_at IS NULL").
		Find(&users).Error

	return users, err
}

func (s *Service) UserGetFollowing(ctx context.Context, userID int) ([]*model.User, error) {
	users := make([]*model.User, 0)

	err := s.DB.WithContext(ctx).
		Table("user AS u").
		Select("u.id, u.name, u.email, u.created_at, u.updated_at").
		Joins("JOIN user_follow AS uf ON uf.following_id = u.id").
		Where("uf.follower_id = ?", userID).
		Where("u.deleted_at IS NULL").
		Find(&users).Error

	return users, err
}

func (s *Service) UserGetByIDWithRelations(ctx context.Context, id int) (*model.User, error) {
	user, err := s.UserGetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Followers, err = s.UserGetFollowers(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Following, err = s.UserGetFollowing(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) UserFollow(ctx context.Context, followerID, followingID int) error {
	if followerID == followingID {
		return tools.NewGqlError("You cannot follow yourself", tools.GqlBadInput)
	}

	if _, err := s.UserGetByID(ctx, followingID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tools.NewGqlError("User not found", tools.GqlNotFound)
		}
		return err
	}

	var count int64
	err := s.DB.WithContext(ctx).
		Table("user_follow").
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return tools.NewGqlError("Already following this user", tools.GqlConflict)
	}

	return s.DB.WithContext(ctx).Exec(
		"INSERT INTO user_follow (follower_id, following_id) VALUES (?, ?)",
		followerID,
		followingID,
	).Error
}

func (s *Service) UserUnfollow(ctx context.Context, followerID, followingID int) error {
	result := s.DB.WithContext(ctx).
		Exec(
			"DELETE FROM user_follow WHERE follower_id = ? AND following_id = ?",
			followerID,
			followingID,
		)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return tools.NewGqlError("You are not following this user", tools.GqlNotFound)
	}

	return nil
}
