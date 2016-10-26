package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lxc/lxd/shared"
)

// Loader functions
func containerKVMCreate(d *Daemon, args containerArgs) (container, error) {
	fmt.Println("Creating KVM container...")

	// Create the container struct
	c := &containerKVM{
		&containerLXC{
			daemon:       d,
			id:           args.Id,
			name:         args.Name,
			ephemeral:    args.Ephemeral,
			architecture: args.Architecture,
			cType:        args.Ctype,
			stateful:     args.Stateful,
			creationDate: args.CreationDate,
			lastUsedDate: args.LastUsedDate,
			profiles:     args.Profiles,
			localConfig:  args.Config,
			localDevices: args.Devices,
		},
	}

	// No need to detect storage here, its a new container.
	c.storage = d.Storage

	// Load the config
	err := c.init()
	if err != nil {
		c.Delete()
		return nil, err
	}

	// Look for a rootfs entry
	rootfs := false
	for _, name := range c.expandedDevices.DeviceNames() {
		m := c.expandedDevices[name]
		if m["type"] == "disk" && m["path"] == "/" {
			rootfs = true
			break
		}
	}

	if !rootfs {
		deviceName := "root"
		for {
			if c.expandedDevices[deviceName] == nil {
				break
			}

			deviceName += "_"
		}

		c.localDevices[deviceName] = shared.Device{"type": "disk", "path": "/"}

		updateArgs := containerArgs{
			Architecture: c.architecture,
			Config:       c.localConfig,
			Devices:      c.localDevices,
			Ephemeral:    c.ephemeral,
			Profiles:     c.profiles,
		}

		err = c.Update(updateArgs, false)
		if err != nil {
			c.Delete()
			return nil, err
		}
	}

	// Validate expanded config
	err = containerValidConfig(d, c.expandedConfig, false, true)
	if err != nil {
		c.Delete()
		return nil, err
	}

	err = containerValidDevices(c.expandedDevices, false, true)
	if err != nil {
		c.Delete()
		return nil, err
	}

	// Setup initial idmap config
	idmap := c.IdmapSet()
	var jsonIdmap string
	if idmap != nil {
		idmapBytes, err := json.Marshal(idmap.Idmap)
		if err != nil {
			c.Delete()
			return nil, err
		}
		jsonIdmap = string(idmapBytes)
	} else {
		jsonIdmap = "[]"
	}

	err = c.ConfigKeySet("volatile.last_state.idmap", jsonIdmap)
	if err != nil {
		c.Delete()
		return nil, err
	}

	// Update lease files
	networkUpdateStatic(d)

	return c, nil
}

func containerKVMLoad(d *Daemon, args containerArgs) (container, error) {
	fmt.Println("Loading KVM container...")

	// Create the container struct
	c := &containerKVM{
		&containerLXC{
			daemon:       d,
			id:           args.Id,
			name:         args.Name,
			ephemeral:    args.Ephemeral,
			architecture: args.Architecture,
			cType:        args.Ctype,
			creationDate: args.CreationDate,
			lastUsedDate: args.LastUsedDate,
			profiles:     args.Profiles,
			localConfig:  args.Config,
			localDevices: args.Devices,
			stateful:     args.Stateful,
		},
	}

	// Detect the storage backend
	s, err := storageForFilename(d, shared.VarPath("containers", strings.Split(c.name, "/")[0]))
	if err != nil {
		return nil, err
	}
	c.storage = s

	// Load the config
	err = c.init()
	if err != nil {
		return nil, err
	}

	return c, nil
}

// The KVM container driver
type containerKVM struct {
	*containerLXC
}

func (c *containerKVM) Start(stateful bool) error {
	fmt.Println("Hello, KVM here!")
	return c.containerLXC.Start(stateful)
}

func (c *containerKVM) Stop(stateful bool) error {
	fmt.Println("Stopping, KVM here!")
	return c.containerLXC.Stop(stateful)
}

func (c *containerKVM) Shutdown(timeout time.Duration) error {
	fmt.Println("Shutdown, KVM here!")
	return c.containerLXC.Shutdown(timeout)
}
