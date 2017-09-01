package qcache_inventory

import (
	"fmt"
	"time"
	"github.com/zpatrick/go-config"

	"github.com/qframe/types/constants"
	"github.com/qframe/types/docker-events"
	"github.com/qframe/types/qchannel"
	"github.com/qframe/types/plugin"
	"reflect"
	"strings"
)

const (
	version = "0.3.1"
	pluginTyp = qtypes_constants.CACHE
	pluginPkg = "inventory"
)

type Plugin struct {
	*qtypes_plugin.Plugin
	Inventory Inventory
}

func New(qChan qtypes_qchannel.QChan, cfg *config.Config, name string) (Plugin, error) {
	return Plugin{
		Plugin: qtypes_plugin.NewNamedPlugin(qChan, cfg, pluginTyp, pluginPkg, name, version),
		Inventory: NewInventory(),
	}, nil
}

// Run fetches everything from the Data channel and flushes it to stdout
func (p *Plugin) Run() {
	p.Log("notice", fmt.Sprintf("Start inventory v%s", p.Version))
	dc := p.QChan.Data.Join()
	tickerTime := p.CfgIntOr("ticker-ms", 500)
	ticker := time.NewTicker(time.Millisecond * time.Duration(tickerTime)).C
	for {
		select {
		case val := <-dc.Read:
			switch val.(type) {
			case qtypes_docker_events.ContainerEvent:
				ce := val.(qtypes_docker_events.ContainerEvent)
				if ce.Event.Type == "container" && ce.Event.Action == "start" {
					ips := []string{}
					for _,v := range ce.Container.NetworkSettings.Networks {
						ips = append(ips, v.IPAddress)
					}
					p.Log("debug", fmt.Sprintf("Add CntID:%s into Inventory (name:%s, IPs:%s)",ce.Container.ID[:13], ce.Container.Name, strings.Join(ips,",")))
					p.Inventory.SetItem(ce.Container.ID, ce.Container)
				}
			case ContainerRequest:
				req := val.(ContainerRequest)
				p.Log("debug", fmt.Sprintf("Received InventoryRequest for %v", req))
				p.Inventory.ServeRequest(req)
			default:
				p.Log("info", fmt.Sprintf("Dunno type '%s': %v", reflect.TypeOf(val), val))

			}
		case <- ticker:
			p.Log("trace", "Ticker came along: p.Inventory.CheckRequests()")
			p.Inventory.CheckRequests()
			continue
		}
	}
}
