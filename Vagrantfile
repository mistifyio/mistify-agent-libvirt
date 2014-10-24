# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.define :libvirt_vm do |vm_config|
    vm_config.vm.box = "mistify/trusty64-vmware"
    vm_config.vm.box_url = "http://www.akins.org/boxes/mistify-ubuntu-vmware.box"
    vm_config.vm.synced_folder "../../../..", "/data/go", create: true

	vm_config.vm.provider "vmware_workstation" do |v|
      v.vmx["memsize"] = "1024"
      v.vmx["numvcpus"] = "2"
      v.vmx["vhv.enable"] = "TRUE"
      v.name = "lv001"
    end

    vm_config.vm.provision "shell", inline: <<-EOB
      apt-get -y update
      apt-get -y upgrade
      apt-get -y install tmux git mercurial gnulib python-dev libxml2-utils libxml2-dev xsltproc libdevmapper-dev \
                         libpciaccess-dev libnl-dev uuid-dev libdbus-1-dev libyajl-dev ubuntu-zfs libzfs-dev

      mkdir /scratch
      for i in {1..3}; do truncate -s 2G /scratch/$i.img; done
      zpool create zpool raidz1 /scratch/1.img /scratch/2.img /scratch/3.img

      cd /tmp
      git clone git://libvirt.org/libvirt.git
      cd /tmp/libvirt
      ./autogen.sh --with-storage-zfs
      make
      make install || true

      if [ ! -e /data ]; then
        mkdir /data
      fi

      if [ ! -e /data/go1.3.linux-amd64.tar.gz ]; then
        cd /data
        wget --quiet http://golang.org/dl/go1.3.linux-amd64.tar.gz
        tar -C /usr/local -xzf go1.3.linux-amd64.tar.gz
      fi

      export PATH=/usr/local/go/bin:$PATH
      export GOPATH=/data/go
      go get github.com/tools/godep

      chmod -R g+w /data
      chgrp -R vagrant /data

      cat > /etc/environment <<EOF
PATH=$PATH
GOPATH=$GOPATH
EOF

      /usr/local/sbin/libvirtd -d
      virsh net-define /tmp/libvirt/src/network/default.xml
      virsh net-start default
    EOB
  end
end
