package main

import (
	"jiacrontab/libs/mailer"
	"jiacrontab/libs/rpc"
	db "jiacrontab/model"
	"jiacrontab/server/conf"
	"jiacrontab/server/handle"
	"jiacrontab/server/model"
	"log"
	_ "net/http/pprof"

	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"

	"jiacrontab/libs"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/kataras/iris"
)

const (
	DefaultLayout = "layouts/layout.html"
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	db.CreateDB("sqlite3", "data/jiacrontab_server.db")
	db.DB().CreateTable(&db.Client{})
	db.DB().AutoMigrate(&db.Client{})

	// mail
	if conf.MailService.Enabled {
		mailer.InitMailer(&mailer.Mailer{
			QueueLength:    conf.MailService.QueueLength,
			SubjectPrefix:  conf.MailService.SubjectPrefix,
			From:           conf.MailService.From,
			Host:           conf.MailService.Host,
			User:           conf.MailService.User,
			Passwd:         conf.MailService.Passwd,
			FromEmail:      conf.MailService.FromEmail,
			DisableHelo:    conf.MailService.DisableHelo,
			HeloHostname:   conf.MailService.HeloHostname,
			SkipVerify:     conf.MailService.SkipVerify,
			UseCertificate: conf.MailService.UseCertificate,
			CertFile:       conf.MailService.CertFile,
			KeyFile:        conf.MailService.KeyFile,
			UsePlainText:   conf.MailService.UsePlainText,
			HookMode:       false,
		})
	}

	model.InitStore(conf.AppService.DataFile)

	app := iris.New()
	if conf.AppService.Debug {
		app.Logger().SetLevel("debug")
		app.Use(logger.New())
	}

	app.Use(recover.New())
	html := iris.HTML(conf.AppService.TplDir, conf.AppService.TplExt)
	html.AddFunc("date", libs.Date)
	html.AddFunc("formatMs", libs.Int2floatstr)
	html.Layout("layouts/layout.html")
	html.Reload(true)
	app.RegisterView(html)
	router(app)
	go rpc.ListenAndServe(conf.AppService.RpcListenAddr, &handle.Logic{})
	app.Run(iris.Addr(conf.AppService.HttpListenAddr))
}
