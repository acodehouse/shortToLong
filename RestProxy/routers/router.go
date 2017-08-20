package routers

import (
	"RestProxy/controllers"
	"github.com/astaxie/beego"
)

func init() {

	beego.SetLogger("file", `{"filename":"call.log"}`)
	//////////////////////////////////////////////
	beego.Router("/", &controllers.MainController{})
	myController := &controllers.RestProxyController{}
	myController.MyInit()
	beego.Router("/RestProxy", myController, "post:Call")
}
