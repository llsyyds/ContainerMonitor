package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/client"
	"io"
	"log"
	"strconv"
	"time"
)

type TContainerMonitor struct {
	Id     string            // Container ID
	Name   string            // Container Name
	Labels map[string]string // Container labels (run-time)
	cli    *client.Client    // Docker Client

	stop bool // thread control flag

	// Callback methods
	OnStatRead TClbOnStatistic
	OnRemove   TClbOnRemove
}

func (m *TContainerMonitor) SetOpt(opt TOpt) error {
	return errors.New(fmt.Sprintf("Unknown option: %s", opt.Name))
}

func (m *TContainerMonitor) GetOpt(name string) *TOpt {
	switch name {
	case "name":
		return &TOpt{
			Name:  "name",
			Value: m.Name,
		}
	case "labels":
		return &TOpt{
			Name:  "labels",
			Value: m.Labels,
		}
	}
	m.Id = strconv.Itoa(0)
	m.Name = ""

	return nil
}

func (m *TContainerMonitor) Exec() error {
	if er := m.init(); er != nil {
		return er
	}

	m.stop = false
	go m.readStream()

	return nil
}

func (m *TContainerMonitor) Stop() error {
	m.stop = true
	if m.cli == nil {
		return nil
	}
	return m.cli.Close()
}

func (m *TContainerMonitor) init() error {
	if m.Id == "" {
		return errors.New("configuration error: container ID must be set")
	}

	if cli, err := client.NewClientWithOpts(client.FromEnv); err != nil {
		return err
	} else {
		m.cli = cli
	}

	if containerInfo, err := m.cli.ContainerInspect(context.Background(), m.Id); err != nil {
		return err
	} else {
		m.Labels = containerInfo.Config.Labels
	}
	return nil
}

func (m *TContainerMonitor) readStream() {
	stream, err := m.cli.ContainerStats(context.Background(), m.Id, true)
	if err != nil {
		log.Println("Error starting container statistic listening: ", err)
		return
	}
	decoder := json.NewDecoder(stream.Body)

	for {
		time.Sleep(1 * time.Second)
		if m.stop {
			break
		}

		statistic := new(TContainerStatistic)
		if er := decoder.Decode(statistic); er != nil {
			if er != io.EOF {
				log.Println("Error reading from input:", er)
			}
			m.stop = true
			break
		}

		containerInspect, err := m.cli.ContainerInspect(context.Background(), m.Id)
		if err != nil {
			log.Println("Error inspecting container:", err)
			return
		}
		containerState := containerInspect.State.Status // 获取容器的运行状态
		statistic.RunningState = containerState

		if m.Name == "" {
			m.Name = statistic.Name
		}

		statistic.Labels = m.Labels

		if m.OnStatRead != nil {
			m.OnStatRead(statistic)
		}
	}
	if m.OnRemove != nil {
		m.OnRemove(m.Id)
	}
}
