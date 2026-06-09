package agent

import (
	"fmt"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// Ссобирает системные метрики
// PollSystemMetrics собирает системные метрики (CPU, память и др.).
func PollSystemMetrics(metrics *Metrics) error {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	metrics.SetGauge("TotalMemory", float64(vm.Total))
	metrics.SetGauge("FreeMemory", float64(vm.Free))

	cpuValues, err := cpu.Percent(0, true)
	if err != nil {
		return err
	}

	for i, val := range cpuValues {
		metrics.SetGauge(fmt.Sprintf("CPUutilization%d", i+1), val)
	}

	return nil
}
