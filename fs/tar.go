package fs

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// TarFromDir 压缩一个目录下的所有文件和目录。支持递归。
// dir 是目录路径，output 是输出的tar文件路径，topDir 如果传非空字符串时，压缩包内的顶级目录使用此名称。
func TarFromDir(output string, dir string, topDir string) error {
	if !IsDir(dir) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	outDir := filepath.Dir(output)
	if !IsDir(outDir) {
		err := os.MkdirAll(outDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create output directory %s: %w", outDir, err)
		}
	}

	// 创建输出文件
	outFile, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("cannot create file: %w", err)
	}
	defer outFile.Close()

	var (
		isGzip    bool
		tarWriter *tar.Writer
	)

	isGzip = strings.HasSuffix(output, ".tar.gz") || strings.HasSuffix(output, ".tgz")

	if isGzip {
		gzipWriter := gzip.NewWriter(outFile)
		defer gzipWriter.Close()

		tarWriter = tar.NewWriter(gzipWriter)
	} else {
		tarWriter = tar.NewWriter(outFile)
	}
	defer tarWriter.Close()

	// 获取目录的绝对路径
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("unable to get the absolute path to the directory: %w", err)
	}

	// 确定要使用的顶级目录名称
	var basePath string
	if topDir != "" {
		basePath = topDir
	} else {
		basePath = filepath.Base(absDir)
	}

	// 遍历目录并添加到tar中
	err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to traverse directory: %w", err)
		}

		// 跳过目录本身
		if path == absDir {
			return nil
		}

		// 计算相对于源目录的相对路径
		relPath, err := filepath.Rel(absDir, path)
		if err != nil {
			return fmt.Errorf("computing relative path failed: %w", err)
		}

		// 构建tar中的文件路径
		var tarPath string
		if topDir != "" {
			tarPath = filepath.Join(basePath, relPath)
		} else {
			tarPath = relPath
		}
		tarPath = strings.ReplaceAll(tarPath, "\\", "/")

		// 如果是目录，则添加目录到tar中
		if info.IsDir() {
			header, err1 := tar.FileInfoHeader(info, info.Name())
			if err1 != nil {
				return fmt.Errorf("creating the tar directory header failed: %w", err1)
			}
			header.Name = tarPath + "/"
			header.Typeflag = tar.TypeDir

			// 创建tar目录条目
			err = tarWriter.WriteHeader(header)
			if err != nil {
				return fmt.Errorf("writing to the tar directory entry failed: %w", err)
			}
			return nil
		}

		// 打开文件
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("unable to open file %s: %w", path, err)
		}
		defer file.Close()

		// 获取文件信息
		fileInfo, err := file.Stat()
		if err != nil {
			return fmt.Errorf("failed to obtain file information: %w", err)
		}

		// 创建tar文件头
		header, err := tar.FileInfoHeader(fileInfo, fileInfo.Name())
		if err != nil {
			return fmt.Errorf("failed to create the tar file header: %w", err)
		}

		// 设置文件在tar中的路径
		header.Name = tarPath

		// 创建tar文件写入器
		err = tarWriter.WriteHeader(header)
		if err != nil {
			return fmt.Errorf("writing to the tar file entry failed: %w", err)
		}

		// 将文件内容复制到tar中
		_, err = io.Copy(tarWriter, file)
		if err != nil {
			return fmt.Errorf("writing to the tar file failed: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("compressing the directory failed: %w", err)
	}

	return nil
}

