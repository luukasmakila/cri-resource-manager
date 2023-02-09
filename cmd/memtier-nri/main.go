/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/sirupsen/logrus"

	"github.com/containerd/nri/pkg/api"
	"github.com/containerd/nri/pkg/stub"
)

type config struct {
	CfgParam1 string `json:"cfgParam1"`
}

type plugin struct {
	stub stub.Stub
	mask stub.EventMask
}

type MemtierdConfig struct {
	Policy Policy `yaml:"policy"`
}

type Policy struct {
	Name   string `yaml:"name"`
	Config string `yaml:"config"`
}

var (
	cfg config
	log *logrus.Logger
)

func (p *plugin) Configure(config, runtime, version string) (stub.EventMask, error) {
	log.Infof("Connected to %s/%s...", runtime, version)

	if config == "" {
		return 0, nil
	}

	err := yaml.Unmarshal([]byte(config), &cfg)
	if err != nil {
		return 0, fmt.Errorf("failed to parse configuration: %w", err)
	}

	log.Info("Got configuration data %+v...", cfg)

	return 0, nil
}

func (p *plugin) Synchronize(pods []*api.PodSandbox, containers []*api.Container) ([]*api.ContainerUpdate, error) {
	log.Info("Synchronizing state with the runtime...")
	return nil, nil
}

func (p *plugin) Shutdown() {
	log.Info("Runtime shutting down...")
}

func (p *plugin) RunPodSandbox(pod *api.PodSandbox) error {
	log.Infof("Started pod %s/%s...", pod.GetNamespace(), pod.GetName())
	return nil
}

func (p *plugin) StopPodSandbox(pod *api.PodSandbox) error {
	log.Infof("Stopped pod %s/%s...", pod.GetNamespace(), pod.GetName())
	return nil
}

func (p *plugin) RemovePodSandbox(pod *api.PodSandbox) error {
	log.Infof("Removed pod %s/%s...", pod.GetNamespace(), pod.GetName())
	return nil
}

