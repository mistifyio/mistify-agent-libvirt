PREFIX := /opt/mistify
SBIN_DIR=$(PREFIX)/sbin
SV_DIR=$(PREFIX)/sv
ETC_DIR=$(PREFIX)/etc

cmd/mistify-libvirt/mistify-libvirt: cmd/mistify-libvirt/main.go
	cd cmd/mistify-libvirt && \
	go get && \
	go build

clean:
	cd cmd/mistify-agent && \
	go clean

install: cmd/mistify-libvirt/mistify-libvirt
	mkdir -p $(DESTDIR)${SBIN_DIR}
	mkdir -p $(DESTDIR)${SV_DIR}

	install -D cmd/mistify-libvirt/mistify-libvirt $(DESTDIR)${SBIN_DIR}/mistify-libvirt
