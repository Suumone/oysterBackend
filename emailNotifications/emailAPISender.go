package emailNotifications

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"log"
	"os"
	"oysterProject/model"
)

var client *sendgrid.Client

func CreateMailClient() {
	client = sendgrid.NewSendClient(os.Getenv("SEND_GRID_KEY"))
}

var templateIdMap = map[string]string{
	"menteeRegistered": "d-be156416c7874548a5d40ab22c6c448f",
}

func SendUserRegisteredEmail(user *model.User, template string) {
	from := mail.NewEmail("The Oyster", "info@oystermentors.com")
	to := mail.NewEmail("", user.Email)

	//dynamicTemplateData := map[string]interface{}{
	//	"name": "John Doe",
	//}
	personalization := mail.NewPersonalization()
	//personalization.DynamicTemplateData = dynamicTemplateData
	personalization.AddTos(to)

	message := mail.NewV3Mail()
	templateID := templateIdMap[template]
	message.SetTemplateID(templateID)
	message.AddPersonalizations(personalization)
	message.SetFrom(from)

	response, err := client.Send(message)
	if err != nil {
		return
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		log.Println("Email sent successfully")
	} else {
		log.Println("Failed to send email. Status:", response.StatusCode)
	}
}
