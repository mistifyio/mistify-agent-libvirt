package libvirt

import (
	"bytes"
	"text/template"

	"github.com/mistifyio/mistify-agent/client"
)

var domainTemplate *template.Template
var networkTemplate *template.Template

func init() {
	const domainXML = `
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
    <interface type="network">
      <source network='{{.Mac}}' portgroup='vlan-all' />
	  {{if .Name}}<guest dev="{{.Name}}" />{{end}}
      {{if .Mac}}<mac address="{{.Mac}}" />{{end}}
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
	domainTemplate = template.Must(template.New("domainXML").Parse(domainXML))

	const networkXML = `
<network>
	<name>{{.Mac}}</name>
	<forward mode='bridge' />
	<bridge name='{{.Network}}' />
	<virtualport type='openvswitch' />
	<portgroup name='vlan-all'>
		<vlan>
		{{range .VLANs}}
			<tag id ='{{.}}' />
		{{end}}
		</vlan>
	</portgroup>
</network>
`
	networkTemplate = template.Must(template.New("networkXML").Parse(networkXML))
}

// DomainXML populates a libvirt domain xml template with guest properties
func (lv *Libvirt) DomainXML(guest *client.Guest) (string, error) {
	buf := new(bytes.Buffer)
	err := domainTemplate.Execute(buf, guest)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// NetworkXML populates a libvirt network xml template with guest properties
func (lv *Libvirt) NetworkXML(nic client.Nic) (string, error) {
	buf := new(bytes.Buffer)
	err := networkTemplate.Execute(buf, nic)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
