// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package global

type (
	// Config is a global antibuild configuration
	Config struct {
		Repositories []string `json:"repositories"`
	}
)

// DefaultGlobalConfig that gets auto opened
var DefaultGlobalConfig *Config
