package fs

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ZipFromDir 递归方式压缩目录下的所有文件和子目录。
// dir 是目录路径，output 是输出的zip文件路径，topDir 如果传非空字符串时，压缩包内的顶级目录使用此名称。
func ZipFromDir(output string, dir string, topDir string) error {
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
		return fmt.Errorf("the zip file cannot be created: %w", err)
	}
	defer outFile.Close()

	// 创建一个新的zip写入器
	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

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

	// 遍历目录并添加到zip中
	err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("traversal of the directory failed: %w", err)
		}

		// 跳过目录本身
		if path == absDir {
			return nil
		}

		// 计算相对于源目录的相对路径
		relPath, err := filepath.Rel(absDir, path)
		if err != nil {
			return fmt.Errorf("computing the relative path failed: %w", err)
		}

		// 构建zip中的文件路径
		var zipPath string
		if topDir != "" {
			zipPath = filepath.Join(basePath, relPath)
		} else {
			zipPath = relPath
		}

		// 如果是目录，则添加目录到zip中
		if info.IsDir() {
			// ZIP中的目录需要以 "/" 结尾
			zipPath = ensureTrailingSlash(zipPath)
			header, err1 := zip.FileInfoHeader(info)
			if err1 != nil {
				return fmt.Errorf("creating the zip directory header failed: %w", err1)
			}
			header.Name = zipPath
			header.Method = zip.Store // 目录通常不压缩

			// 创建zip目录条目
			writer, err2 := zipWriter.CreateHeader(header)
			if err2 != nil {
				return fmt.Errorf("creating the zip directory entry failed: %w", err2)
			}

			// 写入一个空的文件体以表示目录
			_, err3 := writer.Write([]byte{})
			if err3 != nil {
				return fmt.Errorf("writing to the zip directory entry failed: %w", err3)
			}

			return nil
		}

		// 打开文件
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("file %s cannot be opened: %w", path, err)
		}
		defer file.Close()

		// 获取文件信息
		fileInfo, err := file.Stat()
		if err != nil {
			return fmt.Errorf("failed to obtain file information: %w", err)
		}

		// 创建zip文件头
		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return fmt.Errorf("failed to create the zip file header: %w", err)
		}

		// 设置压缩方法
		header.Method = zip.Deflate

		// 设置文件在zip中的路径
		header.Name = zipPath

		// 创建zip文件写入器
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("creating the zip file entry failed: %w", err)
		}

		// 将文件内容复制到zip中
		_, err = io.Copy(writer, file)
		if err != nil {
			return fmt.Errorf("failed to write to the zip file: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("compressing the directory failed: %w", err)
	}

	return nil
}

// ZipFromFiles 压缩一个或多个文件。
// output 是输出的zip文件路径，files 是一个或多个文件路径。(无目录结构)
func ZipFromFiles(output string, files ...string) error {
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
		return fmt.Errorf("zip file cannot be created: %w", err)
	}
	defer outFile.Close()

	// 创建一个新的zip写入器
	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	for _, file := range files {
		err1 := addFileToZip(zipWriter, file)
		if err1 != nil {
			return fmt.Errorf("clamp file %s failed: %w", file, err1)
		}
	}

	return nil
}

// addFileToZip 辅助函数，用于将单个文件添加到zip写入器中
func addFileToZip(zipWriter *zip.Writer, file string) error {
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

	// 创建zip文件头
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("failed to create the zip file header: %w", err)
	}

	// 使用原始文件的修改时间
	// header.ModTime = info.ModTime()

	// 设置压缩方法
	header.Method = zip.Deflate

	// 创建zip文件写入器
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("creating the zip file entry failed: %w", err)
	}

	// 将文件内容复制到zip中
	_, err = io.Copy(writer, f)
	if err != nil {
		return fmt.Errorf("failed to write to the zip file: %w", err)
	}

	return nil
}

// ensureTrailingSlash 确保路径以斜杠结尾，用于目录
func ensureTrailingSlash(path string) string {
	if path[len(path)-1] == '/' || path[len(path)-1] == '\\' {
		return path
	}
	return path + "/"
}

// Unzip 解压缩ZIP文件。参数 filename 是zip文件路径；outputDir 是解压缩的目录路径。(当目录不存在时，自动创建)
func Unzip(filename string, outputDir string) error {
	// 打开zip文件
	reader, err := zip.OpenReader(filename)
	if err != nil {
		return fmt.Errorf("zip file cannot be opened: %w", err)
	}
	defer reader.Close()

	// 确保输出目录存在
	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("output directory cannot be created: %w", err)
	}

	for _, file := range reader.File {
		// 构建完整的文件路径
		filePath := filepath.Join(outputDir, file.Name)

		// 如果是目录，则创建目录
		if file.FileInfo().IsDir() {
			err = os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("directory %s cannot be created: %w", filePath, err)
			}
			continue
		}

		// 确保父目录存在
		parentDir := filepath.Dir(filePath)
		err = os.MkdirAll(parentDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("parent directory %s cannot be created: %w", parentDir, err)
		}

		// 创建文件
		outFile, err1 := os.Create(filePath)
		if err1 != nil {
			return fmt.Errorf("file %s cannot be created: %w", filePath, err1)
		}

		// 获取文件内容
		rc, err2 := file.Open()
		if err2 != nil {
			outFile.Close()
			return fmt.Errorf("file in zip cannot be opened %s: %w", file.Name, err2)
		}

		// 将内容复制到新文件中
		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return fmt.Errorf("failed to write to file %s: %w", filePath, err)
		}
	}

	return nil
}
