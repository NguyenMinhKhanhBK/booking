package repository

import (
	"booking/models"
	"time"
)

//go:generate mockgen -destination=../mocks/mock_database_repo.go -package=mocks -source=${GOFILE}
type DatabaseRepo interface {
	AllUsers() bool
	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction(r models.RoomRestriction) (int, error)
	SearchAvailabilityByDatesByRoomID(roomID int, start, end time.Time) (bool, error)
	SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error)
	GetRoomByID(id int) (models.Room, error)
	UpdateUser(u models.User) error
	Authenticate(email, testPassword string) (int, string, error)
}
