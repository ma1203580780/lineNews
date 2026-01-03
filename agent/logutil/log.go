package logutil

import (
	"fmt"
	"log"
	"runtime"
	"strings"
)

// LogInfo 记录信息级别日志，包含文件名、函数名和行号
func LogInfo(format string, v ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}

	// 获取函数名
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	funcName := "unknown"
	if fn != nil {
		funcName = fn.Name()
		// 简化函数名，只保留包名.函数名
		parts := strings.Split(funcName, "/")
		if len(parts) > 0 {
			funcName = parts[len(parts)-1]
		}
	}

	// 获取文件名
	shortFile := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' || file[i] == '\\' {
			shortFile = file[i+1:]
			break
		}
	}

	log.Printf("[%s:%d %s] %s", shortFile, line, funcName, fmt.Sprintf(format, v...))
}

// LogError 记录错误级别日志，包含文件名、函数名和行号
func LogError(format string, v ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}

	// 获取函数名
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	funcName := "unknown"
	if fn != nil {
		funcName = fn.Name()
		// 简化函数名，只保留包名.函数名
		parts := strings.Split(funcName, "/")
		if len(parts) > 0 {
			funcName = parts[len(parts)-1]
		}
	}

	// 获取文件名
	shortFile := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' || file[i] == '\\' {
			shortFile = file[i+1:]
			break
		}
	}

	log.Printf("[ERROR %s:%d %s] %s", shortFile, line, funcName, fmt.Sprintf(format, v...))
}

// LogDebug 记录调试级别日志，包含文件名、函数名和行号
func LogDebug(format string, v ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}

	// 获取函数名
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	funcName := "unknown"
	if fn != nil {
		funcName = fn.Name()
		// 简化函数名，只保留包名.函数名
		parts := strings.Split(funcName, "/")
		if len(parts) > 0 {
			funcName = parts[len(parts)-1]
		}
	}

	// 获取文件名
	shortFile := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' || file[i] == '\\' {
			shortFile = file[i+1:]
			break
		}
	}

	log.Printf("[DEBUG %s:%d %s] %s", shortFile, line, funcName, fmt.Sprintf(format, v...))
}
