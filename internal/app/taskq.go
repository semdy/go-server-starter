package app

import (
	"go-server-starter/pkg/notify"
	"go-server-starter/pkg/taskq"

	"go.uber.org/zap"
)

// 初始化通知发送器（优先阿里云，失败则降级为日志输出）
func (a *App) initNotify() (notify.SmsSender, notify.EmailSender) {
	var sms notify.SmsSender = notify.LogSender{}
	var email notify.EmailSender = notify.LogSender{}

	if a.config.AlibabaCloud.AccessKeyID != "" {
		if s, err := notify.NewAlibabaSmsSender(
			a.config.AlibabaCloud.AccessKeyID,
			a.config.AlibabaCloud.AccessKeySecret,
			a.logger.Named("ALISMS"),
		); err == nil {
			sms = s
		}
		if s, err := notify.NewAlibabaEmailSender(
			a.config.AlibabaCloud.AccessKeyID,
			a.config.AlibabaCloud.AccessKeySecret,
			a.config.AlibabaCloud.Email.FromAddress,
			a.config.AlibabaCloud.Email.FromName,
			a.logger.Named("ALIMAIL"),
		); err == nil {
			email = s
		}
	}
	return sms, email
}

// 初始化任务队列
func (a *App) initTaskQueue(smsSender notify.SmsSender, emailSender notify.EmailSender) error {
	client, err := taskq.NewClient(a.config.AsynQ, a.logger.Named("TASKQ-CLIENT"))
	if err != nil {
		return err
	}
	a.taskqClient = client

	server := taskq.NewServer(a.config.AsynQ, taskq.ServerConfig{
		Concurrency: a.config.AsynQ.Concurrency,
		Queues: map[string]int{
			"default": 3,
			"low":     1,
		},
	}, a.logger, nil) // nil alerter = default no-op

	// 注入任务处理器依赖
	taskq.HandlerDeps.SmsSender = smsSender
	taskq.HandlerDeps.EmailSender = emailSender
	taskq.HandlerDeps.SMSSignName = a.config.AlibabaCloud.SMS.SignName
	taskq.HandlerDeps.SMSTemplateCode = a.config.AlibabaCloud.SMS.TemplateCode

	// 注册任务处理器
	server.HandleFunc(taskq.TaskEmailWelcome, taskq.HandleEmailWelcome)
	server.HandleFunc(taskq.TaskSendSMSCode, taskq.HandleSendSMSCode)
	server.HandleFunc(taskq.TaskSendEmailCode, taskq.HandleSendEmailCode)

	// 启动后台 worker
	go func() {
		if err := server.Start(); err != nil {
			a.logger.Error("taskq worker stopped with error", zap.Error(err))
		}
	}()

	a.taskqServer = server

	return nil
}

// gracefully stop the task worker and client.
func (a *App) shutdownTaskQueue() {
	if a.taskqServer != nil {
		a.taskqServer.Shutdown()
	}
	if a.taskqClient != nil {
		if err := a.taskqClient.Close(); err != nil {
			a.logger.Error("Failed to close taskq client", zap.Error(err))
		}
	}
}
