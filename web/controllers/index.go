package controllers

import "github.com/astaxie/beego"

// IndexController exists for backwards compatibility with the old
// dashboard URLs but every action now redirects to the bot list. The
// admin panel was simplified to a single page that shows connected
// bots together with their SOCKS5 credentials.
type IndexController struct {
	BaseController
}

func (s *IndexController) Index() {
	s.Redirect(beego.AppConfig.String("web_base_url")+"/client/list", 302)
}
