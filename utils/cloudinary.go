package utils

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var cld *cloudinary.Cloudinary

// Init Cloudinary
func InitCloudinary() error {
    cloudinaryURL := os.Getenv("CLOUDINARY_URL")
    
    if cloudinaryURL == "" {
        return fmt.Errorf("CLOUDINARY_URL environment variable is required")
    }
    
    var err error
    cld, err = cloudinary.NewFromURL(cloudinaryURL)
    if err != nil {
        return fmt.Errorf("failed to initialize Cloudinary: %v", err)
    }
    
    log.Println("✅ Cloudinary initialized")
    return nil
}

// Get Cloudinary instance
func GetCloudinary() *cloudinary.Cloudinary {
    return cld
}

// UploadResult result dari upload
type UploadResult struct {
    PublicID  string `json:"public_id"`
    URL       string `json:"url"`
    SecureURL string `json:"secure_url"`
    Format    string `json:"format"`
    Bytes     int    `json:"bytes"`
}

// UploadFile upload file ke Cloudinary
func UploadFile(file *multipart.FileHeader, folder string) (*UploadResult, error) {
    if cld == nil {
        return nil, fmt.Errorf("cloudinary not initialized")
    }

    // Buka file
    f, err := file.Open()
    if err != nil {
        return nil, fmt.Errorf("failed to open file: %v", err)
    }
    defer f.Close()

    // Generate unique filename
    filename := generateFilename(file.Filename, folder)

    // Upload ke Cloudinary
    ctx := context.Background()
    result, err := cld.Upload.Upload(ctx, f, uploader.UploadParams{
        PublicID:       filename,
        ResourceType:   "image",
        Folder:         folder,
        Overwrite:      api.Bool(true),
        UniqueFilename: api.Bool(true),
        UseFilename:    api.Bool(true),
    })

    if err != nil {
        return nil, fmt.Errorf("upload failed: %v", err)
    }

    return &UploadResult{
        PublicID:  result.PublicID,
        URL:       result.URL,
        SecureURL: result.SecureURL,
        Format:    result.Format,
        Bytes:     result.Bytes,
    }, nil
}

// DeleteFile hapus file dari Cloudinary
func DeleteFile(publicID string) error {
    if cld == nil {
        return fmt.Errorf("cloudinary not initialized")
    }

    ctx := context.Background()
    _, err := cld.Upload.Destroy(ctx, uploader.DestroyParams{
        PublicID:     publicID,
        ResourceType: "image",
    })

    return err
}

// Generate unique filename
func generateFilename(originalName, folder string) string {
    // Ekstrak nama tanpa ekstensi
    ext := filepath.Ext(originalName)
    name := strings.TrimSuffix(originalName, ext)
    
    // Clean nama file
    name = strings.ReplaceAll(name, " ", "_")
    name = strings.ToLower(name)
    
    // Tambah timestamp
    timestamp := time.Now().Unix()
    filename := fmt.Sprintf("%s_%d", name, timestamp)  
    
    return filename
}

// ValidateImage validasi file gambar
func ValidateImage(file *multipart.FileHeader) error {
    // Max 5MB
    if file.Size > 5*1024*1024 {
        return fmt.Errorf("file too large (max 5MB)")
    }

    // Allowed extensions
    ext := strings.ToLower(filepath.Ext(file.Filename))
    allowed := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"}
    
    valid := false
    for _, allowedExt := range allowed {
        if ext == allowedExt {
            valid = true
            break
        }
    }
    
    if !valid {
        return fmt.Errorf("invalid file type. Allowed: jpg, jpeg, png, gif, webp, svg")
    }
    
    return nil
}