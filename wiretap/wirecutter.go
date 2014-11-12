package wiretap

type Config struct {
	Host       string
	Latency    int
	Bandwidth  int
	PacketLoss float64
}

type Wiretap interface {
	Setup(*Config) error
	Teardown() error
	Exists() bool
	Check() string
}
