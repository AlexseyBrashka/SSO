package tls

type Config struct {
	CertFile string
	KeyFile  string
}

func NewConfig(certFile string, keyFile string) *Config {
	return &Config{
		CertFile: certFile,
		KeyFile:  keyFile,
	}
}
