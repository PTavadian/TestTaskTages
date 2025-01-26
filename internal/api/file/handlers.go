package file

import (
	pb "app/api/proto"
	"app/pkg/logging"
	"context"
	"fmt"
)

type Server struct {
	pb.UnimplementedFileServiceServer

	Logger         *logging.Logger
	FileRepository FileRepository

	UploadSemaphore   chan struct{}
	DownloadSemaphore chan struct{}
	ListSemaphore     chan struct{}
}

func NewServer(logger *logging.Logger, fileRepository FileRepository) *Server {
	return &Server{
		FileRepository:    fileRepository,
		Logger:            logger,
		UploadSemaphore:   make(chan struct{}, 10),
		DownloadSemaphore: make(chan struct{}, 10),
		ListSemaphore:     make(chan struct{}, 100),
	}
}

func (s *Server) UploadFile(ctx context.Context, req *pb.UploadFileRequest) (*pb.UploadFileResponse, error) {
	s.UploadSemaphore <- struct{}{}
	fmt.Println("Uploading started:", req.FileName)
	defer func() {
		<-s.UploadSemaphore
		fmt.Println("Uploading finished:", req.FileName)
	}()

	newFile := File{Name: req.FileName, Data: req.Data}

	err := s.FileRepository.Create(ctx, &newFile)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("Failed to create file: %v", err))
		return nil, err
	}

	s.Logger.Info(fmt.Sprintf("File uploaded successfully: %s", newFile.Name))
	return &pb.UploadFileResponse{Id: newFile.ID}, nil
}

func (s *Server) DownloadFile(ctx context.Context, req *pb.DownloadFileRequest) (*pb.DownloadFileResponse, error) {
	s.DownloadSemaphore <- struct{}{}
	defer func() { <-s.DownloadSemaphore }()

	fl, err := s.FileRepository.FindOne(ctx, req.Id)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("Failed to find file: %v", err))
		return nil, err
	}

	return &pb.DownloadFileResponse{FileName: fl.Name, Data: fl.Data, CreatedAt: fl.CreatedAt.Unix(), UpdatedAt: fl.UpdatedAt.Unix()}, nil
}

func (s *Server) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	s.ListSemaphore <- struct{}{}
	defer func() { <-s.ListSemaphore }()

	files, err := s.FileRepository.FindAll(ctx)
	if err != nil {
		s.Logger.Error(fmt.Sprintf("Failed to find all files: %v", err))
		return nil, err
	}

	var fileMetadataList []*pb.FileMetadata
	for _, file := range files {
		fileMetadataList = append(fileMetadataList, &pb.FileMetadata{
			Name:      file.Name,
			CreatedAt: file.CreatedAt.Unix(),
			UpdatedAt: file.UpdatedAt.Unix(),
		})
	}

	return &pb.ListFilesResponse{Files: fileMetadataList}, nil
}
