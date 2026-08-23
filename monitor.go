package config

import (
	"github.com/infrago/base"
	"github.com/infrago/infra"
)

func (c *Module) Ready() bool { return len(c.drivers) > 0 }

func (c *Module) Health() infra.ModuleHealth {
	return infra.NewModuleHealth("config", c.Ready(), nil, base.Map{"drivers": len(c.drivers)})
}

func (c *Module) Stats() infra.ModuleStats {
	return infra.NewModuleStats("config", c.Ready(), base.Map{"drivers": len(c.drivers)})
}
