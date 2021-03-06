package broadcom

import (
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"strconv"

	"github.com/elastic/beats/libbeat/common"
	"github.com/halo779/broadcombeat/config"
	"github.com/ziutek/telnet"
)

const timeout = 10 * time.Second

func checkErr(err error) error {
	if err != nil {
		log.Println("Telnet Error:", err)
	}
	return err
}
func round(v float64, decimals int) float64 {
	var pow float64 = 1
	for i := 0; i < decimals; i++ {
		pow *= 10
	}
	return float64(int((v*pow)+0.5)) / pow
}

func expect(t *telnet.Conn, d ...string) {
	checkErr(t.SetReadDeadline(time.Now().Add(timeout)))
	checkErr(t.SkipUntil(d...))
}

func sendln(t *telnet.Conn, s string) {
	checkErr(t.SetWriteDeadline(time.Now().Add(timeout)))
	buf := make([]byte, len(s)+1)
	copy(buf, s)
	buf[len(s)] = '\n'
	_, err := t.Write(buf)
	checkErr(err)
}

func toNum(s string) float64 {
	val, _ := strconv.ParseFloat(s, 32)
	val = round(val, 1)
	return val
}

func Process(evt common.MapStr, cfg config.Config) common.MapStr {
	Results := evt

	//Allocate default values.
	dst, user, passwd := "192.168.1.1:23", "admin", "admin"
	dst = cfg.Host
	Results.Put("DataSource", dst)
	t, err := telnet.Dial("tcp", dst)
	atLinkStats := false
	checkErr(err)
	if err != nil {
		return nil
	}
	t.SetUnixWriteMode(true)

	var xdslstat []byte

	expect(t, "ogin: ")
	sendln(t, user)
	expect(t, "ssword: ")
	sendln(t, passwd)
	expect(t, ">")
	sendln(t, "xdslctl info --stats")
	xdslstat, err = t.ReadBytes('>')
	checkErr(err)

	xdslstatsstr := string(xdslstat[:])
	xdslstatsa := strings.Split(xdslstatsstr, "\n")

	for _, line := range xdslstatsa {
		if strings.Contains(line, "Max") {
			re := regexp.MustCompile(`(\d+){2}`)
			matches := re.FindAllStringSubmatch(line, -1)
			up, down := matches[0][0], matches[1][0]
			Results.Put("AttainableDownstream", toNum(down))
			Results.Put("AttainableUpstream", toNum(up))
		}
		if strings.HasPrefix(line, "Status: ") {
			status := strings.TrimSpace(strings.Split(line, ": ")[1])
			Results.Put("DSLStatus", status)
		}
		if strings.HasPrefix(line, "Mode: ") {
			mode := strings.TrimSpace(strings.Split(line, ": ")[1])
			Results.Put("DSLMode", mode)
		}
		if strings.HasPrefix(line, "TPS-TC: ") {
			tps := strings.TrimSpace(strings.Split(line, ": ")[1])
			Results.Put("TransportConvergence", tps)
		}
		if strings.HasPrefix(line, "Line Status: ") {
			status := strings.TrimSpace(strings.Split(line, ": ")[1])
			Results.Put("LineStatus", status)
		}
		if strings.HasPrefix(line, "VDSL2 Profile: ") {
			profile := strings.TrimSpace(strings.Split(line, ": ")[1])
			profile = strings.TrimSpace(strings.Split(profile, "Profile ")[1])
			Results.Put("VDSL2BandProfile", profile)
		}
		if strings.Contains(line, "Last Retrain Reason: ") {
			re := regexp.MustCompile(`(\d+){1}`)
			matches := re.FindAllStringSubmatch(line, -1)
			retrain := matches[0][0]
			Results.Put("LastRetrainReason", retrain)
		}
		if strings.Contains(line, "Trellis: ") {
			re := regexp.MustCompile(`([A-Z])\w+`)
			matches := re.FindAllStringSubmatch(line, -1)
			up, down := matches[1][0], matches[2][0]
			Results.Put("TrellisDownstream", down)
			Results.Put("TrellisUpstream", up)
		}
		if strings.Contains(line, "Last initialization procedure status: ") {
			re := regexp.MustCompile(`(\d+){1}`)
			matches := re.FindAllStringSubmatch(line, -1)
			init := matches[0][0]
			Results.Put("LastInitializationProc", init)
		}
		if strings.HasPrefix(line, "Link Power State:") {
			state := strings.TrimSpace(strings.Split(line, ":")[1])
			Results.Put("PowerState", state)
		}
		if strings.Contains(line, "SNR") {
			re := regexp.MustCompile(`(\d*\.\d*)`)
			matches := re.FindAllStringSubmatch(line, -1)
			down, up := matches[0][0], matches[1][0]
			Results.Put("SNRDownstream", toNum(down))
			Results.Put("SNRUpstream", toNum(up))
		}
		if strings.Contains(line, "Bearer: 0,") {
			re := regexp.MustCompile(`(\d+){2}`)
			matches := re.FindAllStringSubmatch(line, -1)
			up, down := matches[0][0], matches[1][0]
			Results.Put("DownstreamSync", toNum(down))
			Results.Put("UpstreamSync", toNum(up))
		}
		if strings.Contains(line, "Attn(") {
			re := regexp.MustCompile(`(\d*\.\d*)`)
			matches := re.FindAllStringSubmatch(line, -1)
			down, up := matches[0][0], matches[1][0]
			Results.Put("LineAtteuationUpstream", toNum(up))
			Results.Put("LineAtteuationDownstream", toNum(down))
		}
		if strings.HasPrefix(line, "INP:") {
			re := regexp.MustCompile(`(\d*\.\d*)`)
			matches := re.FindAllStringSubmatch(line, -1)
			down, up := matches[0][0], matches[1][0]
			Results.Put("INPUpstream", toNum(up))
			Results.Put("INPDownstream", toNum(down))
		}
		if strings.HasPrefix(line, "INPRein:") {
			re := regexp.MustCompile(`(\d*\.\d*)`)
			matches := re.FindAllStringSubmatch(line, -1)
			down, up := matches[0][0], matches[1][0]
			Results.Put("INPReinUpstream", toNum(up))
			Results.Put("INPReinDownstream", toNum(down))
		}
		if strings.HasPrefix(line, "delay:") {
			re := regexp.MustCompile(`(\d+)`)
			matches := re.FindAllStringSubmatch(line, -1)
			up, down := matches[0][0], matches[1][0]
			Results.Put("DelayUpstream", toNum(up))
			Results.Put("DelayDownstream", toNum(down))
		}
		if strings.HasPrefix(line, "D:") {
			re := regexp.MustCompile(`(\d+)`)
			matches := re.FindAllStringSubmatch(line, -1)
			up, down := matches[0][0], matches[1][0]
			Results.Put("InterleavingDepthUpstream", toNum(up))
			Results.Put("InterleavingDepthDownstream", toNum(down))
		}
		if strings.Contains(line, "Pwr(dBm)") {
			re := regexp.MustCompile(`(\d*\.\d*)`)
			matches := re.FindAllStringSubmatch(line, -1)
			up, down := matches[0][0], matches[1][0]
			Results.Put("PowerUpstream", toNum(up))
			Results.Put("PowerDownstream", toNum(down))
		}
		if strings.HasPrefix(line, "Since Link time = ") {
			re := regexp.MustCompile(`= ([\s\S]*)`)
			matches := re.FindAllStringSubmatch(line, -1)
			dslUptime := strings.TrimSpace(matches[0][1])
			Results.Put("LinkUptime", dslUptime)

			dslUptimeSec, dslUptimeMin, dslUptimeHours, dslUptimeDays := "0", "0", "0", "0"

			re = regexp.MustCompile(`(?:(?P<sec>\d+)(?: sec))`)
			matches = re.FindAllStringSubmatch(dslUptime, -1)
			if len(matches) != 0 {
				dslUptimeSec = strings.TrimSpace(matches[0][1])
			}
			re = regexp.MustCompile(`(?:(?P<min>\d+)(?: min))`)
			matches = re.FindAllStringSubmatch(dslUptime, -1)
			if len(matches) != 0 {
				dslUptimeMin = strings.TrimSpace(matches[0][1])
			}
			re = regexp.MustCompile(`(?:(?P<hours>\d+)(?: hours))`)
			matches = re.FindAllStringSubmatch(dslUptime, -1)
			if len(matches) != 0 {
				dslUptimeHours = strings.TrimSpace(matches[0][1])
			}
			re = regexp.MustCompile(`(?:(?P<days>\d+)(?: days))`)
			matches = re.FindAllStringSubmatch(dslUptime, -1)
			if len(matches) != 0 {
				dslUptimeDays = strings.TrimSpace(matches[0][1])
			}

			dslUptimeTotalSeconds := toNum(dslUptimeDays)*86400 + toNum(dslUptimeHours)*3600 + toNum(dslUptimeMin)*60 + toNum(dslUptimeSec)
			Results.Put("DslUpTimeSeconds", dslUptimeTotalSeconds)
		}

		if strings.HasPrefix(line, "Total time = ") {
			re := regexp.MustCompile(`= ([\s\S]*)`)
			matches := re.FindAllStringSubmatch(line, -1)
			deviceUptime := strings.TrimSpace(matches[0][1])
			Results.Put("DeviceUpTime", deviceUptime)
			deviceUptimeSec, deviceUptimeMin, deviceUptimeHours, deviceUptimeDays := "0", "0", "0", "0"
			re = regexp.MustCompile(`(?:(?P<sec>\d+)(?: sec))`)
			matches = re.FindAllStringSubmatch(deviceUptime, -1)
			if len(matches) != 0 {
				deviceUptimeSec = strings.TrimSpace(matches[0][1])
			}
			re = regexp.MustCompile(`(?:(?P<min>\d+)(?: min))`)
			matches = re.FindAllStringSubmatch(deviceUptime, -1)
			if len(matches) != 0 {
				deviceUptimeMin = strings.TrimSpace(matches[0][1])
			}
			re = regexp.MustCompile(`(?:(?P<hours>\d+)(?: hours))`)
			matches = re.FindAllStringSubmatch(deviceUptime, -1)
			if len(matches) != 0 {
				deviceUptimeHours = strings.TrimSpace(matches[0][1])
			}
			re = regexp.MustCompile(`(?:(?P<days>\d+)(?: days))`)
			matches = re.FindAllStringSubmatch(deviceUptime, -1)
			if len(matches) != 0 {
				deviceUptimeDays = strings.TrimSpace(matches[0][1])
			}

			deviceUptimeTotalSeconds := toNum(deviceUptimeDays)*86400 + toNum(deviceUptimeHours)*3600 + toNum(deviceUptimeMin)*60 + toNum(deviceUptimeSec)
			Results.Put("DeviceUpTimeSeconds", deviceUptimeTotalSeconds)
		}

		if strings.HasPrefix(line, "Since Link time =") {
			atLinkStats = true
		}
		if strings.HasPrefix(line, "FEC: ") && atLinkStats {
			re := regexp.MustCompile(`(\d+)`)
			matches := re.FindAllStringSubmatch(line, -1)
			down, up := matches[0][0], matches[1][0]
			Results.Put("LinkFECUpstream", toNum(up))
			Results.Put("LinkFECDownstream", toNum(down))
		}
		if strings.HasPrefix(line, "CRC: ") && atLinkStats {
			re := regexp.MustCompile(`(\d+)`)
			matches := re.FindAllStringSubmatch(line, -1)
			down, up := matches[0][0], matches[1][0]
			Results.Put("LinkCRCUpstream", toNum(up))
			Results.Put("LinkCRCDownstream", toNum(down))
		}
		if strings.HasPrefix(line, "ES: ") && atLinkStats {
			re := regexp.MustCompile(`(\d+)`)
			matches := re.FindAllStringSubmatch(line, -1)
			down, up := matches[0][0], matches[1][0]
			Results.Put("LinkESUpstream", toNum(up))
			Results.Put("LinkESDownstream", toNum(down))
		}
		if strings.HasPrefix(line, "SES: ") && atLinkStats {
			re := regexp.MustCompile(`(\d+)`)
			matches := re.FindAllStringSubmatch(line, -1)
			down, up := matches[0][0], matches[1][0]
			Results.Put("LinkSESUpstream", toNum(up))
			Results.Put("LinkSESDownstream", toNum(down))
		}
		if strings.HasPrefix(line, "UAS: ") && atLinkStats {
			re := regexp.MustCompile(`(\d+)`)
			matches := re.FindAllStringSubmatch(line, -1)
			down, up := matches[0][0], matches[1][0]
			Results.Put("LinkUASUpstream", toNum(up))
			Results.Put("LinkUASDownstream", toNum(down))
		}
		if strings.HasPrefix(line, "LOS: ") && atLinkStats {
			re := regexp.MustCompile(`(\d+)`)
			matches := re.FindAllStringSubmatch(line, -1)
			down, up := matches[0][0], matches[1][0]
			Results.Put("LinkLOSUpstream", toNum(up))
			Results.Put("LinkLOSDownstream", toNum(down))
		}
		if strings.HasPrefix(line, "LOF: ") && atLinkStats {
			re := regexp.MustCompile(`(\d+)`)
			matches := re.FindAllStringSubmatch(line, -1)
			down, up := matches[0][0], matches[1][0]
			Results.Put("LinkLOFUpstream", toNum(up))
			Results.Put("LinkLOFDownstream", toNum(down))
		}
		if strings.HasPrefix(line, "LOM: ") && atLinkStats {
			re := regexp.MustCompile(`(\d+)`)
			matches := re.FindAllStringSubmatch(line, -1)
			down, up := matches[0][0], matches[1][0]
			Results.Put("LinkLOMUpstream", toNum(up))
			Results.Put("LinkLOMDownstream", toNum(down))
			atLinkStats = false
		}

	}
	sendln(t, "exit")
	os.Stdout.WriteString("\n")
	return Results
}
