package repository

import (
	"booking/config"
	"booking/models"
	sqldriver "booking/sql_driver"
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type postgressDBRepo struct {
	App *config.AppConfig
	DB  *sqldriver.DB
}

func NewPostgresRepo(a *config.AppConfig, db *sqldriver.DB) DatabaseRepo {
	return &postgressDBRepo{
		App: a,
		DB:  db,
	}
}
func (p *postgressDBRepo) AllUsers() bool {
	return true
}

func (p *postgressDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into reservations (first_name, last_name, email, phone, start_date, end_date, room_id, created_at, updated_at)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`

	var newID int
	err := p.DB.SQL.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now()).Scan(&newID)

	return newID, err
}

func (p *postgressDBRepo) InsertRoomRestriction(r models.RoomRestriction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into room_restrictions (start_date, end_date, room_id, reservation_id, created_at, updated_at, restriction_id) 
	values ($1, $2, $3, $4, $5, $6, $7)`

	_, err := p.DB.SQL.ExecContext(ctx, stmt,
		r.StartDate,
		r.EndDate,
		r.RoomID,
		r.ReservationID,
		time.Now(),
		time.Now(),
		r.RestrictionID)

	return 0, err
}

func (p *postgressDBRepo) SearchAvailabilityByDatesByRoomID(roomID int, start, end time.Time) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		select 
			count(id)
		from
			room_restrictions
		where 
			room_id = $1 
			and $2 <= end_date and $3 >= start_date;
	`

	var numRows int
	err := p.DB.SQL.QueryRowContext(ctx, query, roomID, start, end).Scan(&numRows)

	if err != nil {
		return false, err
	}

	return numRows == 0, nil
}

func (p *postgressDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `
	select 
		r.id, r.room_name
	from 
		rooms r
	where 
		r.id not in (select room_id from room_restrictions rr where $1 < rr.end_date and $2 > rr.start_date)
	`

	rows, err := p.DB.SQL.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.ID, &room.RoomName)

		if err != nil {
			return rooms, err
		}

		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

func (p *postgressDBRepo) GetRoomByID(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	query := `select id, room_name, created_at, updated_at from rooms where id = $1`

	row := p.DB.SQL.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&room.ID,
		&room.RoomName,
		&room.CreatedAt,
		&room.UpdatedAt,
	)

	return room, err
}

func (p *postgressDBRepo) GetUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, first_name, last_name, email, password, acces_level, created_at, updated_at
			  from users 
			  wher id = $1`

	row := p.DB.SQL.QueryRowContext(ctx, query, id)
	var u models.User
	err := row.Scan(&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.AccessLevel,
		&u.CreatedAt,
		&u.UpdatedAt)

	if err != nil {
		return models.User{}, err
	}

	return u, nil
}

func (p *postgressDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update users set first_name = $1, last_name = $2, email = $3, acces_level = $4, updated_at = $5
	`

	_, err := p.DB.SQL.ExecContext(ctx, query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.AccessLevel,
		time.Now())

	return err
}

func (p *postgressDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	row := p.DB.SQL.QueryRowContext(ctx, "select id, password from users where email = $1", email)

	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil
}

func (p *postgressDBRepo) AllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at, r.processed, rm.id, rm.room_name
		from reservations r
		left join rooms rm 
		on (r.room_id = rm.id)
		order by r.start_date asc
	`

	rows, err := p.DB.SQL.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var m models.Reservation
		err = rows.Scan(
			&m.ID,
			&m.FirstName,
			&m.LastName,
			&m.Email,
			&m.Phone,
			&m.StartDate,
			&m.EndDate,
			&m.RoomID,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.Processed,
			&m.Room.ID,
			&m.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, m)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

func (p *postgressDBRepo) AllNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at, rm.id, rm.room_name
		from reservations r
		left join rooms rm 
		on (r.room_id = rm.id)
		where processed = 0
		order by r.start_date asc
	`

	rows, err := p.DB.SQL.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var m models.Reservation
		err = rows.Scan(
			&m.ID,
			&m.FirstName,
			&m.LastName,
			&m.Email,
			&m.Phone,
			&m.StartDate,
			&m.EndDate,
			&m.RoomID,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.Room.ID,
			&m.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, m)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

func (p *postgressDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var res models.Reservation

	query := `
		select r.id, r.first_name, r.last_name, r.email, r.phone, 
			r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at, r.processed,
			rm.id, rm.room_name
		from reservations r
		left join rooms rm 
		on (r.room_id = rm.id)
		where r.id = $1
	`

	row := p.DB.SQL.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&res.ID,
		&res.FirstName,
		&res.LastName,
		&res.Email,
		&res.Phone,
		&res.StartDate,
		&res.EndDate,
		&res.RoomID,
		&res.CreatedAt,
		&res.UpdatedAt,
		&res.Processed,
		&res.Room.ID,
		&res.Room.RoomName,
	)

	if err != nil {
		return res, err
	}

	return res, nil
}

func (p *postgressDBRepo) UpdateReservation(r models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update reservations set first_name = $1, last_name = $2, email = $3, phone = $4, updated_at = $5
		where id = $6
	`

	_, err := p.DB.SQL.ExecContext(ctx, query,
		r.FirstName,
		r.LastName,
		r.Email,
		r.Phone,
		time.Now(),
		r.ID)

	return err
}

func (p *postgressDBRepo) DeleteReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		delete from reservations where id = $1
	`

	_, err := p.DB.SQL.ExecContext(ctx, query, id)
	return err
}

func (p *postgressDBRepo) UpdateProcessedForReservation(id, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update reservations set processed = $1 where id = $2
	`

	_, err := p.DB.SQL.ExecContext(ctx, query, processed, id)
	return err
}

func (p *postgressDBRepo) AllRooms() ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `select id, room_name, created_at, updated_at from rooms order by room_name`

	rows, err := p.DB.SQL.QueryContext(ctx, query)
	if err != nil {
		return rooms, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.Room
		err = rows.Scan(
			&r.ID,
			&r.RoomName,
			&r.CreatedAt,
			&r.UpdatedAt,
		)

		if err != nil {
			return rooms, err
		}

		rooms = append(rooms, r)
	}

	return rooms, rows.Err()
}

func (p *postgressDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var restrictions []models.RoomRestriction
	query := `select id, coalesce(reservation_id, 0), restriction_id, room_id, start_date, end_date
			  from room_restrictions where $1 < end_date and $2 >= start_date and room_id = $3`

	rows, err := p.DB.SQL.QueryContext(ctx, query, start, end, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.RoomRestriction
		err = rows.Scan(
			&r.ID,
			&r.ReservationID,
			&r.RestrictionID,
			&r.RoomID,
			&r.StartDate,
			&r.EndDate,
		)

		if err != nil {
			return nil, err
		}

		restrictions = append(restrictions, r)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return restrictions, nil
}

func (p *postgressDBRepo) InsertBlockForRoom(id int, startDate time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `insert into room_restrictions (start_date, end_date, room_id, restriction_id, created_at, updated_at) 
			  values ($1, $2, $3, $4, $5, $6)`

	_, err := p.DB.SQL.ExecContext(ctx, query, startDate, startDate.AddDate(0, 0, 1), id, 2, time.Now(), time.Now())
	return err
}

func (p *postgressDBRepo) DeleteBlockByID(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `delete from room_restrictions where id = $1`

	_, err := p.DB.SQL.ExecContext(ctx, query, id)
	return err
}
