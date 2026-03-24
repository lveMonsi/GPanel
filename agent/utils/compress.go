package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type CompressType string

const (
	Zip   CompressType = "zip"
	TarGz CompressType = "tar.gz"
)

func (f FileOp) CompressFiles(files []string, dst, name string, cType CompressType) error {
	if !f.Stat(dst) {
		if err := f.CreateDir(dst, 0755); err != nil {
			return fmt.Errorf("failed to create destination directory: %v", err)
		}
	}
	dstPath := filepath.Join(dst, name)
	switch cType {
	case Zip:
		return f.compressToZip(files, dstPath)
	case TarGz:
		return f.compressToTarGz(files, dstPath)
	default:
		return fmt.Errorf("unsupported compress type: %s", cType)
	}
}

func (f FileOp) compressToZip(files []string, dstPath string) error {
	zipFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %v", err)
	}
	defer zipFile.Close()
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	for _, file := range files {
		if err := filepath.Walk(file, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			relPath, err := filepath.Rel(filepath.Dir(file), filePath)
			if err != nil {
				return err
			}
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			header.Name = relPath
			header.Method = zip.Deflate
			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return err
			}
			if !info.IsDir() {
				srcFile, err := os.Open(filePath)
				if err != nil {
					return err
				}
				defer srcFile.Close()
				if _, err := io.Copy(writer, srcFile); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return fmt.Errorf("failed to walk file %s: %v", file, err)
		}
	}
	return nil
}

func (f FileOp) compressToTarGz(files []string, dstPath string) error {
	tarGzFile, err := f.Fs.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create tar.gz file: %v", err)
	}
	defer tarGzFile.Close()
	gzWriter := gzip.NewWriter(tarGzFile)
	defer gzWriter.Close()
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()
	for _, file := range files {
		if err := filepath.Walk(file, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			relPath, err := filepath.Rel(filepath.Dir(file), filePath)
			if err != nil {
				return err
			}
			header.Name = relPath
			if err := tarWriter.WriteHeader(header); err != nil {
				return err
			}
			if !info.IsDir() {
				file, err := os.Open(filePath)
				if err != nil {
					return err
				}
				defer file.Close()
				if _, err := io.Copy(tarWriter, file); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return fmt.Errorf("failed to walk file %s: %v", file, err)
		}
	}
	return nil
}

func (f FileOp) DecompressFile(srcFile, dst string, cType CompressType) error {
	if !f.Stat(dst) {
		if err := f.CreateDir(dst, 0755); err != nil {
			return fmt.Errorf("failed to create destination directory: %v", err)
		}
	}
	if !f.Stat(srcFile) {
		return fmt.Errorf("source file not found: %s", srcFile)
	}
	switch cType {
	case Zip:
		return f.decompressFromZip(srcFile, dst)
	case TarGz:
		return f.decompressFromTarGz(srcFile, dst)
	default:
		return fmt.Errorf("unsupported compress type: %s", cType)
	}
}

func (f FileOp) decompressFromZip(srcFile, dst string) error {
	zipReader, err := zip.OpenReader(srcFile)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer zipReader.Close()
	for _, file := range zipReader.File {
		filePath := filepath.Join(dst, file.Name)
		if file.FileInfo().IsDir() {
			if err := f.CreateDir(filePath, file.Mode()); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", filePath, err)
			}
			continue
		}
		if err := f.CreateDir(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %v", err)
		}
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %v", err)
		}
		destFile, err := f.Fs.Create(filePath)
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create file %s: %v", filePath, err)
		}
		if _, err := io.Copy(destFile, rc); err != nil {
			destFile.Close()
			rc.Close()
			return fmt.Errorf("failed to write file %s: %v", filePath, err)
		}
		destFile.Close()
		rc.Close()
		if err := f.Fs.Chmod(filePath, file.Mode()); err != nil {
			return fmt.Errorf("failed to set file permissions: %v", err)
		}
	}
	return nil
}

func (f FileOp) decompressFromTarGz(srcFile, dst string) error {
	file, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz file: %v", err)
	}
	defer file.Close()
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()
	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %v", err)
		}
		targetPath := filepath.Join(dst, header.Name)
		if header.Typeflag == tar.TypeDir {
			if err := f.CreateDir(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", targetPath, err)
			}
			continue
		}
		if err := f.CreateDir(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %v", err)
		}
		outFile, err := f.Fs.Create(targetPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %v", targetPath, err)
		}
		if _, err := io.Copy(outFile, tarReader); err != nil {
			outFile.Close()
			return fmt.Errorf("failed to write file %s: %v", targetPath, err)
		}
		outFile.Close()
		if err := f.Fs.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
			return fmt.Errorf("failed to set file permissions: %v", err)
		}
	}
	return nil
}

func DetectCompressType(filename string) CompressType {
	lower := strings.ToLower(filename)
	if strings.HasSuffix(lower, ".zip") {
		return Zip
	}
	if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") {
		return TarGz
	}
	return ""
}
