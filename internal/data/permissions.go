package data

import (
	"context"
	"database/sql"
	"time"
)

// define permission slice
type Permissions []string

func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}

	return false
}

// permission model for database
type PermissionsModel struct {
	DB *sql.DB
}

func (m PermissionsModel) GetAllForUser(userID int64) (Permissions, error) {
	query := `SELECT  permissions.code 
            FROM permissions
            INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
            INNER JOIN users ON users_permissions.user_id = users.id
            WHERE users.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var permissions Permissions

	for rows.Next() {
		var permission string

		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil

}
