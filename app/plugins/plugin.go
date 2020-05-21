package plugins

import (
	"github.com/cosasdepuma/elliot/app/plugins/portscanner"
	"github.com/cosasdepuma/elliot/app/plugins/robots"
	"github.com/cosasdepuma/elliot/app/plugins/subdomain"
)

var (
	// Plugins collects all available plugins that can be run
	Plugins = map[string]interface {
		Check() error
		Run()
	}{
		"portscanner": portscanner.Plugin{},
		"robots.txt":  robots.Plugin{},
		"subdomain":   subdomain.Plugin{},
	}
	// Required specifies what parameters are necessary to run a plugin
	Required = map[string]string{
		"portscanner": "Ports",
	}
	// Amount specifies how many plugins are available
	Amount = len(Plugins)
)
