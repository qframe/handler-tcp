package qhandler_tcp

import (
	"fmt"
	"net"
	"reflect"
	"github.com/zpatrick/go-config"

	"github.com/qframe/types/qchannel"
	"github.com/qframe/types/plugin"
	"github.com/qframe/types/messages"
	"github.com/qframe/types/constants"
	"github.com/qframe/cache-inventory"
)

const (
	version   = "0.1.5"
	pluginTyp = qtypes_constants.HANDLER
	pluginPkg = "tcp-out"
)

type Plugin struct {
	*qtypes_plugin.Plugin
}

func New(qChan qtypes_qchannel.QChan, cfg *config.Config, name string) (Plugin, error) {
	p := Plugin{
		Plugin: qtypes_plugin.NewNamedPlugin(qChan, cfg, pluginTyp, pluginPkg, name, version),
	}
	p.Version = version
	p.Name = name
	return p, nil
}

// Run fetches everything from the Data channel and flushes it to stdout
func (p *Plugin) Run() {
	p.Log("notice", fmt.Sprintf("Start log handler v%s", p.Version))
	bg := p.QChan.Data.Join()
	host := p.CfgStringOr("host", "127.0.0.1")
	port := p.CfgStringOr("port", "10001")
	addr := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.Dial("tcp", addr)
	defer conn.Close()
	if err != nil {
		p.Log("error", err.Error())
	}
	p.Log("info", fmt.Sprintf("Connected to '%s'", addr))
	for {
		select {
		case val := <-bg.Read:
			switch val.(type) {
			case qtypes_messages.Message:
				qm := val.(qtypes_messages.Message)
				if qm.StopProcessing(p.Plugin, false) {
					continue
				}
				p.Log("debug", fmt.Sprintf("Sending '%s'", qm.Message))
				_, err := fmt.Fprintf(conn, qm.Message + "\n")
				if err != nil {
					p.Log("error", fmt.Sprintf("Error sending '%s': %s", qm.Message, err.Error()))
				}
			case qcache_inventory.ContainerRequest:
				continue
			default:
				p.Log("info" , fmt.Sprintf("Unkown type '%s': %v", reflect.TypeOf(val), val))
			}
		}
	}
}
