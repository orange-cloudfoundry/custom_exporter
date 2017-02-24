package collector

import (
	"github.com/orange-cloudfoundry/custom_exporter/custom_config"
	"fmt"
)

const (
	CollectorShell = "shell"
)

func NewShell(config custom_config.MetricsItem) (*CollectorCustom, error) {
	var myCol *CollectorCustom
	var err error

	if myCol, err = NewCollectorCustom(config); err != nil {
		return nil, err
	}

	if myCol.config.Credentials.Collector != CollectorShell {
		err := error(
			fmt.Printf("Error mismatching collector type : config type = %s & current type = %s",
				myCol.config.Credentials.Collector,
				CollectorShell,
		))
		return nil, err
	}

	if len(myCol.config.Commands) < 1 {
		err := error("Error empty commands to run !!")
		return nil, err
	}

	return myCol, nil
}

