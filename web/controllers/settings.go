package controllers

import (
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/file"
	"ehang.io/nps/lib/version"
	"ehang.io/nps/server"
	"github.com/astaxie/beego"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

// SettingsController exposes a single Settings page that shows the
// server load and lets the admin change a small set of high-level
// configuration values without restarting nps.
type SettingsController struct {
	BaseController
}

// Index renders the Settings page (admin only).
func (s *SettingsController) Index() {
	if !isAdmin(&s.BaseController) {
		s.error()
		return
	}
	s.Data["menu"] = "settings"
	s.SetInfo("settings")
	s.display("settings/index")
}

// Stats returns live server / process metrics that the page polls.
func (s *SettingsController) Stats() {
	if !isAdmin(&s.BaseController) {
		s.AjaxErr("forbidden")
		return
	}
	out := make(map[string]interface{})

	out["version"] = version.VERSION
	out["coreVersion"] = version.GetVersion()
	out["uptimeSeconds"] = int64(time.Since(server.StartTime).Seconds())

	cpuPercent, _ := cpu.Percent(0, false)
	if len(cpuPercent) > 0 {
		out["cpuPercent"] = round(cpuPercent[0])
	} else {
		out["cpuPercent"] = 0
	}
	out["cpuCount"] = runtime.NumCPU()

	if vm, err := mem.VirtualMemory(); err == nil {
		out["memTotal"] = vm.Total
		out["memUsed"] = vm.Used
		out["memPercent"] = round(vm.UsedPercent)
	}
	if sm, err := mem.SwapMemory(); err == nil {
		out["swapTotal"] = sm.Total
		out["swapUsed"] = sm.Used
		out["swapPercent"] = round(sm.UsedPercent)
	}
	if la, err := load.Avg(); err == nil {
		out["load1"] = round(la.Load1)
		out["load5"] = round(la.Load5)
		out["load15"] = round(la.Load15)
	}

	total, online := countBots()
	out["botsTotal"] = total
	out["botsOnline"] = online

	out["socks5Port"] = beego.AppConfig.String("socks5_port")
	out["socks5Ip"] = beego.AppConfig.String("socks5_ip")
	out["bridgePort"] = beego.AppConfig.String("bridge_port")
	out["bridgeType"] = beego.AppConfig.String("bridge_type")
	out["webPort"] = beego.AppConfig.String("web_port")
	out["webUsername"] = beego.AppConfig.String("web_username")

	s.Data["json"] = map[string]interface{}{"status": 1, "data": out}
	s.ServeJSON()
}

// SaveSocks5Port updates the shared SOCKS5 listening port and bind ip.
func (s *SettingsController) SaveSocks5Port() {
	if s.Ctx.Request.Method != "POST" {
		s.AjaxErr("post only")
		return
	}
	if !isAdmin(&s.BaseController) {
		s.AjaxErr("forbidden")
		return
	}
	portStr := s.getEscapeString("socks5_port")
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 0 || port > 65535 {
		s.AjaxErr("invalid port")
		return
	}
	ip := s.getEscapeString("socks5_ip")
	if ip == "" {
		ip = "0.0.0.0"
	}
	if err := server.StartSharedSocks5(ip, port); err != nil {
		s.AjaxErr(err.Error())
		return
	}
	beego.AppConfig.Set("socks5_port", portStr)
	beego.AppConfig.Set("socks5_ip", ip)
	if err := persistConf(map[string]string{
		"socks5_port": portStr,
		"socks5_ip":   ip,
	}); err != nil {
		s.AjaxErr(err.Error())
		return
	}
	s.AjaxOk("save success")
}

// SaveAdmin updates the admin web username / password.
func (s *SettingsController) SaveAdmin() {
	if s.Ctx.Request.Method != "POST" {
		s.AjaxErr("post only")
		return
	}
	if !isAdmin(&s.BaseController) {
		s.AjaxErr("forbidden")
		return
	}
	username := s.getEscapeString("web_username")
	password := s.getEscapeString("web_password")
	if username == "" || password == "" {
		s.AjaxErr("username and password are required")
		return
	}
	beego.AppConfig.Set("web_username", username)
	beego.AppConfig.Set("web_password", password)
	if err := persistConf(map[string]string{
		"web_username": username,
		"web_password": password,
	}); err != nil {
		s.AjaxErr(err.Error())
		return
	}
	s.AjaxOk("save success")
}

func persistConf(kv map[string]string) error {
	path := filepath.Join(common.GetRunPath(), "conf", "nps.conf")
	return common.UpdateConfFile(path, kv)
}

func countBots() (total, online int) {
	file.GetDb().JsonDb.Clients.Range(func(key, value interface{}) bool {
		c, ok := value.(*file.Client)
		if !ok || c == nil || c.NoDisplay {
			return true
		}
		total++
		if server.Bridge != nil {
			if _, ok := server.Bridge.Client.Load(c.Id); ok {
				online++
			}
		}
		return true
	})
	return
}

func round(f float64) float64 {
	return float64(int64(f*100)) / 100
}
