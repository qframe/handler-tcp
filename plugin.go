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
	"github.com/qframe/types/syslog"
	"io"
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
	return p, nil
}

// Run fetches everything from the Data channel and flushes it to stdout
func (p *Plugin) Run() {
	p.Log("notice", fmt.Sprintf("Start %s-%s '%s' v%s", p.Typ, p.Pkg, p.Name, p.Version))
	bg := p.QChan.Data.Join()
	host := p.CfgStringOr("host", "127.0.0.1")
	port := p.CfgStringOr("port", "10001")
	isSyslog := p.CfgBoolOr("syslog5424", false)
	setSyslogCeeKey := p.CfgStringOr("syslog5424-cee-key", "")
	ignoreContainerEvents := p.CfgBoolOr("ignore-container-events", true)
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
				if isSyslog {
					err := p.sendSyslog(conn, qm.Tags, setSyslogCeeKey)
					if err != nil {
						p.Log("error", err.Error())
					}
					continue
				}
				b, err := fmt.Fprintf(conn, qm.Message + "\n")
				if err != nil {
					p.Log("error", fmt.Sprintf("Error sending '%s': %s", qm.Message, err.Error()))
				} else {
					p.Log("debug", fmt.Sprintf("Send %d bytes: '%s'", b, qm.Message))
				}
			case qtypes_messages.ContainerMessage:
				qm := val.(qtypes_messages.ContainerMessage)
				if qm.StopProcessing(p.Plugin, false) {
					continue
				}
				if isSyslog {
					err := p.sendSyslog(conn, qm.Tags, setSyslogCeeKey)
					if err != nil {
						p.Log("error", err.Error())
					}
					continue
				}
				b, err := fmt.Fprintf(conn, qm.Message.Message + "\n")
				if err != nil {
					p.Log("error", fmt.Sprintf("Error sending '%s': %s", qm.Message, err.Error()))
				} else {
					p.Log("debug", fmt.Sprintf("Send %d bytes: '%s'", b, qm.Message.Message))
				}
			default:
				if ignoreContainerEvents {
					continue
				}
				p.Log("debug" , fmt.Sprintf("Unkown type '%s': %v", reflect.TypeOf(val), val))
			}

		}
	}
}

func (p *Plugin) sendSyslog(conn io.Writer, kv map[string]string, setSyslogCeeKey string ) (err error) {
	isCee := false
	if setSyslogCeeKey != "" {
		if _, ok := kv[setSyslogCeeKey]; !ok {
			err = fmt.Errorf("could not find 'setSyslogMsgKey' '%s' in kv", setSyslogCeeKey)
		} else {
			isCee = true
			kv[qtypes_syslog.KEY_MSG] = kv[setSyslogCeeKey]
			p.Log("debug", fmt.Sprintf("Overwrite KY_MSG '%s' with '%s'", qtypes_syslog.KEY_MSG, setSyslogCeeKey))
		}

	}
	sl, err := qtypes_syslog.NewSyslogFromKV(kv)
	if err != nil {
		return fmt.Errorf("error parsing Tags to syslog5424 struct: %s", err.Error())
	}
	if isCee {
		p.Log("debug", "Enable CEE prefix to message")
		sl.EnableCEE()
	}
	msg, err := sl.ToRFC5424()
	if err != nil {
		return fmt.Errorf("error creating syslog5424 message: %s", err.Error())
	}
	p.Log("debug", fmt.Sprintf("Sending '%s'", msg))
	_, err = fmt.Fprintf(conn, msg + "\n")
	if err != nil {
		return fmt.Errorf("rrror sending '%s': %s", msg, err.Error())
	}
	return

}