// TarFromFiles 压缩一个或多个文件。
// output 是输出的tar文件路径；files 是一个或多个文件路径(文件都直接放在tar包下面，无目录结构)；
func TarFromFiles(output string, files ...string) error {
	outDir := filepath.Dir(output)
	if !IsDir(outDir) {
		err := os.MkdirAll(outDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create output directory %s: %w", outDir, err)
		}
	}

	var (
		isGzip    bool
		tarWriter *tar.Writer
	)

	isGzip = strings.HasSuffix(output, ".tar.gz") || strings.HasSuffix(output, ".tgz")

	// 创建输出文件
	outFile, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("tar file cannot be created: %w", err)
	}
	defer outFile.Close()

	if isGzip {
		gzipWriter := gzip.NewWriter(outFile)
		defer gzipWriter.Close()

		tarWriter = tar.NewWriter(gzipWriter)
	} else {
		tarWriter = tar.NewWriter(outFile)
	}

	defer tarWriter.Close()

	for _, file := range files {
		err = addFileToTar(tarWriter, file)
		if err != nil {
			return fmt.Errorf("add file %s failed: %w", file, err)
		}
	}

	return nil
}

// addFileToTar 辅助函数，用于将单个文件添加到tar写入器中
func addFileToTar(tarWriter *tar.Writer, file string) error {
	// 打开文件
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("file %s cannot be opened: %w", file, err)
	}
	defer f.Close()

	// 获取文件信息
	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to obtain file information: %w", err)
	}

	// 创建tar文件头
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return fmt.Errorf("failed to create the tar file header: %w", err)
	}

	// 创建tar文件写入器
	err = tarWriter.WriteHeader(header)
	if err != nil {
		return fmt.Errorf("creating the tar file entry failed: %w", err)
	}

	// 将文件内容复制到tar中
	_, err = io.Copy(tarWriter, f)
	if err != nil {
		return fmt.Errorf("writing to the tar file failed: %w", err)
	}

	return nil
}

// Untar 解压缩tar文件。参数 filename 是tar文件路径；outputDir 是解压缩的目录路径。(当目录不存在时，自动创建)
func Untar(filename string, outputDir string) error {
	// 打开tar文件
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("tar file cannot be opened: %w", err)
	}
	defer file.Close()

	var (
		isGzip    bool
		tarReader *tar.Reader
	)

	isGzip = strings.HasSuffix(filename, ".tar.gz") || strings.HasSuffix(filename, ".tgz")

	// 创建一个新的tar读取器
	if isGzip {
		gzipReader, err1 := gzip.NewReader(file)
		if err1 != nil {
			return fmt.Errorf("failed to create the gzip reader: %w", err1)
		}
		defer gzipReader.Close()

		tarReader = tar.NewReader(gzipReader)
	} else {
		tarReader = tar.NewReader(file)
	}

	for {
		header, err1 := tarReader.Next()

		// 到达文件末尾
		if err1 == io.EOF {
			break
		}

		if err1 != nil {
			return fmt.Errorf("reading tar file entry failed: %w", err1)
		}

		// 构建完整的文件路径
		filePath := filepath.Join(outputDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// 创建目录
			err = os.MkdirAll(filePath, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("unable to create directory %s: %w", filePath, err)
			}
		case tar.TypeReg:
			// 确保父目录存在
			parentDir := filepath.Dir(filePath)
			err = os.MkdirAll(parentDir, os.ModePerm)
			if err != nil {
				return fmt.Errorf("unable to create parent directory %s: %w", parentDir, err)
			}

			// 创建文件
			outFile, err2 := os.Create(filePath)
			if err2 != nil {
				return fmt.Errorf("unable to create file %s: %w", filePath, err2)
			}

			// 将tar中的内容写入文件
			_, err = io.Copy(outFile, tarReader)
			outFile.Close()
			if err != nil {
				return fmt.Errorf("failed to write to file %s: %w", filePath, err)
			}

			// 设置文件权限
			err = os.Chmod(filePath, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("unable to set file permissions %s: %w", filePath, err)
			}
		default:
			// 处理其他类型的文件（如符号链接、设备文件等）
			// 这里简单跳过不支持的类型
			fmt.Printf("Unsupported file types: %v - %s\n", header.Typeflag, header.Name)
		}
	}

	return nil
}
