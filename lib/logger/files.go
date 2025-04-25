package logger

import "os"

// 实现文件管理，检查日志目录是否存在，是否有权限写入，创建目录等功能
// check if the user has permission to write to it
func checkPermissions(dir string) bool {
	_, err := os.Stat(dir)
	return os.IsPermission(err)
}

// create a directory if it does not exist
func isNotExistMkdir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func mustOpen(fileName, dir string) (*os.File, error) {
	if checkPermissions(dir) {
		return nil, os.ErrPermission
	}
	if err := isNotExistMkdir(dir); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(dir+string(os.PathSeparator)+fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}
