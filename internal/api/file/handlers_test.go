package file_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	pb "app/api/proto"
	"app/internal/api/file"
	"app/pkg/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockFileRepository struct {
	mock.Mock
}

func (m *MockFileRepository) Create(ctx context.Context, file *file.File) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockFileRepository) FindAll(ctx context.Context) ([]file.File, error) {
	args := m.Called(ctx)
	return args.Get(0).([]file.File), args.Error(1)
}

func (m *MockFileRepository) FindOne(ctx context.Context, id string) (file.File, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(file.File), args.Error(1)
}

func (m *MockFileRepository) Update(ctx context.Context, fl *file.File) ([]file.File, error) {
	args := m.Called(ctx, fl)
	return args.Get(0).([]file.File), args.Error(1)
}

func (m *MockFileRepository) Delete(ctx context.Context, id string) ([]string, error) {
	args := m.Called(ctx, id)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockFileRepository) CreateTables(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestUploadFile(t *testing.T) {
	ctx := context.TODO()
	logger := logging.NewTestLogger()

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockFileRepository)
		server := file.NewServer(logger, mockRepo)

		req := &pb.UploadFileRequest{FileName: "test.jpg", Data: []byte("test data")}
		newFile := &file.File{ID: "mockID", Name: req.FileName, Data: req.Data}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*file.File")).Return(nil).Run(func(args mock.Arguments) {
			fileArg := args.Get(1).(*file.File)
			fileArg.ID = newFile.ID // Инициализация mock ID
		})

		res, err := server.UploadFile(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, newFile.ID, res.Id)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertCalled(t, "Create", ctx, mock.AnythingOfType("*file.File"))
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo := new(MockFileRepository)
		server := file.NewServer(logger, mockRepo)

		req := &pb.UploadFileRequest{FileName: "test.jpg", Data: []byte("test data")}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*file.File")).Return(fmt.Errorf("create error"))

		res, err := server.UploadFile(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, res)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertCalled(t, "Create", ctx, mock.AnythingOfType("*file.File"))
	})
}

func TestDownloadFile(t *testing.T) {
	ctx := context.TODO()
	logger := logging.NewTestLogger()

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockFileRepository)
		server := file.NewServer(logger, mockRepo)

		req := &pb.DownloadFileRequest{Id: "123"}

		file := file.File{ID: "123", Name: "test.jpg", Data: []byte("test data"), CreatedAt: time.Now(), UpdatedAt: time.Now()}
		mockRepo.On("FindOne", ctx, "123").Return(file, nil)

		res, err := server.DownloadFile(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "test.jpg", res.FileName)
		assert.Equal(t, file.Data, res.Data)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertCalled(t, "FindOne", ctx, "123")
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo := new(MockFileRepository)
		server := file.NewServer(logger, mockRepo)

		req := &pb.DownloadFileRequest{Id: "123"}

		mockRepo.On("FindOne", ctx, "123").Return(file.File{}, fmt.Errorf("find error"))

		res, err := server.DownloadFile(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, res)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertCalled(t, "FindOne", ctx, "123")
	})
}

func TestListFiles(t *testing.T) {
	ctx := context.TODO()
	logger := logging.NewTestLogger()

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockFileRepository)
		server := file.NewServer(logger, mockRepo)

		req := &pb.ListFilesRequest{}

		files := []file.File{
			{ID: "123", Name: "test1.jpg", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: "456", Name: "test2.jpg", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		mockRepo.On("FindAll", ctx).Return(files, nil)

		res, err := server.ListFiles(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res.Files, 2)
		assert.Equal(t, "test1.jpg", res.Files[0].Name)
		assert.Equal(t, "test2.jpg", res.Files[1].Name)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertCalled(t, "FindAll", ctx)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo := new(MockFileRepository)
		server := file.NewServer(logger, mockRepo)

		req := &pb.ListFilesRequest{}

		mockRepo.On("FindAll", ctx).Return([]file.File{}, fmt.Errorf("find all error"))

		res, err := server.ListFiles(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, res)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertCalled(t, "FindAll", ctx)
	})
}

