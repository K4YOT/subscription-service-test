package main

import (
	"log/slog"
	"os"
)

func main() {
	// Настройка логов: вывод в консоль в читаемом виде
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("Starting subscription service")
	// ... остальной код
}
