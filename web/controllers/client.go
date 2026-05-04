package controllers

import (
	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/file"
	"ehang.io/nps/server"
	"github.com/astaxie/beego"
)

type ClientController struct {
	BaseController
}

// List renders the bot list page on GET, or returns the bot rows as
// JSON for the data-table component on POST. Each bot exposes its
// SOCKS5 credentials so the admin can copy a ready-to-use
// socks5://user:pass@host:port link.
func (s *ClientController) List() {
	if s.Ctx.Request.Method == "GET" {
		s.Data["menu"] = "client"
		s.Data["socks5Port"] = beego.AppConfig.String("socks5_port")
		s.SetInfo("bots")
		s.display("client/list")
		return
	}
	start, length := s.GetAjaxParams()
	clientIdSession := s.GetSession("clientId")
	var clientId int
	if clientIdSession != nil {
		if v, ok := clientIdSession.(int); ok {
			clientId = v
		}
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

// Edit updates the optional remark of an existing bot. All other
// fields (MachineGuid, VerifyKey, SOCKS5 credentials) are immutable
// from the panel: identity is owned by the bot itself and the
// SOCKS5 credentials are auto-generated to guarantee uniqueness.
func (s *ClientController) Edit() {
	if s.Ctx.Request.Method != "POST" {
		s.Redirect(beego.AppConfig.String("web_base_url")+"/client/list", 302)
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
	c.Remark = s.getEscapeString("remark")
	if err := file.GetDb().UpdateClient(c); err != nil {
		s.AjaxErr(err.Error())
		return
	}
	s.AjaxOk("save success")
}

// ChangeStatus enables / disables a bot (without deleting the row).
func (s *ClientController) ChangeStatus() {
	id := s.GetIntNoErr("id")
	if client, err := file.GetDb().GetClient(id); err == nil {
		client.Status = s.GetBoolNoErr("status")
		if !client.Status {
			server.DelClientConnect(client.Id)
		}
		if err := file.GetDb().UpdateClient(client); err != nil {
			s.AjaxErr(err.Error())
			return
		}
		s.AjaxOk("modified success")
	}
	s.AjaxErr("modified fail")
}

// Del removes a bot from the panel. The next time the same physical
// host (same MachineGuid) connects, the bridge auto-registers it
// again as a brand-new bot with fresh SOCKS5 credentials.
func (s *ClientController) Del() {
	id := s.GetIntNoErr("id")
	if err := file.GetDb().DelClient(id); err != nil {
		s.AjaxErr("delete error")
	}
	server.DelTunnelAndHostByClientId(id, false)
	server.DelClientConnect(id)
	s.AjaxOk("delete success")
}
