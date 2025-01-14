package file

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"app/pkg/client/postgresql"
	"app/pkg/logging"

	"github.com/jackc/pgconn"
)

type repository struct {
	client postgresql.Client
	logger *logging.Logger
}

func NewRepository(logger *logging.Logger, client postgresql.Client) FileRepository {
	return &repository{
		client: client,
		logger: logger,
	}
}

func formatQuery(q string) string {
	return strings.ReplaceAll(strings.ReplaceAll(q, "\t", ""), "\n", " ")
}

func (r *repository) Create(ctx context.Context, curFile *File) error {
	q := `
		INSERT INTO files 
			(name, data)
		VALUES 
		    ($1, $2)
		RETURNING id;
	`

	if err := r.client.QueryRow(ctx, q, curFile.Name, curFile.Data).Scan(&curFile.ID); err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			pgErr = err.(*pgconn.PgError)
			newErr := fmt.Errorf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
			r.logger.Error(newErr)
			return newErr
		}
		r.logger.Error(err)
		return err
	}

	response := fmt.Sprintf("SQL Query: %s\n\tResult: adding tool %s", formatQuery(q), curFile)

	r.logger.Debug(response)

	return nil
}

func (r *repository) FindAll(ctx context.Context) (files []File, err error) {
	q := `
		SELECT 
			id,
			name,
			data,
			create_time, 
			update_time
		FROM 
			files
	`

	rows, err := r.client.Query(ctx, q)

	if err != nil {
		r.logger.Error(rows.CommandTag(), err)
		return nil, err
	}

	files = make([]File, 0)
	for rows.Next() {
		var fl File

		err = rows.Scan(
			&fl.ID,
			&fl.Name,
			&fl.Data,
			&fl.CreatedAt,
			&fl.UpdatedAt,
		)
		if err != nil {
			r.logger.Error(err)
			return nil, err
		}

		files = append(files, fl)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error(err)
		return nil, err
	}

	res := rows.CommandTag()
	response := fmt.Sprintf("SQL Query: %s", formatQuery(q)+"\n\tResult: "+res.String())
	r.logger.Debug(response)

	return files, nil
}

func (r *repository) FindOne(ctx context.Context, id string) (File, error) {
	q := `
	SELECT id, name, data, create_time, update_time FROM files WHERE id = $1;
	`

	var fl File
	err := r.client.QueryRow(ctx, q, id).Scan(&fl.ID, &fl.Name, &fl.Data, &fl.CreatedAt, &fl.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return File{}, nil
		}
		r.logger.Error(err)
		return File{}, err
	}

	return fl, nil
}

func (r *repository) Update(ctx context.Context, curFile *File) (files []File, err error) {
	q := `
	UPDATE files SET
		name = $1,
		data = $2
	WHERE id = $3
	RETURNING 
		create_time,
		update_time;
	`

	rows, err := r.client.Query(ctx, q, curFile.Name, curFile.Data, curFile.ID)
	if err != nil {
		return nil, err
	}
	files = make([]File, 0)
	for rows.Next() {
		var fl File
		err = rows.Scan(
			&fl.CreatedAt,
			&fl.UpdatedAt,
		)
		if err != nil {
			r.logger.Error(err, rows.CommandTag())
			return nil, err
		}
		files = append(files, fl)
	}
	if err = rows.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			pgErr = err.(*pgconn.PgError)
			newErr := fmt.Errorf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState())
			r.logger.Error(newErr)
		}
		return nil, err

	}

	res := rows.CommandTag()
	response := fmt.Sprintf("SQL Query: %s", formatQuery(q)+"\n\tResult: "+res.String())
	r.logger.Debug(response)

	return files, nil
}

func (r *repository) Delete(ctx context.Context, id string) ([]string, error) {
	q := `
	DELETE FROM files
	WHERE id = $1
	RETURNING id;
	`

	rows, err := r.client.Query(ctx, q, id)
	if err != nil {
		r.logger.Error(err)
		return nil, err
	}

	ids := []string{}
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			r.logger.Error(err, rows.CommandTag())
			return nil, err
		}
		ids = append(ids, id)
	}

	res := rows.CommandTag()
	response := fmt.Sprintf("SQL Query: %s", formatQuery(q)+"\n\tResult: "+res.String())
	r.logger.Debug(response)

	return ids, nil
}
