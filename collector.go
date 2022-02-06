package tc4400exporter

import (
	"context"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

var _ prometheus.Collector = &Collector{}

// Collector is a prometheus.Collector for a TC4400 modem.
type Collector struct {
	ModemInfo *prometheus.Desc

	UptimeSecondsTotal *prometheus.Desc

	RxBytesTotal   *prometheus.Desc
	RxPacketsTotal *prometheus.Desc
	RxErrorsTotal  *prometheus.Desc
	RxDropsTotal   *prometheus.Desc

	TxBytesTotal   *prometheus.Desc
	TxPacketsTotal *prometheus.Desc
	TxErrorsTotal  *prometheus.Desc
	TxDropsTotal   *prometheus.Desc

	client *Client
}

// NewCollector constructs a collector using a device.
func NewCollector(c *Client) *Collector {
	return &Collector{
		ModemInfo: prometheus.NewDesc(
			"tc4400_modem_info",
			"Information about a modem.",
			[]string{"serial_number", "board_id", "build_timestamp", "hardware_version", "software_version", "linux_version", "system_time", "ipv4_addr", "ipv6_addr", "modem_mac", "lan_mac"},
			nil,
		),

		UptimeSecondsTotal: prometheus.NewDesc(
			"tc4400_uptime_seconds_total",
			"Device uptime in seconds.",
			nil,
			nil,
		),

		RxBytesTotal: prometheus.NewDesc(
			"tc4400_rx_bytes_total",
			"Current number of bytes recieved.",
			[]string{"interface"},
			nil,
		),

		RxPacketsTotal: prometheus.NewDesc(
			"tc4400_rx_packets_total",
			"Current number of packets recieved.",
			[]string{"interface"},
			nil,
		),

		RxErrorsTotal: prometheus.NewDesc(
			"tc4400_rx_errors_total",
			"Current number of errors while attempting to recieve.",
			[]string{"interface"},
			nil,
		),

		RxDropsTotal: prometheus.NewDesc(
			"tc4400_rx_drops_total",
			"Current number of drops occured while attempting to recieve.",
			[]string{"interface"},
			nil,
		),

		TxBytesTotal: prometheus.NewDesc(
			"tc4400_tx_bytes_total",
			"Current number of bytes sent.",
			[]string{"interface"},
			nil,
		),

		TxPacketsTotal: prometheus.NewDesc(
			"tc4400_tx_packets_total",
			"Current number of packets sent.",
			[]string{"interface"},
			nil,
		),

		TxErrorsTotal: prometheus.NewDesc(
			"tc4400_tx_errors_total",
			"Current number of errors while attempting to send.",
			[]string{"interface"},
			nil,
		),

		TxDropsTotal: prometheus.NewDesc(
			"tc4400_tx_drops_total",
			"Current number of drops occured while attempting to send.",
			[]string{"interface"},
			nil,
		),

		client: c,
	}
}

// Describe implements prometheus.Collector.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ds := []*prometheus.Desc{
		c.ModemInfo,
		c.UptimeSecondsTotal,
		c.RxBytesTotal,
		c.RxPacketsTotal,
		c.RxErrorsTotal,
		c.RxDropsTotal,
		c.TxBytesTotal,
		c.TxPacketsTotal,
		c.TxErrorsTotal,
		c.TxDropsTotal,
	}

	for _, d := range ds {
		ch <- d
	}
}

// Collect implements prometheus.Collector.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	info, err := c.client.Info(ctx)
	if err != nil {
		log.Println(err)
		ch <- prometheus.NewInvalidMetric(c.UptimeSecondsTotal, err)
		return
	}

	stats, err := c.client.Stats(ctx)
	if err != nil {
		log.Println(err)
		ch <- prometheus.NewInvalidMetric(c.UptimeSecondsTotal, err)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		c.ModemInfo,
		prometheus.GaugeValue,
		1.0,
		info.SerialNumber, info.BoardID, info.BuildTimestamp.String(), info.HardwareVersion, info.SoftwareVersion, info.LinuxVersion, info.SystemTime.String(), info.IPv4Addr.String(), info.IPv6Addr.String(), info.ModemHardwareAddr.String(), info.LANHardwareAddr.String(),
	)

	ch <- prometheus.MustNewConstMetric(
		c.UptimeSecondsTotal,
		prometheus.CounterValue,
		info.Uptime.Seconds(),
	)

	for _, stat := range stats {
		ch <- prometheus.MustNewConstMetric(
			c.RxBytesTotal,
			prometheus.CounterValue,
			float64(stat.RxBytes),
			stat.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RxPacketsTotal,
			prometheus.CounterValue,
			float64(stat.RxPackets),
			stat.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RxErrorsTotal,
			prometheus.CounterValue,
			float64(stat.RxErrors),
			stat.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RxDropsTotal,
			prometheus.CounterValue,
			float64(stat.RxDrops),
			stat.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TxBytesTotal,
			prometheus.CounterValue,
			float64(stat.RxBytes),
			stat.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TxPacketsTotal,
			prometheus.CounterValue,
			float64(stat.RxPackets),
			stat.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TxErrorsTotal,
			prometheus.CounterValue,
			float64(stat.RxErrors),
			stat.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TxDropsTotal,
			prometheus.CounterValue,
			float64(stat.RxDrops),
			stat.Name,
		)
	}
}
