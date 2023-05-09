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
	"syscall"

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
	Policy   Policy     `yaml:"policy"`
	Routines []Routines `yaml:"routines"`
}

type Routines struct {
	Name   string `yaml:"name"`
	Config string `yaml:"config"`
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

	podName := pod.GetName()
	containerName := ctr.GetName()
	annotations := pod.GetAnnotations()
	template := ""

	// If memtierd annotation is not present, don't execute further
	priority, ok := annotations["class.memtierd.nri"]
	if !ok {
		return nil
	}

	if priority == "low-prio" {
		template = "low-prio.yaml"
	}
	if priority == "high-prio" {
		template = "high-prio.yaml"
	}

	fullCgroupPath := getFullCgroupPath(ctr)
	podDirectory, outputFilePath, configFilePath := editMemtierdConfig(fullCgroupPath, podName, containerName, template)
	startMemtierd(podName, containerName, podDirectory, outputFilePath, configFilePath)

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

	podName := pod.GetName()
	dirPath := fmt.Sprintf("/tmp/memtierd/%s", podName)

	// Kill the memtierd process
	out, err := exec.Command("sudo", "pkill", "-f", dirPath).CombinedOutput()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok || exitErr.ExitCode() != 1 {
			// Error occurred that is not related to "no processes found"
			log.Fatalf("Error killing memtierd process: %v. Output: %s\n", err, out)
		} else {
			// "No processes found" error, do nothing
			log.Printf("No processes found for memtierd process\n")
		}
	}

	err = os.RemoveAll(dirPath)
	if err != nil {
		fmt.Println(err)
	}

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

func editMemtierdConfig(fullCgroupPath []byte, podName string, containerName string, template string) (string, string, string) {
	templatePath := fmt.Sprintf("/host/home/ubuntu/templates/%s", template)
	yamlFile, err := ioutil.ReadFile(templatePath)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v\n", err)
	}

	// Create directory if it doesn't exist
	podDirectory := fmt.Sprintf("/tmp/memtierd/%s", podName)
	if err := os.MkdirAll("/host"+podDirectory, 0755); err != nil {
		log.Fatalf("Error creating directory: %v", err)
	}

	outputFilePath := fmt.Sprintf("%s/memtierd.%s.output", podDirectory, containerName)

	var memtierdConfig MemtierdConfig
	err = yaml.Unmarshal(yamlFile, &memtierdConfig)
	if err != nil {
		log.Fatalf("Error unmarshaling YAML: %v\n", err)
	}

	fullCgroupPathString := string(fullCgroupPath)

	// Edit the Policy and Routine configs
	policyConfigFieldString := string(memtierdConfig.Policy.Config)
	policyConfigFieldString = strings.Replace(policyConfigFieldString, "$CGROUP2_ABS_PATH", fullCgroupPathString, 1)

	// Loop through the routines
	for i := 0; i < len(memtierdConfig.Routines); i++ {
		routineConfigFieldString := string(memtierdConfig.Routines[i].Config)
		routineConfigFieldString = strings.Replace(routineConfigFieldString, "$MEMTIERD_SWAP_STATS_PATH", outputFilePath, 1)
		memtierdConfig.Routines[i].Config = routineConfigFieldString
	}

	memtierdConfig.Policy.Config = policyConfigFieldString

	out, err := yaml.Marshal(&memtierdConfig)
	if err != nil {
		log.Fatalf("Error marshaling YAML: %v\n", err)
	}

	configFilePath := fmt.Sprintf("/host"+podDirectory+"/%s.yaml", containerName)
	err = ioutil.WriteFile(configFilePath, out, 0644)
	if err != nil {
		log.Fatalf("Error writing YAML file: %v\n", err)
	}

	cat, err := exec.Command("cat", configFilePath).Output()
	if err != nil {
		log.Fatalf("Error writing YAML file: %v\n", err)
	}

	log.Infof("Yaml: %s", cat)

	log.Infof("YAML file successfully modified.")
	return podDirectory, outputFilePath, configFilePath
}

func startMemtierd(podName string, containerName string, podDirectory string, outputFilePath string, configFilePath string) {
	log.Infof("Starting Memtierd")

	// Open the output file for writing
	outputFile, err := os.OpenFile("/host"+outputFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Failed to open output file: %v\n", err)
	}
	//defer outputFile.Close()

	socketPath := fmt.Sprintf("/host"+podDirectory+"/memtierd.%s.sock", containerName)

	file, err := os.Create(socketPath)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	log.Infof(socketPath)
	log.Infof(configFilePath)

	// Create the command and write its output to the output file
	socatCommand := fmt.Sprintf("socat unix-listen:%s,fork,unlink-early - | memtierd -config %s", socketPath, configFilePath)
	cmd := exec.Command("sh", "-c", socatCommand)
	cmd.Stdout = outputFile
	cmd.Stderr = outputFile

	// Start the command in a new session and process group
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	// Start the command in the background
	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start command: %v\n", err)
	}
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
