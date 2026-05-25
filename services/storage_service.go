package services

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

type StorageService interface {
	UploadAvatar(ctx context.Context, originalFileName string, file io.Reader, fileSize int64) (string, error)
	DeleteFile(ctx context.Context, fileURL string) error
	GetPresignedURL(ctx context.Context, fileURL string) (string, error)
}

type storageServiceImpl struct {
	minioClient *minio.Client
	bucketName  string
}

func NewStorageService(client *minio.Client) StorageService {
	return &storageServiceImpl{
		minioClient: client,
		bucketName:  "cobago",
	}
}

func (s *storageServiceImpl) GetPresignedURL(ctx context.Context, fileURL string) (string, error) {
	if fileURL == "" {
		return "", nil
	}

	// 1. Ambil nama objek dari URL asli
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", err
	}
	pathParts := strings.Split(parsedURL.Path, "/")
	if len(pathParts) < 3 {
		return "", fmt.Errorf("invalid file URL structure")
	}
	objectName := pathParts[len(pathParts)-1]

	// 2. Set durasi kedaluwarsa URL (misal 15 menit)
	expiry := time.Duration(15) * time.Minute

	// 3. Set Header agar browser langsung merender sebagai gambar, bukan mengalihkan ke Console
	reqParams := make(url.Values)
	reqParams.Set("response-content-type", "image/jpeg") // Memaksa browser tahu ini gambar
	reqParams.Set("response-content-disposition", "inline")

	// 4. Generate Presigned URL resmi
	presignedURL, err := s.minioClient.PresignedGetObject(ctx, s.bucketName, objectName, expiry, reqParams)
	if err != nil {
		return "", err
	}

	return presignedURL.String(), nil
}

func (s *storageServiceImpl) UploadAvatar(ctx context.Context, originalFileName string, file io.Reader, fileSize int64) (string, error) {
	ext := filepath.Ext(originalFileName)

	uniqueFileName := fmt.Sprintf("%d_avatar%s", time.Now().UnixNano(), ext)

	_, err := s.minioClient.PutObject(ctx, s.bucketName, uniqueFileName, file, fileSize, minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return "", err
	}

	return "http://localhost:9000/" + s.bucketName + "/" + uniqueFileName, nil
}

func (s *storageServiceImpl) DeleteFile(ctx context.Context, fileURL string) error {
	if fileURL == "" {
		return nil // Jika tidak ada URL gambar lama, abaikan
	}

	// Parsing URL untuk mengambil nama filenya saja
	// Contoh: http://localhost:9000/cobago/12345_avatar.jpg -> Ambil "12345_avatar.jpg"
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return err
	}

	// Ambil bagian path (misal: /cobago/12345_avatar.jpg)
	pathParts := strings.Split(parsedURL.Path, "/")
	if len(pathParts) < 3 {
		return nil
	}

	// Nama file berada di elemen paling terakhir dari split path
	objectName := pathParts[len(pathParts)-1]

	// Jalankan perintah hapus dari MinIO
	err = s.minioClient.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}