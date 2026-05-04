package controllers

import (
	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/crypt"
	"ehang.io/nps/lib/file"
	"ehang.io/nps/lib/rate"
	"ehang.io/nps/server"
	"github.com/astaxie/beego"
)

type ClientController struct {
	BaseController
}

// List renders the bot list page on GET, or returns the bot rows as JSON
// for the bootstrap-table component on POST. Each bot exposes its
// SOCKS5 credentials as well as the bridge command for the admin.
func (s *ClientController) List() {
	if s.Ctx.Request.Method == "GET" {
		s.Data["menu"] = "client"
		s.SetInfo("bots")
		s.display("client/list")
		return
	}
	start, length := s.GetAjaxParams()
	clientIdSession := s.GetSession("clientId")
	var clientId int
	if clientIdSession == nil {
		clientId = 0
	} else {
		clientId = clientIdSession.(int)
	}
	list, cnt := server.GetClientList(start, length,
		s.getEscapeString("search"),
		s.getEscapeString("sort"),
		s.getEscapeString("order"),
		clientId)
	cmd := make(map[string]interface{})
	ip := s.Ctx.Request.Host
	cmd["ip"] = common.GetIpByAddr(ip)
	cmd["bridgeType"] = beego.AppConfig.String("bridge_type")
	cmd["bridgePort"] = server.Bridge.TunnelPort
	cmd["socks5Port"] = beego.AppConfig.String("socks5_port")
	s.AjaxTable(list, cnt, cnt, cmd)
}

// Add provisions a new bot. Everything (vkey, SOCKS5 username and
// SOCKS5 password) is auto-generated; the admin only chooses an
// optional remark. The endpoint accepts a POST request from the
// bot list page.
func (s *ClientController) Add() {
	if s.Ctx.Request.Method != "POST" {
		// The legacy add page is no longer used; redirect to the list.
		s.Redirect(beego.AppConfig.String("web_base_url")+"/client/list", 302)
		return
	}
	t := &file.Client{
		VerifyKey: crypt.GetRandomString(16),
		Id:        int(file.GetDb().JsonDb.GetClientId()),
		Status:    true,
		Remark:    s.getEscapeString("remark"),
		Cnf: &file.Config{
			U:        crypt.GetRandomString(8),
			P:        crypt.GetRandomString(16),
			Compress: false,
			Crypt:    false,
		},
		ConfigConnAllow: true,
		Flow: &file.Flow{
			ExportFlow: 0,
			InletFlow:  0,
			FlowLimit:  0,
		},
	}
	if err := file.GetDb().NewClient(t); err != nil {
		s.AjaxErr(err.Error())
	}
	s.AjaxOk("add success")
}

// GetClient returns a single bot description as JSON.
func (s *ClientController) GetClient() {
	if s.Ctx.Request.Method == "POST" {
		id := s.GetIntNoErr("id")
		data := make(map[string]interface{})
		if c, err := file.GetDb().GetClient(id); err != nil {
			data["code"] = 0
		} else {
			data["code"] = 1
			data["data"] = c
		}
		s.Data["json"] = data
		s.ServeJSON()
	}
}

// Edit allows the admin to update the optional remark and the SOCKS5
// credentials for an existing bot.
func (s *ClientController) Edit() {
	id := s.GetIntNoErr("id")
	if s.Ctx.Request.Method != "POST" {
		s.Redirect(beego.AppConfig.String("web_base_url")+"/client/list", 302)
		return
	}
	c, err := file.GetDb().GetClient(id)
	if err != nil {
		s.AjaxErr("client ID not found")
		return
	}
	if !isAdmin(&s.BaseController) {
		s.AjaxErr("forbidden")
		return
	}
	if c.Cnf == nil {
		c.Cnf = &file.Config{}
	}
	c.Remark = s.getEscapeString("remark")
	if u := s.getEscapeString("u"); u != "" {
		c.Cnf.U = u
	}
	if p := s.getEscapeString("p"); p != "" {
		c.Cnf.P = p
	}
	if c.Rate != nil {
		c.Rate.Stop()
	}
	c.Rate = rate.NewRate(int64(2 << 23))
	c.Rate.Start()
	if err := file.GetDb().UpdateClient(c); err != nil {
		s.AjaxErr(err.Error())
		return
	}
	s.AjaxOk("save success")
}

// ChangeStatus enables / disables a bot.
func (s *ClientController) ChangeStatus() {
	id := s.GetIntNoErr("id")
	if client, err := file.GetDb().GetClient(id); err == nil {
		client.Status = s.GetBoolNoErr("status")
		if client.Status == false {
			server.DelClientConnect(client.Id)
		}
		s.AjaxOk("modified success")
	}
	s.AjaxErr("modified fail")
}

// Del removes a bot. Tunnels and hosts that may have been created in
// older versions are cleaned up too so we never leave orphans behind.
func (s *ClientController) Del() {
	id := s.GetIntNoErr("id")
	if err := file.GetDb().DelClient(id); err != nil {
		s.AjaxErr("delete error")
	}
	server.DelTunnelAndHostByClientId(id, false)
	server.DelClientConnect(id)
	s.AjaxOk("delete success")
}

// Regenerate creates a fresh SOCKS5 username and password for the
// selected bot.
func (s *ClientController) Regenerate() {
	if s.Ctx.Request.Method != "POST" {
		s.AjaxErr("post only")
		return
	}
	if !isAdmin(&s.BaseController) {
		s.AjaxErr("forbidden")
		return
	}
	id := s.GetIntNoErr("id")
	c, err := file.GetDb().GetClient(id)
	if err != nil {
		s.AjaxErr("client not found")
		return
	}
	if c.Cnf == nil {
		c.Cnf = &file.Config{}
	}
	c.Cnf.U = crypt.GetRandomString(8)
	c.Cnf.P = crypt.GetRandomString(16)
	if err := file.GetDb().UpdateClient(c); err != nil {
		s.AjaxErr(err.Error())
		return
	}
	s.AjaxOk("save success")
}
