package libvirt

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/mistifyio/mistify-agent/client"
)

func (lv *Libvirt) DomainXML(guest *client.Guest) (string, error) {
	fmt.Println(guest)

	const xml = `
<domain type="{{.Type}}">
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
    <interface type="bridge">
      <guest dev="{{.Name}}" />
      {{if .Mac}}<mac address="{{.Mac}}" />{{end}}
      <source bridge='{{.Network}}'/>
      <target dev="{{.Device}}" />
      {{if .Model}}<model type="{{.Model}}" />{{end}}
    </interface>
    {{end}}

    {{range .Disks}}
    <disk type="block" device="disk">
      <driver name="qemu" type="raw" />
      <source dev="{{.Source}}" />
      <target dev="{{.Device}}" bus="{{.Bus}}" />
    </disk>
    {{end}}
  </devices>
</domain>
`

	tmpl := template.New("xml")
	template.Must(tmpl.Parse(xml))

	buf := new(bytes.Buffer)
	err := tmpl.Execute(buf, guest)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
