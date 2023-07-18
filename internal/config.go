package internal

type Config struct {
	ServerAddress  string `env:"ADDRESS"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
}
