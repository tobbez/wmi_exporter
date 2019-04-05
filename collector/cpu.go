// +build windows

package collector

import (
	"strings"

	"github.com/leoluk/perflib_exporter/perflib"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["cpu"] = NewCPUCollector
}

// A CPUCollector is a Prometheus collector for WMI Win32_PerfRawData_PerfOS_Processor metrics
type CPUCollector struct {
	CStateSecondsTotal *prometheus.Desc
	TimeTotal          *prometheus.Desc
	InterruptsTotal    *prometheus.Desc
	DPCsTotal          *prometheus.Desc
}

// NewCPUCollector constructs a new CPUCollector
func NewCPUCollector() (Collector, error) {
	const subsystem = "cpu"
	return &CPUCollector{
		CStateSecondsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cstate_seconds_total"),
			"Time spent in low-power idle state",
			[]string{"core", "state"},
			nil,
		),
		TimeTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "time_total"),
			"Time that processor spent in different modes (idle, user, system, ...)",
			[]string{"core", "mode"},
			nil,
		),

		InterruptsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "interrupts_total"),
			"Total number of received and serviced hardware interrupts",
			[]string{"core"},
			nil,
		),
		DPCsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dpcs_total"),
			"Total number of received and serviced deferred procedure calls (DPCs)",
			[]string{"core"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *CPUCollector) Collect(ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting cpu metrics:", desc, err)
		return err
	}
	return nil
}

type PerflibProcessor struct {
	Name                  string
	C1Transitions         float64 `perflib:"C1 Transitions/sec"`
	C2Transitions         float64 `perflib:"C2 Transitions/sec"`
	C3Transitions         float64 `perflib:"C3 Transitions/sec"`
	DPCRate               float64 `perflib:"DPC Rate"`
	DPCsQueued            float64 `perflib:"DPCs Queued/sec"`
	Interrupts            float64 `perflib:"Interrupts/sec"`
	PercentC2Time         float64 `perflib:"% C1 Time"`
	PercentC3Time         float64 `perflib:"% C2 Time"`
	PercentC1Time         float64 `perflib:"% C3 Time"`
	PercentDPCTime        float64 `perflib:"% DPC Time"`
	PercentIdleTime       float64 `perflib:"% Idle Time"`
	PercentInterruptTime  float64 `perflib:"% Interrupt Time"`
	PercentPrivilegedTime float64 `perflib:"% Privileged Time"`
	PercentProcessorTime  float64 `perflib:"% Processor Time"`
	PercentUserTime       float64 `perflib:"% User Time"`
}

func (c *CPUCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	const processorCounterIndex = "238"
	objects, err := perflib.QueryPerformanceData(processorCounterIndex)
	if err != nil {
		return nil, err
	}
	ps := make([]PerflibProcessor, 0)
	err = UnmarshalObject(objects[0], &ps)
	if err != nil {
		return nil, err
	}

	for _, data := range ps {
		if strings.Contains(strings.ToLower(data.Name), "_total") {
			continue
		}

		core := data.Name

		ch <- prometheus.MustNewConstMetric(
			c.CStateSecondsTotal,
			prometheus.GaugeValue,
			data.PercentC1Time,
			core, "c1",
		)
		ch <- prometheus.MustNewConstMetric(
			c.CStateSecondsTotal,
			prometheus.GaugeValue,
			data.PercentC2Time,
			core, "c2",
		)
		ch <- prometheus.MustNewConstMetric(
			c.CStateSecondsTotal,
			prometheus.GaugeValue,
			data.PercentC3Time,
			core, "c3",
		)

		ch <- prometheus.MustNewConstMetric(
			c.TimeTotal,
			prometheus.GaugeValue,
			data.PercentIdleTime,
			core, "idle",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TimeTotal,
			prometheus.GaugeValue,
			data.PercentInterruptTime,
			core, "interrupt",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TimeTotal,
			prometheus.GaugeValue,
			data.PercentDPCTime,
			core, "dpc",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TimeTotal,
			prometheus.GaugeValue,
			data.PercentPrivilegedTime,
			core, "privileged",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TimeTotal,
			prometheus.GaugeValue,
			data.PercentUserTime,
			core, "user",
		)

		ch <- prometheus.MustNewConstMetric(
			c.InterruptsTotal,
			prometheus.CounterValue,
			data.Interrupts,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.DPCsTotal,
			prometheus.CounterValue,
			data.DPCsQueued,
			core,
		)
	}

	return nil, nil
}
