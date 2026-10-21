package email

import (
	"fmt"

	"github.com/jpabloVega/dokodemoYomou/secure"
	"github.com/spf13/viper"
	"github.com/wneessen/go-mail"
)

func MailToKindle(attachments []string) error {
	sender := viper.GetString("usermail")
	var password string
	var err error
	if sender == "" {
		sender, password, err = secure.NewEmailPassword()
		if err != nil {
			return err
		}
	} else {
		password, err = secure.GetPassword(sender)
		if err != nil {
			return err
		}
	}

	reciever := viper.GetString("kindlemail")
	if reciever == "" {
		reciever, err = secure.SetKindleEmail()
		if err != nil {
			return err
		}
	}

	message := mail.NewMsg()

	err = message.From(sender)
	if err != nil {
		return err
	}

	err = message.To(reciever)
	if err != nil {
		return err
	}

	message.Subject("convert")
	message.SetBodyString(mail.TypeTextPlain, "")

	for _, attachment := range attachments {
		message.AttachFile(attachment)
	}

	client, err := mail.NewClient("smtp.gmail.com",
		mail.WithPort(587),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(sender),
		mail.WithPassword(password),
		mail.WithTLSPolicy(mail.TLSMandatory),
	)
	if err != nil {
		return err
	}

	err = client.DialAndSend(message)
	if err != nil {
		return err
	}

	fmt.Println("Emailed successfully")
	return nil
}
