package internal

type Config struct {
	ServerAddress   string `env:"ADDRESS"`
	PollInterval    int64  `env:"POLL_INTERVAL"`
	ReportInterval  int64  `env:"REPORT_INTERVAL"`
	LogLevel        string `env:"LOG_LEVEL"`
	StoreInterval   int64  `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}
