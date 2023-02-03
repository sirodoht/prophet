package internal

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type SQLStore struct {
	db *sqlx.DB
}

func NewSQLStore(db *sqlx.DB) *SQLStore {
	return &SQLStore{
		db: db,
	}
}

func (s *SQLStore) InsertPost(
	ctx context.Context,
	d *Post,
) (int64, error) {
	var id int64
	rows, err := s.db.NamedQuery(`
		INSERT INTO posts (
			title,
			body,
			created_at,
			updated_at
		) VALUES (
			:title,
			:body,
			:created_at,
			:updated_at
		) RETURNING id`, d)
	if err != nil {
		return 0, err
	}
	if rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			panic(err)
		}
	}
	return id, nil
}

func (s *SQLStore) UpdatePost(
	ctx context.Context,
	id int64,
	field string,
	value string,
) error {
	sql := fmt.Sprintf(`
		UPDATE posts
		SET
			%s=:value,
			updated_at=now()
		WHERE id=:id
	`, field)
	_, err := s.db.NamedExec(sql, map[string]interface{}{
		"field": field,
		"value": value,
		"id":    id,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLStore) GetAllPost(ctx context.Context) ([]*Post, error) {
	var posts []*Post
	err := s.db.SelectContext(
		ctx,
		&posts,
		`SELECT * FROM posts ORDER BY title ASC`,
	)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (s *SQLStore) GetOnePost(
	ctx context.Context,
	id int64,
) (*Post, error) {
	var posts []*Post
	err := s.db.SelectContext(
		ctx,
		&posts,
		`SELECT * FROM posts WHERE id=$1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	return posts[0], nil
}
