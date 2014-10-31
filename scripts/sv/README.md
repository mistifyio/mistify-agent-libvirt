Example runit scripts and config.

To install (assuming binaries are in /opt/mistify/sbin):

```
install -d -m 0644 libvirt.json /etc/mistify-libvirt/libvirt.json
install -d -m 0755 run /etc/sv/mistify-libvirt/run
install -d -m 0755 log /etc/sv/mistify-libvirt/log/run
ln -sf /etc/sv/mistify-libvirt /etc/service
```
