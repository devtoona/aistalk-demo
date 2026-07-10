package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	fileLogger *log.Logger
	file       *os.File
)

// InitLogger ログファイルを初期化します。
// Cloud Run（K_SERVICE あり）では stdout のみ。ファイル書き込み失敗で起動しないようにする。
func InitLogger(logDir string) error {
	if os.Getenv("K_SERVICE") != "" {
		return nil
	}

	// ログディレクトリが存在しない場合は作成
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// ログファイル名を日付で生成
	logFileName := fmt.Sprintf("dev_%s.log", time.Now().Format("2006-01-02"))
	logFilePath := filepath.Join(logDir, logFileName)

	// ログファイルを開く（追記モード）
	var err error
	file, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// ファイルロガーを初期化
	fileLogger = log.New(file, "", log.LstdFlags)

	return nil
}

// Close ログファイルを閉じます
func Close() {
	if file != nil {
		file.Close()
	}
}

// Info 情報ログを出力します
func Info(format string, v ...interface{}) {
	message := fmt.Sprintf("[INFO] "+format, v...)

	// コンソールに出力
	log.Printf(message)

	// ファイルに出力
	if fileLogger != nil {
		fileLogger.Printf(message)
	}
}

// Error エラーログを出力します
func Error(format string, v ...interface{}) {
	message := fmt.Sprintf("[ERROR] "+format, v...)

	// コンソールに出力
	log.Printf(message)

	// ファイルに出力
	if fileLogger != nil {
		fileLogger.Printf(message)
	}
}

// Debug デバッグログを出力します
func Debug(format string, v ...interface{}) {
	message := fmt.Sprintf("[DEBUG] "+format, v...)

	// コンソールに出力
	log.Printf(message)

	// ファイルに出力
	if fileLogger != nil {
		fileLogger.Printf(message)
	}
}

// Warn 警告ログを出力します
func Warn(format string, v ...interface{}) {
	message := fmt.Sprintf("[WARN] "+format, v...)

	// コンソールに出力
	log.Printf(message)

	// ファイルに出力
	if fileLogger != nil {
		fileLogger.Printf(message)
	}
}

// Fatal 致命的エラーログを出力してプログラムを終了します
func Fatal(format string, v ...interface{}) {
	message := fmt.Sprintf("[FATAL] "+format, v...)

	// コンソールに出力
	log.Printf(message)

	// ファイルに出力
	if fileLogger != nil {
		fileLogger.Printf(message)
	}

	// ログファイルを閉じる
	Close()

	// プログラムを終了
	os.Exit(1)
}
