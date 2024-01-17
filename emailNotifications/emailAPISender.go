package emailNotifications

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"log"
	"os"
	"oysterProject/model"
	"time"
)

const (
	MenteeRegisteredTemplateID      = "d-be156416c7874548a5d40ab22c6c448f"
	MentorRegisteredTemplateID      = "d-be156416c7874548a5d40ab22c6c448f"
	MentorFilledQuestionsTemplateID = "d-acb26f6c4b9a41309e281021ebdafe77"
	MenteeFilledQuestionsTemplateID = ""
	SessionSetUpForMentorTemplateID = "d-e67f8b8a8373472089bfe585a2119388"
)

var client *sendgrid.Client

func CreateMailClient() {
	client = sendgrid.NewSendClient(os.Getenv("SEND_GRID_KEY"))
}

func sendEmailMessage(message *mail.SGMailV3) {
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

func SendUserRegisteredEmail(user *model.User) {
	from := mail.NewEmail("The Oyster", "info@oystermentors.com")
	to := mail.NewEmail("", user.Email)

	//dynamicTemplateData := map[string]interface{}{
	//	"name": "John Doe",
	//}
	personalization := mail.NewPersonalization()
	//personalization.DynamicTemplateData = dynamicTemplateData
	personalization.AddTos(to)

	message := mail.NewV3Mail()
	if user.AsMentor {
		message.SetTemplateID(MentorRegisteredTemplateID)
	} else {
		message.SetTemplateID(MenteeRegisteredTemplateID)
	}
	message.AddPersonalizations(personalization)
	message.SetFrom(from)

	sendEmailMessage(message)
}

func SendUserFilledQuestionsEmail(user *model.User) {
	from := mail.NewEmail("The Oyster", "info@oystermentors.com")
	to := mail.NewEmail(user.Username, user.Email)

	dynamicTemplateData := map[string]interface{}{
		"name": user.Username,
	}
	personalization := mail.NewPersonalization()
	personalization.DynamicTemplateData = dynamicTemplateData
	personalization.AddTos(to)

	message := mail.NewV3Mail()
	if user.AsMentor {
		templateID := MentorFilledQuestionsTemplateID
		message.SetTemplateID(templateID)
	} else {
		templateID := MenteeFilledQuestionsTemplateID
		message.SetTemplateID(templateID)
	}
	message.AddPersonalizations(personalization)
	message.SetFrom(from)

	sendEmailMessage(message)
}

func SendSessionSetUpForMentorEmail(session *model.SessionResponse) {
	from := mail.NewEmail("The Oyster", "info@oystermentors.com")
	to := mail.NewEmail(session.Mentor.Name, session.Mentor.Name)

	dynamicTemplateData := map[string]interface{}{
		"mentorName":  session.Mentor.Name,
		"menteeName":  session.Mentee.Name,
		"sessionDate": session.SessionTimeStart.Format(time.DateOnly),
		"sessionTime": session.SessionTimeStart.Format(time.TimeOnly),
	}
	personalization := mail.NewPersonalization()
	personalization.DynamicTemplateData = dynamicTemplateData
	personalization.AddTos(to)
	personalization.SetHeader("Importance", "high")

	message := mail.NewV3Mail()
	message.SetTemplateID(SessionSetUpForMentorTemplateID)
	message.AddPersonalizations(personalization)
	message.SetFrom(from)

	sendEmailMessage(message)
}
