package controllers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"

	"../../util"
)

// MarkAsUnreadController 处理邮件标记为已读的请求
type MarkAsUnreadController struct {
	beego.Controller
}

// Get 处理 GET 请求
func (controller *MarkAsUnreadController) Get() {
	controller.Post()
}

// Post 处理 POST 请求
func (controller *MarkAsUnreadController) Post() {
	var ids = controller.GetString("ids")
	if ids == "" {
		controller.Abort(strconv.Itoa(http.StatusBadRequest))
	}

	_, err := gSrvConfig.Ormer.QueryTable("email").
		Filter("Id__in", strings.Split(ids, ",")).
		Update(orm.Params{"IsRead": 0})

	if err != nil {
		log.Println(err)
		controller.Abort(strconv.Itoa(http.StatusInternalServerError))
	}

	controller.Data["json"] = util.SimpleResponse()

	controller.ServeJson()
}
