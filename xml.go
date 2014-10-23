package libvirt

import (
	"text/template"
	"bytes"
	"github.com/mistifyio/mistify-agent/client"
	"strings"
)

func (lv *Libvirt) DomainXML(guest *client.Guest) string {
	const xml = `
<domain type="kvm">
  <name>{{.Id}}</name>
  <memory unit="MiB">{{.Memory}}</memory>
  <vcpu>{{.Cpu}}</vcpu>

  {{if .Metadata}}
  <metadata>
    {{range .Metadata}}
    {{end}}
  </metadata>
  {{end}}

  <os>
    <type>hvm</type>
  </os>
  <devices>
    {{range .Nics}}
    <interface type="network">
      <guest dev="{{.Name}}" />
      <mac address="{{.Mac}}" />
      {{if .Model}}<model type="{{.Model}}" />{{end}}
    </interface>
    {{end}}

    {{range .Disks}}
    <disk type="file" device="disk">
      <source file="{{.Source}}" />
      <target dev="{{.Device}}" bus="{{.Bus}}" />
    </disk>
    {{end}}
  </devices>
</domain>
`

	tmpl := template.New("xml")
	template.Must(tmpl.Parse(xml))

	buf := make([]byte, 1 << 12)
	io := bytes.NewBuffer(buf)

	err := tmpl.Execute(io, guest)
	if err != nil {
		panic(err)
	}

	str := strings.Trim(io.String(), "\x00")
	return str
}
