package main

import (
  "bytes"
  "fmt"
  "github.com/mitchellh/packer/common"
  "github.com/mitchellh/packer/packer"
  "os/exec"
  "strings"
)

var builtins = map[string]string{
  "mitchellh.vmware": "vmware",
}

type Config struct {
  common.PackerConfig `mapstructure:",squash"`

  Insecure     bool   `mapstructure:"insecure"`
  Cluster      string `mapstructure:"cluster"`
  Datacenter   string `mapstructure:"datacenter"`
  Datastore    string `mapstructure:"datastore"`
  Host         string `mapstructure:"host"`
  Password     string `mapstructure:"password"`
  ResourcePool string `mapstructure:"resource_pool"`
  Username     string `mapstructure:"username"`
  VMFolder     string `mapstructure:"vm_folder"`
  VMName       string `mapstructure:"vm_name"`
  VMNetwork    string `mapstructure:"vm_network"`

  tpl *packer.ConfigTemplate
}

type PostProcessor struct {
  configs map[string]*Config
}

func (p *PostProcessor) Configure(raws ...interface{}) error {

  p.configs = make(map[string]*Config)
  p.configs[""] = new(Config)
  if err := p.configureSingle(p.configs[""], raws...); err != nil {
    return err
  }

  return nil

}

func (p *PostProcessor) configureSingle(config *Config, raws ...interface{}) error {
  md, err := common.DecodeConfig(config, raws...)
  if err != nil {
    return err
  }

  config.tpl, err = packer.NewConfigTemplate()
  if err != nil {
    return err
  }
  config.tpl.UserVars = config.PackerUserVars

  // Defaults
  if config.VMName == "" {
    config.VMName = "packer_{{ .BuildName }}"
  }

  // Accumulate any errors
  errs := common.CheckUnusedConfig(md)

  validates := map[string]*string{
    "cluster":       &config.Cluster,
    "datacenter":    &config.Datacenter,
    "datastore":     &config.Datastore,
    "host":          &config.Host,
    "vm_network":    &config.VMNetwork,
    "password":      &config.Password,
    "resource_pool": &config.ResourcePool,
    "username":      &config.Username,
    "vm_folder":     &config.VMFolder,
    "vm_name":       &config.VMName,
  }

  for n, ptr := range validates {
    if err := config.tpl.Validate(*ptr); err != nil {
      errs = packer.MultiErrorAppend(
        errs, fmt.Errorf("Error parsing %s: %s", n, err))
    }
  }

  if _, err := exec.LookPath("ovftool"); err != nil {
    errs = packer.MultiErrorAppend(
      errs, fmt.Errorf("ovftool not found: %s", err))
  }

  if errs != nil && len(errs.Errors) > 0 {
    return errs
  }

  return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
  if _, ok := builtins[artifact.BuilderId()]; !ok {
    return nil, false, fmt.Errorf("Unknown artifact type, can't build box: %s", artifact.BuilderId())
  }

  config := p.configs[""]

  VMName, err := config.tpl.Process(config.VMName, &VMNameTemplate{
    ArtifactId: artifact.Id(),
    BuildName:  config.PackerBuildName,
  })
  if err != nil {
    return nil, false, err
  }

  vmx := ""
  for _, path := range artifact.Files() {
    if strings.HasSuffix(path, ".vmx") {
      vmx = path
      break
    }
  }

  if vmx == "" {
    return nil, false, fmt.Errorf("VMX file not found")
  }

  args := []string{
    fmt.Sprintf("--noSSLVerify=%t", config.Insecure),
    "--acceptAllEulas",
    fmt.Sprintf("--name=%s", VMName),
    fmt.Sprintf("--datastore=%s", config.Datastore),
    fmt.Sprintf("--network=%s", config.VMNetwork),
    fmt.Sprintf("--vmFolder=%s", config.VMFolder),
    fmt.Sprintf("%s", vmx),
    fmt.Sprintf("vi://%s:%s@%s/%s/host/%s/Resources/%s",
      config.Username,
      config.Password,
      config.Host,
      config.Datacenter,
      config.Cluster,
      config.ResourcePool),
  }

  ui.Message(fmt.Sprintf("Uploading %s to vSphere", vmx))
  var out bytes.Buffer
  cmd := exec.Command("ovftool", args...)
  cmd.Stdout = &out
  if err := cmd.Run(); err != nil {
    return nil, false, fmt.Errorf("Failed: %s\nStdout: %s", err, out.String())
  }

  ui.Message(fmt.Sprintf("%s", out.String()))

  return artifact, false, nil
}

type VMNameTemplate struct {
  ArtifactId string
  BuildName  string
}
