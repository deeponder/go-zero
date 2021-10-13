package rest

import (
	"time"

	"gitlab.deepwisdomai.com/infra/go-zero/core/service"
)

type (
	// A PrivateKeyConf is a private key config.
	PrivateKeyConf struct {
		Fingerprint string
		KeyFile     string
	}

	// A SignatureConf is a signature config.
	SignatureConf struct {
		Strict      bool          `json:",default=false"`
		Expiry      time.Duration `json:",default=1h"`
		PrivateKeys []PrivateKeyConf
	}

	NacosConf struct {
		UseNacos bool `json:",default=false"`
		Ip       string
		Port     uint64

		TimeoutMs           uint64 `json:",default=5000"`
		NotLoadCacheAtStart bool   `json:",default=true"`
		LogDir              string `json:",default="`
		CacheDir            string `json:",default=/tmp/nacos/cache"`
		RotateTime          string `json:",default=1h"`
		MaxAge              int64  `json:",default=3"`
		LogLevel            string `json:",default=debug"`
	}

	// A RestConf is a http service config.
	// Why not name it as Conf, because we need to consider usage like:
	//  type Config struct {
	//     zrpc.RpcConf
	//     rest.RestConf
	//  }
	// if with the name Conf, there will be two Conf inside Config.
	RestConf struct {
		service.ServiceConf
		Host     string `json:",default=0.0.0.0"`
		Port     int
		CertFile string `json:",optional"`
		KeyFile  string `json:",optional"`
		Verbose  bool   `json:",optional"`
		MaxConns int    `json:",default=10000"`
		MaxBytes int64  `json:",default=1048576,range=[0:33554432]"`
		// milliseconds
		Timeout      int64         `json:",default=3000"`
		CpuThreshold int64         `json:",default=900,range=[0:1000]"`
		Signature    SignatureConf `json:",optional"`
		NacosConf    NacosConf     `json:",optional"`
	}
)
