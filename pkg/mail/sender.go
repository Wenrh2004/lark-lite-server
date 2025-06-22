package mail

import (
	"fmt"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

// Config 邮件配置
type Config struct {
	Host     string // SMTP服务器地址
	Port     int    // SMTP服务器端口
	Username string // 邮箱账号
	Password string // 邮箱密码
	From     string // 发件人地址
	FromName string // 发件人名称
	SSL      bool   // 是否使用SSL
}

// Sender 邮件发送器接口
type Sender interface {
	Send(to []string, subject string, body string, isHTML bool) error
}

// EmailSender 邮件发送器实现
type EmailSender struct {
	config *Config
}

// NewEmailSender 创建邮件发送器
func NewEmailSender(conf *viper.Viper) Sender {
	return &EmailSender{
		config: &Config{
			Host:     conf.GetString("app.mail.host"),
			Port:     conf.GetInt("app.mail.port"),
			Username: conf.GetString("app.mail.username"),
			Password: conf.GetString("app.mail.password"),
			From:     conf.GetString("app.mail.from"),
			FromName: conf.GetString("app.mail.from_name"),
			SSL:      conf.GetBool("app.mail.ssl"),
		},
	}
}

// Send 发送邮件
func (s *EmailSender) Send(to []string, subject string, body string, isHTML bool) error {
	m := gomail.NewMessage()

	// 设置发件人
	if s.config.FromName != "" {
		m.SetAddressHeader("From", s.config.From, s.config.FromName)
	} else {
		m.SetHeader("From", s.config.From)
	}

	// 设置收件人
	m.SetHeader("To", to...)

	// 设置主题
	m.SetHeader("Subject", subject)

	// 设置内容
	if isHTML {
		m.SetBody("text/html", body)
	} else {
		m.SetBody("text/plain", body)
	}

	// 配置邮件发送
	d := gomail.NewDialer(s.config.Host, s.config.Port, s.config.Username, s.config.Password)

	// 配置TLS
	if s.config.SSL {
		d.SSL = true
	} else {
		d.SSL = false
		// 如果端口是常见的TLS端口，启用TLS
		if s.config.Port == 587 {
			d.TLSConfig = nil // 使用默认TLS配置
		}
	}

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// TemplateData 渲染邮件模板的数据
type TemplateData map[string]interface{}

// 预定义的邮件模板
var (
	// VerificationTemplate 验证码邮件模板
	VerificationTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>验证码</title>
</head>
<body>
    <div style="max-width: 600px; margin: 0 auto; padding: 20px; font-family: Arial, sans-serif;">
        <div style="text-align: center; padding: 20px 0;">
            <h2>验证码</h2>
        </div>
        <div style="padding: 20px; background-color: #f7f7f7; border-radius: 5px;">
            <p>尊敬的用户：</p>
            <p>您的验证码是：<strong style="font-size: 24px; color: #007bff;">{{.Code}}</strong></p>
            <p>此验证码将在 {{.ExpiresIn}} 分钟后失效。</p>
            <p>如非本人操作，请忽略此邮件。</p>
        </div>
        <div style="text-align: center; color: #999; padding: 20px 0; font-size: 12px;">
            <p>此邮件由系统自动发送，请勿回复。</p>
        </div>
    </div>
</body>
</html>
`

	// WelcomeTemplate 欢迎邮件模板
	WelcomeTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>欢迎加入</title>
</head>
<body>
    <div style="max-width: 600px; margin: 0 auto; padding: 20px; font-family: Arial, sans-serif;">
        <div style="text-align: center; padding: 20px 0;">
            <h2>欢迎加入</h2>
        </div>
        <div style="padding: 20px; background-color: #f7f7f7; border-radius: 5px;">
            <p>尊敬的 {{.Username}}：</p>
            <p>欢迎您加入我们的平台！</p>
            <p>您的账号已成功注册，现在您可以使用我们的服务了。</p>
            <p>如有任何问题，请随时与我们联系。</p>
        </div>
        <div style="text-align: center; color: #999; padding: 20px 0; font-size: 12px;">
            <p>此邮件由系统自动发送，请勿回复。</p>
        </div>
    </div>
</body>
</html>
`
)
