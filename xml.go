package libvirt

import (
	"bytes"
	"github.com/mistifyio/mistify-agent/client"
	"text/template"
)

func (lv *Libvirt) DomainXML(guest *client.Guest) (string, error) {
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

	buf := new(bytes.Buffer)
	err := tmpl.Execute(buf, guest)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
