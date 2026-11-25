package service

import (
	"context"
	"fmt"

	"simple-userfollow-graphql/graph/model"
	"simple-userfollow-graphql/tools"

	"strings"
	"time"

	"gorm.io/gorm"
)

func (s *Service) UserOnCreate(ctx context.Context, input model.NewUser) (*model.User, error) {
	if input.Email == "" || input.Password == "" {
		err := tools.NewGqlError("Email and password is required", tools.GqlBadInput)
		return nil, err
	}

	input.Email = strings.ToLower(input.Email)
	input.Email = strings.TrimSpace(input.Email)

	valid := tools.CheckEmailValidity(input.Email)
	if !valid {
		err := tools.NewGqlError("Invalid email format", tools.GqlBadInput)
		return nil, err
	}

	exist, err := s.UserCheckExistByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}

	if exist {
		err := tools.NewGqlError("Email already registered", tools.GqlConflict)
		return nil, err
	}

	password, err := tools.HashAndSalt(input.Password)
	if err != nil {
		return nil, fmt.Errorf("error hash password: %v", err)
	}

	input.Password = password

	return s.UserCreate(ctx, input)
}

func (s *Service) UserCreate(ctx context.Context, input model.NewUser) (*model.User, error) {
	user := model.User{
		Name:      input.Name,
		Email:     input.Email,
		Password:  input.Password,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.DB.Model(&user).Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Service) UserLogin(ctx context.Context, input model.LoginUser) (*model.LoginResponse, error) {
	if input.Email == "" || input.Password == "" {
		err := tools.NewGqlError("Email and password is required", tools.GqlBadInput)
		return nil, err
	}

	input.Email = strings.ToLower(input.Email)
	input.Email = strings.TrimSpace(input.Email)

	valid := tools.CheckEmailValidity(input.Email)
	if !valid {
		err := tools.NewGqlError("Invalid email format", tools.GqlBadInput)
		return nil, err
	}

	user, err := s.UserGetByEmail(ctx, input.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = tools.NewGqlError("Invalid email/password", tools.GqlBadInput)
			return nil, err

		} else {
			return nil, err
		}
	}

	valid, err = tools.CompareHash(user.Password, input.Password)
	if err != nil {
		return nil, err
	}

	if !valid {
		err := tools.NewGqlError("Invalid email/password", tools.GqlBadInput)
		return nil, err
	}

	accessToken, _, err := tools.TokenCreate(user.ID)
	if err != nil {
		return nil, fmt.Errorf("error generate access token: %v", err)
	}

	return &model.LoginResponse{
		Token: accessToken,
		User:  user,
	}, nil
}

func (s *Service) UserCheckExistByEmail(ctx context.Context, email string) (bool, error) {
	var (
		count int64
		exist bool = false
	)

	if err := s.DB.Model(&model.User{}).Scopes(tools.IsDeletedAtNull).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}

	if int(count) > 0 {
		exist = true
	}

	return exist, nil
}
func (s *Service) UserGetByEmail(ctx context.Context, email string) (*model.User, error) {
	var (
		user model.User
	)

	if err := s.DB.Model(&user).Scopes(tools.IsDeletedAtNull).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
