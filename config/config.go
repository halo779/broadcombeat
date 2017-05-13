// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Config struct {
	Period time.Duration `config:"period"`
	Host   string        `config:"host"`
	User   string        `config:"user"`
	Pass   string        `config:"pass"`
}

var DefaultConfig = Config{
	Period: 1 * time.Second,
	Host:   "192.168.1.1:23",
	User:   "admin",
	Pass:   "admin",
}
