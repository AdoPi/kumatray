package main

import (
	"bufio"
	"flag"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/getlantern/systray"
	"github.com/pkg/browser"
)

const (
	StatusUnknown int32 = iota
	StatusOK
	StatusProblem
)

func main() {
	urlFlag := flag.String("url", "http://127.0.0.1:3001", "Uptime Kuma metrics URL")
	intervalFlag := flag.Int("interval", 30, "Intervalle en secondes")
	flag.Parse()

	var status int32 = StatusUnknown

	onReady := func() {
		openItem := systray.AddMenuItem("Open in browser", "")
		forceItem := systray.AddMenuItem("Force refresh", "")
		systray.AddSeparator()
		quitItem := systray.AddMenuItem("Quit", "")

		updateIcon(atomic.LoadInt32(&status))

		go func() {
			ticker := time.NewTicker(time.Duration(*intervalFlag) * time.Second)
			defer ticker.Stop()
			for {
				checkMetrics(*urlFlag, &status)
				updateIcon(atomic.LoadInt32(&status))
				<-ticker.C
			}
		}()

		go func() {
			for {
				select {
				case <-openItem.ClickedCh:
					_ = browser.OpenURL(*urlFlag)
				case <-forceItem.ClickedCh:
					checkMetrics(*urlFlag, &status)
					updateIcon(atomic.LoadInt32(&status))
				case <-quitItem.ClickedCh:
					systray.Quit()
					return
				}
			}
		}()
	}

	systray.Run(onReady, func() {})
}

func checkMetrics(url string, status *int32) {
	url = url + "/metrics"
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		atomic.StoreInt32(status, StatusProblem)
		return
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)

	foundDown := false
	foundUp := false
	onlyMaintenanceOrPending := true

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "monitor_status{") {
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			valueStr := parts[len(parts)-1]
			val, err := strconv.Atoi(valueStr)
			if err != nil {
				continue
			}

			switch val {
			case 0: // DOWN
				foundDown = true
				onlyMaintenanceOrPending = false
			case 1: // UP
				foundUp = true
				onlyMaintenanceOrPending = false
			case 2: // PENDING
			case 3: // MAINTENANCE
			}
		}
	}

	if foundDown {
		atomic.StoreInt32(status, StatusProblem)
	} else if foundUp && !onlyMaintenanceOrPending {
		atomic.StoreInt32(status, StatusOK)
	} else {
		atomic.StoreInt32(status, StatusUnknown)
	}
}

func updateIcon(s int32) {
	switch s {
	case StatusOK:
		systray.SetTitle("🟢")
		systray.SetTooltip("Everything is UP")
	case StatusProblem:
		systray.SetTitle("🔴")
		systray.SetTooltip("At least one service is DOWN")
	default:
		systray.SetTitle("🟡")
		systray.SetTooltip("Pending...")
	}
}

