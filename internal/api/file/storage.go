package file

import (
	"context"
)

type FileRepository interface {
	Create(ctx context.Context, fl *File) error
	FindAll(ctx context.Context) (files []File, err error)
	FindOne(ctx context.Context, id string) (File, error)
	Update(ctx context.Context, fl *File) (files []File, err error)
	Delete(ctx context.Context, id string) ([]string, error)
}