func (p *plugin) CreateContainer(pod *api.PodSandbox, ctr *api.Container) (*api.ContainerAdjustment, []*api.ContainerUpdate, error) {
	log.Infof("Creating container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())

	//
	// This is the container creation request handler. Because the container
	// has not been created yet, this is the lifecycle event which allows you
	// the largest set of changes to the container's configuration, including
	// some of the later immautable parameters. Take a look at the adjustment
	// functions in pkg/api/adjustment.go to see the available controls.
	//
	// In addition to reconfiguring the container being created, you are also
	// allowed to update other existing containers. Take a look at the update
	// functions in pkg/api/update.go to see the available controls.
	//

	adjustment := &api.ContainerAdjustment{}
	updates := []*api.ContainerUpdate{}
	return adjustment, updates, nil
}

func (p *plugin) PostCreateContainer(pod *api.PodSandbox, ctr *api.Container) error {
	log.Infof("Created container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())
	return nil
}

func (p *plugin) StartContainer(pod *api.PodSandbox, ctr *api.Container) error {
	log.Infof("Starting container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())

	fullCgroupPath := getFullCgroupPath(ctr)
	addCgroupPathToConfig(fullCgroupPath)
	startMemtierd()

	return nil
}

func (p *plugin) PostStartContainer(pod *api.PodSandbox, ctr *api.Container) error {
	log.Infof("Started container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())
	return nil
}

func (p *plugin) UpdateContainer(pod *api.PodSandbox, ctr *api.Container) ([]*api.ContainerUpdate, error) {
	log.Infof("Updating container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())

	//
	// This is the container update request handler. You can make changes to
	// the container update before it is applied. Take a look at the functions
	// in pkg/api/update.go to see the available controls.
	//
	// In addition to altering the pending update itself, you are also allowed
	// to update other existing containers.
	//

	updates := []*api.ContainerUpdate{}

	return updates, nil
}

func (p *plugin) PostUpdateContainer(pod *api.PodSandbox, ctr *api.Container) error {
	log.Infof("Updated container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())
	return nil
}

func (p *plugin) StopContainer(pod *api.PodSandbox, ctr *api.Container) ([]*api.ContainerUpdate, error) {
	log.Infof("Stopped container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())

	//
	// This is the container (post-)stop request handler. You can update any
	// of the remaining running containers. Take a look at the functions in
	// pkg/api/update.go to see the available controls.
	//

	return []*api.ContainerUpdate{}, nil
}

func (p *plugin) RemoveContainer(pod *api.PodSandbox, ctr *api.Container) error {
	log.Infof("Removed container %s/%s/%s...", pod.GetNamespace(), pod.GetName(), ctr.GetName())
	return nil
}

func (p *plugin) onClose() {
	log.Infof("Connection to the runtime lost, exiting...")
	os.Exit(0)
}

func getFullCgroupPath(ctr *api.Container) []byte {
	cgroupPath := ctr.Linux.CgroupsPath

	split := strings.Split(cgroupPath, ":")

	partOne := split[0]
	partTwo := fmt.Sprintf("%s-%s.scope", split[1], split[2])

	partialPath := fmt.Sprintf("%s/%s", partOne, partTwo)

	fullPath := fmt.Sprintf("*/kubepods*/%s", partialPath)

	file, err := os.Open("/proc/mounts")
	if err != nil {
		log.Fatalf("failed to open /proc/mounts: %v", err)
	}
	defer file.Close()

	cgroupMountPoint := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if fields[0] == "cgroup2" {
			cgroupMountPoint = fields[1]
			break
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("failed to read /proc/mounts: %v", err)
	}

	findCmd := exec.Command("find", cgroupMountPoint, "-type", "d", "-wholename", fullPath)

	fullCgroupPath, err := findCmd.Output()
	if err != nil {
		log.Fatalf("failed to run find command: %v", err)
	}

	log.Printf("Cgroup path: %s", string(fullCgroupPath))

	return fullCgroupPath
}

func addCgroupPathToConfig(fullCgroupPath []byte) {
	yamlFile, err := ioutil.ReadFile("memtierd-config.yaml")
	if err != nil {
		log.Fatalf("Error reading YAML file: %v\n", err)
	}

	var memtierdConfig MemtierdConfig
	err = yaml.Unmarshal(yamlFile, &memtierdConfig)
	if err != nil {
		log.Fatalf("Error unmarshaling YAML: %v\n", err)
	}

	configFieldString := string(memtierdConfig.Policy.Config)

	fullCgroupPathString := string(fullCgroupPath)
	modifiedConfigYamlData := strings.Replace(configFieldString, "/sys/fs/cgroup/swapus", fullCgroupPathString, 1)

	memtierdConfig.Policy.Config = modifiedConfigYamlData

	out, err := yaml.Marshal(&memtierdConfig)
	if err != nil {
		log.Fatalf("Error marshaling YAML: %v\n", err)
	}

	err = ioutil.WriteFile("memtierd-config.yaml", out, 0644)
	if err != nil {
		log.Fatalf("Error writing YAML file: %v\n", err)
	}

	cat, err := exec.Command("cat", "memtierd-config.yaml").Output()
	if err != nil {
		log.Fatalf("Error writing YAML file: %v\n", err)
	}

	log.Infof("Yaml: %s", cat)

	log.Infof("YAML file successfully modified.")
}

func startMemtierd() {
	log.Infof("Starting Memtierd")

	out2, err := exec.Command("sh", "-c", "socat unix-listen:/tmp/memtierd.pod0c0.sock,fork,unlink-early - | memtierd -config memtierd-config.yaml >/tmp/memtierd.pod0c0.output 2>&1").Output()
	if err != nil {
		log.Fatalf("An error occurred: %s", err)
	}
	output2 := string(out2[:])
	log.Infof("Result of memtierd: %s", output2)
}

func main() {
	var (
		pluginName string
		pluginIdx  string
		err        error
	)

	log = logrus.StandardLogger()
	log.SetFormatter(&logrus.TextFormatter{
		PadLevelText: true,
	})

	flag.StringVar(&pluginName, "name", "", "plugin name to register to NRI")
	flag.StringVar(&pluginIdx, "idx", "", "plugin index to register to NRI")
	flag.Parse()

	p := &plugin{}
	opts := []stub.Option{
		stub.WithOnClose(p.onClose),
	}
	if pluginName != "" {
		opts = append(opts, stub.WithPluginName(pluginName))
	}
	if pluginIdx != "" {
		opts = append(opts, stub.WithPluginIdx(pluginIdx))
	}

	if p.stub, err = stub.New(p, opts...); err != nil {
		log.Fatalf("failed to create plugin stub: %v", err)
	}

	if err = p.stub.Run(context.Background()); err != nil {
		log.Errorf("plugin exited (%v)", err)
		os.Exit(1)
	}
}
