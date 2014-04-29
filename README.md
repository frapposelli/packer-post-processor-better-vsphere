# packer-post-processor-better-vsphere

This packer plugin is a dirty hack using @mitchellh code and leverages [VMware OVF Tool](http://www.vmware.com/support/developer/ovf) to upload a packer output to vSphere expanding {{ .BuildName }} variable in order to make it suitable for multi-builder packer files.


## Prerequisites

Software:

  * VMware OVF Tool
  
Notes:

  * This post processor only works with VMware

## Installation

Add

```
{
  "post-processors": {
    "better-vsphere": "packer-post-processor-better-vsphere"
  }
}
```

to your packer configuration (see: http://www.packer.io/docs/other/core-configuration.html -> Core Configuration)

Make sure that the directory which contains the packer-post-processor-better-vsphere executable is your PATH environmental variable (see http://www.packer.io/docs/extend/plugins.html -> Installing Plugins)

## Usage

The post-processor syntax is identical to the ```vsphere``` packer plugin (see: http://www.packer.io/docs/post-processors/vsphere.html) but you can use {{ .Buildname }} to append the name of your builder to the ```vm_name``` parameter (default for vm_name is "packer_{{ .BuildName }}").

```
  "post-processors": [
    {
        "type": "better-vsphere",
        "cluster": "myCluster",
        "datacenter": "myDC",
        "datastore": "myDatastore",
        "host": "myvCenter",
        "username": "myUsername",
        "password": "myPassword",
        "vm_folder": "/my/Folder",
        "resource_pool": "myResourcePool",
        "vm_network": "myNetwork",
        "vm_name": "{{ .BuildName }}_template",
        "insecure": "true"
    }
  ]
```
