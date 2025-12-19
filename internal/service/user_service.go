package service

import (
	"flashsale/internal/domain/user"
	"fmt"
)

// mock user service, for load testing purpose only

type MockUserService struct {
	mockUsers map[string]*user.User
}

func NewMockUserService() *MockUserService {
	return &MockUserService{
		mockUsers: map[string]*user.User{
			"1": {ID: "1", Email: "user1@test.com"},
			"2": {ID: "2", Email: "user2@test.com"},
			"3": {ID: "3", Email: "user3@test.com"},
		},
	}
}

func (s *MockUserService) GetUserByID(userID string) (*user.User, error) {
	u, ok := s.mockUsers[userID]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", userID)
	}
	return u, nil
}
