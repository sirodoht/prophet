package internal

import (
	"context"

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
			npub,
			relaylist,
			body,
			created_at
		) VALUES (
			:npub,
			:relaylist,
			:body,
			:created_at
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

func (s *SQLStore) GetAllPostByNpub(
	ctx context.Context,
	npub string,
) ([]*Post, error) {
	var posts []*Post
	err := s.db.SelectContext(
		ctx,
		&posts,
		`SELECT * FROM posts WHERE npub=$1 ORDER BY created_at DESC`,
		npub,
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
