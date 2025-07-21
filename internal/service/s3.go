package service

import (
	"evolution-postgres-backup/internal/config"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Client struct {
	client   *s3.S3
	uploader *s3manager.Uploader
	session  *session.Session
	bucket   string
}

func NewS3Client() *S3Client {
	return &S3Client{}
}

func (s *S3Client) Initialize(cfg *config.Config) error {
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			cfg.S3Config.AccessKeyID,
			cfg.S3Config.SecretAccessKey,
			"",
		),
		Region:           aws.String(cfg.S3Config.Region),
		DisableSSL:       aws.Bool(!cfg.S3Config.UseSSL),
		S3ForcePathStyle: aws.Bool(true), // Required for MinIO and other S3-compatible services
	}

	if cfg.S3Config.Endpoint != "" {
		s3Config.Endpoint = aws.String(cfg.S3Config.Endpoint)
	}

	sess, err := session.NewSession(s3Config)
	if err != nil {
		return fmt.Errorf("failed to create S3 session: %w", err)
	}

	s.client = s3.New(sess)
	s.uploader = s3manager.NewUploader(sess)
	s.session = sess
	s.bucket = cfg.S3Config.Bucket

	// Test connection
	_, err = s.client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		return fmt.Errorf("failed to access S3 bucket '%s': %w", s.bucket, err)
	}

	log.Printf("S3 client initialized successfully for bucket: %s", s.bucket)
	return nil
}

func (s *S3Client) UploadFile(filePath, s3Key string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	_, err = s.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}

	log.Printf("Successfully uploaded %s to S3 as %s", filePath, s3Key)
	return nil
}

func (s *S3Client) DownloadFile(s3Key, localPath string) error {
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file %s: %w", localPath, err)
	}
	defer file.Close()

	downloader := s3manager.NewDownloader(s.session)
	_, err = downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return fmt.Errorf("failed to download file from S3: %w", err)
	}

	log.Printf("Successfully downloaded %s from S3 to %s", s3Key, localPath)
	return nil
}

func (s *S3Client) DeleteFile(s3Key string) error {
	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	log.Printf("Successfully deleted %s from S3", s3Key)
	return nil
}

func (s *S3Client) ListFiles(prefix string) ([]*s3.Object, error) {
	result, err := s.client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files from S3: %w", err)
	}

	return result.Contents, nil
}

func (s *S3Client) GetFileSize(s3Key string) (int64, error) {
	result, err := s.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get file size from S3: %w", err)
	}

	return *result.ContentLength, nil
}

func (s *S3Client) FileExists(s3Key string) bool {
	_, err := s.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
	})
	return err == nil
}

func (s *S3Client) CleanupOldBackups(prefix string, retentionCount int) error {
	objects, err := s.ListFiles(prefix)
	if err != nil {
		return err
	}

	if len(objects) <= retentionCount {
		return nil
	}

	// Sort by last modified (oldest first)
	// AWS SDK already returns objects sorted by key name, which should be chronological
	// for our backup naming scheme

	objectsToDelete := objects[:len(objects)-retentionCount]

	for _, obj := range objectsToDelete {
		if err := s.DeleteFile(*obj.Key); err != nil {
			log.Printf("Failed to delete old backup %s: %v", *obj.Key, err)
		}
	}

	log.Printf("Cleaned up %d old backups with prefix %s", len(objectsToDelete), prefix)
	return nil
}

func GenerateS3Key(postgresID, backupType, timestamp, filename string) string {
	// Format: backups/{postgres_id}/{backup_type}/{year}/{month}/{filename}
	parts := strings.Split(timestamp, "-")
	year, month := parts[0], parts[1]

	return fmt.Sprintf("backups/%s/%s/%s/%s/%s", postgresID, backupType, year, month, filename)
}
