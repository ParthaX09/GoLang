package database

import (
	"database/sql"
	"server/models"
)

func CreateUser(user *models.User) (int64, error) {
	result, err := DB.Exec(`INSERT INTO users (name, email, phone, password, role) VALUES (?, ?, ?, ?, ?)`,
		user.Name, user.Email, user.Phone, user.Password, user.Role)

	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetUserByEmail(email string) (models.User, error) {
	var user models.User
	err := DB.QueryRow("SELECT id, password, role FROM users WHERE email = ?", email).
		Scan(&user.ID, &user.Password, &user.Role)
	return user, err
}

func GetAllUsersByRole(role string) (*sql.Rows, error) {
	switch role {

	case "admin":
		return DB.Query("SELECT id, name, email, phone, password, role, created, updated FROM users")

	case "sub_admin":
		return DB.Query("SELECT id, name, email, phone, password, role, created, updated FROM users WHERE role IN ('sub_admin', 'client')")

	case "client":
		return DB.Query("SELECT id, name, email, phone, password, role, created, updated FROM users WHERE role = 'client'")

	default:
		return nil, sql.ErrNoRows
	}
}

func GetUserByID(id string) (models.User, error) {
	var user models.User
	err := DB.QueryRow("SELECT id, name, email, phone, password, role, created, updated FROM users WHERE id = ?", id).
		Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.Password, &user.Role, &user.Created, &user.Updated)
	return user, err
}

func UpdateUser(user models.User) error {
	_, err := DB.Exec(`UPDATE users SET name = ?, email = ?, phone = ?, password = ?, role = ? WHERE id = ?`,
		user.Name, user.Email, user.Phone, user.Password, user.Role, user.ID)
	return err
}