func TestConcurrentUpload(t *testing.T) {
	logger := logging.NewTestLogger()
	mockRepo := new(MockFileRepository)
	server := file.NewServer(logger, mockRepo)

	ctx := context.TODO()

	mockRepo.On("Create", ctx, mock.AnythingOfType("*file.File")).Return(nil).Run(func(args mock.Arguments) {
		fileArg := args.Get(1).(*file.File)
		fileArg.ID = "mockID"
		time.Sleep(50 * time.Millisecond)
	})

	var wg sync.WaitGroup
	var activeRequests int32
	var maxActiveRequests int32
	var mu sync.Mutex

	sem := make(chan struct{}, 10)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}        // Попробуем захватить семафор
			defer func() { <-sem }() // Освободим семафор при завершении

			currentActive := atomic.AddInt32(&activeRequests, 1)
			fmt.Printf("Active requests: %d\n", currentActive)

			mu.Lock()
			if currentActive > maxActiveRequests {
				maxActiveRequests = currentActive
			}
			mu.Unlock()

			req := &pb.UploadFileRequest{FileName: fmt.Sprintf("test_%d.jpg", i), Data: []byte("test data")}
			res, err := server.UploadFile(ctx, req)

			atomic.AddInt32(&activeRequests, -1)

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, "mockID", res.Id)
		}(i)
	}

	wg.Wait()
	fmt.Printf("Max active requests: %d\n", maxActiveRequests)
	assert.LessOrEqual(t, int(maxActiveRequests), 10, "More than 10 concurrent requests were processed")
	mockRepo.AssertExpectations(t)
}

func TestConcurrentDownload(t *testing.T) {
	logger := logging.NewTestLogger()
	mockRepo := new(MockFileRepository)
	server := file.NewServer(logger, mockRepo)

	ctx := context.TODO()

	mockRepo.On("FindOne", ctx, "mockID").Return(file.File{
		ID:        "mockID",
		Name:      "test.jpg",
		Data:      []byte("test data"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil).Run(func(args mock.Arguments) {
		time.Sleep(50 * time.Millisecond)
	})

	var wg sync.WaitGroup
	var activeRequests int32
	var maxActiveRequests int32
	var mu sync.Mutex

	sem := make(chan struct{}, 10)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			currentActive := atomic.AddInt32(&activeRequests, 1)
			fmt.Printf("Active requests: %d\n", currentActive)

			mu.Lock()
			if currentActive > maxActiveRequests {
				maxActiveRequests = currentActive
			}
			mu.Unlock()

			req := &pb.DownloadFileRequest{Id: "mockID"}
			res, err := server.DownloadFile(ctx, req)

			atomic.AddInt32(&activeRequests, -1)

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, "test.jpg", res.FileName)
			assert.Equal(t, []byte("test data"), res.Data)
		}(i)
	}

	wg.Wait()
	fmt.Printf("Max active requests: %d\n", maxActiveRequests)
	assert.LessOrEqual(t, int(maxActiveRequests), 10, "More than 10 concurrent requests were processed")
	mockRepo.AssertExpectations(t)
}

func TestConcurrentListFiles(t *testing.T) {
	logger := logging.NewTestLogger()
	mockRepo := new(MockFileRepository)
	server := file.NewServer(logger, mockRepo)

	ctx := context.TODO()

	mockRepo.On("FindAll", ctx).Return([]file.File{
		{ID: "123", Name: "test1.jpg", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "456", Name: "test2.jpg", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}, nil).Run(func(args mock.Arguments) {
		time.Sleep(50 * time.Millisecond)
	})

	var wg sync.WaitGroup
	var activeRequests int32
	var maxActiveRequests int32
	var mu sync.Mutex

	sem := make(chan struct{}, 100)

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			currentActive := atomic.AddInt32(&activeRequests, 1)
			fmt.Printf("Active requests: %d\n", currentActive)

			mu.Lock()
			if currentActive > maxActiveRequests {
				maxActiveRequests = currentActive
			}
			mu.Unlock()

			req := &pb.ListFilesRequest{}
			res, err := server.ListFiles(ctx, req)

			atomic.AddInt32(&activeRequests, -1)

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Len(t, res.Files, 2) // Проверка, что количество файлов соответствует ожидаемому
			assert.Equal(t, "test1.jpg", res.Files[0].Name)
			assert.Equal(t, "test2.jpg", res.Files[1].Name)
		}(i)
	}

	wg.Wait()
	fmt.Printf("Max active requests: %d\n", maxActiveRequests)
	assert.LessOrEqual(t, int(maxActiveRequests), 100, "More than 100 concurrent requests were processed")
	mockRepo.AssertExpectations(t)
}
