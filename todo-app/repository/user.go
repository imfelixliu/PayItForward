package repository

import (
	"database/sql"
	"todo-app/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Upsert(githubID int64, email, name, avatarURL string) (models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
		INSERT INTO users (github_id, email, name, avatar_url)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (github_id) DO UPDATE
		  SET email = EXCLUDED.email,
		      name = EXCLUDED.name,
		      avatar_url = EXCLUDED.avatar_url
		RETURNING id, github_id, COALESCE(email,''), COALESCE(name,''), COALESCE(avatar_url,''), created_at
	`, githubID, nullableString(email), name, avatarURL,
	).Scan(&user.ID, &user.GithubID, &user.Email, &user.Name, &user.AvatarURL, &user.CreatedAt)
	return user, err
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
