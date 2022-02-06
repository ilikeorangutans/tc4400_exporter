package tc4400exporter

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	SummaryURL  = "/info.html"
	SoftwareURL = "/cmswinfo.html"
	StatsURL    = "/statsifc.html"

	labelNetworkSpecification = "Standard Specification Compliant"
	labelHardwareVersion      = "Hardware Version"
	labelSerialNumber         = "Cable Modem Serial Number"
	ipAddrsRegExPattern       = "IPAddrValue = 'IPv4=(.*) IPv6=(.*)';"

	labelBoardID           = "Board ID:"
	labelBuildTimestamp    = "Build Timestamp:"
	labelSoftwareVersion   = "Software Version:"
	labelLinuxVersion      = "Linux Version:"
	labelUptime            = "Uptime:"
	labelSystemTime        = "Systime:"
	labelModemHardwareAddr = "CM Hardware Address:"
	labelLANHardwareAddr   = "LAN Hardware Address:"
)

type Client struct {
	HTTPClient *http.Client
	RootURL    string
	Username   string
	Password   string
}

type Info struct {
	NetworkSpecification string
	SerialNumber         string
	BoardID              string
	BuildTimestamp       time.Time
	HardwareVersion      string
	SoftwareVersion      string
	LinuxVersion         string
	Uptime               time.Duration
	SystemTime           time.Time
	IPv4Addr             net.IP
	IPv6Addr             net.IP
	ModemHardwareAddr    net.HardwareAddr
	LANHardwareAddr      net.HardwareAddr
}

func (c *Client) Info(ctx context.Context) (*Info, error) {
	var info Info
	var parseErr error

	{
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.RootURL+SummaryURL, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating hardware info request: %v", err)
		}

		req.SetBasicAuth(c.Username, c.Password)

		res, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error performing hardware info request: %v", err)
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return nil, fmt.Errorf("error parsing hardware info html response: %v", err)
		}

		doc.Find("table tr").Each(func(i int, s *goquery.Selection) {
			firstTDNode := s.Find("td").First()
			firstTDText := firstTDNode.Text()

			secondTDNode := firstTDNode.Next()
			secondTDText := secondTDNode.Text()

			switch firstTDText {
			case labelBoardID:
				info.BoardID = secondTDText
			case labelBuildTimestamp:
				buildTimestamp, err := time.Parse("20060102_1504", secondTDText)
				if err != nil {
					parseErr = fmt.Errorf("error parsing build timestamp: %v", err)
					return
				}

				info.BuildTimestamp = buildTimestamp
			case labelSoftwareVersion:
				info.SoftwareVersion = secondTDText
			case labelLinuxVersion:
				info.LinuxVersion = secondTDText
			case labelUptime:
				uptimeParts := strings.Split(secondTDText, "D ")

				days, err := strconv.ParseInt(uptimeParts[0], 10, 64)
				if err != nil {
					parseErr = fmt.Errorf("error parsing uptime days: %v", err)
					return
				}

				uptimeDays := time.Duration(days) * 24 * time.Hour

				uptime, err := time.ParseDuration(strings.ToLower(strings.ReplaceAll(uptimeParts[1], " ", "")))
				if err != nil {
					parseErr = fmt.Errorf("error parsing uptime remainer: %v", err)
					return
				}

				info.Uptime = uptimeDays + uptime
			case labelSystemTime:
				systemTime, err := time.Parse(time.RFC3339, secondTDText)
				if err != nil {
					parseErr = fmt.Errorf("error parsing system time: %v", err)
					return
				}

				info.SystemTime = systemTime
			case labelModemHardwareAddr:
				addr, err := net.ParseMAC(secondTDText)
				if err != nil {
					parseErr = fmt.Errorf("error parsing modem hardware address: %v", err)
				}
				info.ModemHardwareAddr = addr
			case labelLANHardwareAddr:
				addr, err := net.ParseMAC(secondTDText)
				if err != nil {
					parseErr = fmt.Errorf("error parsing LAN hardware address: %v", err)
				}
				info.LANHardwareAddr = addr
			}
		})
		if parseErr != nil {
			return nil, parseErr
		}
	}

	{
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.RootURL+SoftwareURL, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating software info request: %v", err)
		}

		req.SetBasicAuth(c.Username, c.Password)

		res, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error performing software info request: %v", err)
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return nil, fmt.Errorf("error parsing software info html response: %v", err)
		}

		doc.Find("table tr").Each(func(i int, s *goquery.Selection) {
			firstTDNode := s.Find("td").First()
			firstTDText := firstTDNode.Text()

			secondTDNode := firstTDNode.Next()
			secondTDText := secondTDNode.Text()

			switch firstTDText {
			case labelNetworkSpecification:
				info.NetworkSpecification = secondTDText
			case labelHardwareVersion:
				info.HardwareVersion = secondTDText
			case labelSerialNumber:
				info.SerialNumber = secondTDText
			}
		})

		html, err := doc.Html()
		if err != nil {
			return nil, fmt.Errorf("error getting html from doc: %v", err)
		}

		ipAddrsRegEx := regexp.MustCompile(ipAddrsRegExPattern)
		ipAddrs := ipAddrsRegEx.FindAllStringSubmatch(html, -1)

		info.IPv4Addr = net.ParseIP(ipAddrs[0][1])
		info.IPv6Addr = net.ParseIP(ipAddrs[0][2])
	}

	return &info, nil
}

type InterfaceStats struct {
	Name string

	RxBytes   int64
	RxPackets int64
	RxErrors  int64
	RxDrops   int64

	TxBytes   int64
	TxPackets int64
	TxErrors  int64
	TxDrops   int64
}

func (c *Client) Stats(ctx context.Context) ([]InterfaceStats, error) {
	var stats []InterfaceStats
	var parseErr error

	{
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.RootURL+StatsURL, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating stats request: %v", err)
		}

		req.SetBasicAuth(c.Username, c.Password)

		res, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error performing stats request: %v", err)
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return nil, fmt.Errorf("error parsing stats html response: %v", err)
		}

		doc.Find("table tr").Each(func(i int, s *goquery.Selection) {
			if i >= 3 {
				firstTDNode := s.Find("td").First()

				stat := InterfaceStats{
					Name: firstTDNode.Text(),
				}

				firstTDNode.NextAll().Each(func(i int, s *goquery.Selection) {
					val, err := strconv.ParseInt(s.Text(), 10, 64)
					if err != nil {
						parseErr = err
					}

					switch i {
					case 0:
						stat.RxBytes = val
					case 1:
						stat.RxPackets = val
					case 2:
						stat.RxErrors = val
					case 3:
						stat.RxDrops = val
					case 4:
						stat.TxBytes = val
					case 5:
						stat.TxPackets = val
					case 6:
						stat.TxErrors = val
					case 7:
						stat.TxDrops = val
					}
				})

				stats = append(stats, stat)
			}
		})
		if parseErr != nil {
			return nil, parseErr
		}
	}

	return stats, nil
}
