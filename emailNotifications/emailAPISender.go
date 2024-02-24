package emailNotifications

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"log"
	"os"
	"oysterProject/model"
	"oysterProject/utils"
	"strings"
)

const (
	mentorFilledQuestionsTemplateID        = "d-acb26f6c4b9a41309e281021ebdafe77"
	menteeFilledQuestionsTemplateID        = "d-be156416c7874548a5d40ab22c6c448f"
	mentorSessionCreatedTemplateID         = "d-e67f8b8a8373472089bfe585a2119388"
	menteeSessionCreatedFreeTemplateID     = "d-747f16f016924f30ae6fa431df681ad2"
	menteeSessionCreatedDonationTemplateID = "d-ce0dc32b002c446d9ef1e9e057bf4e31"
	menteeSessionCreatedPaidTemplateID     = "d-13aec49e40c6406da7c73591c3423806"
	menteeSessionConfirmedTemplateID       = "d-ae48b32f57c34243a9e6dfcc9f08dbba"
	mentorSessionConfirmedTemplateID       = "d-734268aa40dd4916bde04cbde8e63bc9"
	menteeSessionRescheduledTemplateID     = "d-ab0511265c0646d58acc823ba92a3376"
	mentorSessionRescheduledTemplateID     = "d-e1bf34de616946f1afc738bd898901b0"
	sessionMenteeNotificationTemplateID    = "d-cee7fe5b36084acc8274d586384d90a2"
	sessionMentorNotificationTemplateID    = "d-3a115ebf9fe94d57bc165a70669e0bde"
)

var (
	client    *sendgrid.Client
	emailFrom = mail.NewEmail("Oyster", "info@oystermentors.com")
)

func InitMailClient() {
	client = sendgrid.NewSendClient(os.Getenv("SEND_GRID_KEY"))
}

func sendEmailMessage(message *mail.SGMailV3) {
	response, err := client.Send(message)
	if err != nil {
		log.Println("Failed to send email:", err)
		return
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		log.Println("Email sent successfully")
	} else {
		log.Println("Failed to send email. Status:", response.StatusCode)
	}
}

func sendTemplateEmail(templateID, toName, toEmail string, dynamicTemplateData map[string]any) {
	message := mail.NewV3Mail()
	personalization := mail.NewPersonalization()
	personalization.DynamicTemplateData = dynamicTemplateData
	personalization.AddTos(mail.NewEmail(toName, toEmail))
	personalization.SetHeader("Importance", "high")

	message.SetTemplateID(templateID)
	message.AddPersonalizations(personalization)
	message.SetFrom(emailFrom)

	sendEmailMessage(message)
}

func SendUserFilledQuestionsEmail(user *model.User) {
	templateID := menteeFilledQuestionsTemplateID
	if user.AsMentor {
		templateID = mentorFilledQuestionsTemplateID
	}

	dynamicTemplateData := map[string]any{
		"name": user.Username,
	}
	sendTemplateEmail(templateID, user.Username, user.Email, dynamicTemplateData)
}

func SendSessionWasCreatedEmail(session *model.SessionResponse) {
	dynamicTemplateData := map[string]any{
		"mentorName":  session.Mentor.Name,
		"menteeName":  session.Mentee.Name,
		"sessionDate": session.SessionTimeStart.Format(utils.DateLayout),
		"sessionTime": session.SessionTimeStart.Format(utils.TimeLayout),
		"price":       session.PaymentDetails,
	}
	sendTemplateEmail(mentorSessionCreatedTemplateID, session.Mentor.Name, session.Mentor.Email, dynamicTemplateData)
	var templateId string
	if strings.EqualFold(session.PaymentDetails, "free") {
		templateId = menteeSessionCreatedFreeTemplateID
	} else if strings.EqualFold(session.PaymentDetails, "donation") {
		templateId = menteeSessionCreatedDonationTemplateID
	} else {
		templateId = menteeSessionCreatedPaidTemplateID
	}
	sendTemplateEmail(templateId, session.Mentee.Name, session.Mentee.Email, dynamicTemplateData)
}

func SendSessionConfirmedEmail(session *model.SessionResponse) {
	dynamicTemplateData := map[string]any{
		"mentorName":  session.Mentor.Name,
		"menteeName":  session.Mentee.Name,
		"sessionDate": session.SessionTimeStart.Format(utils.DateLayout),
		"sessionTime": session.SessionTimeStart.Format(utils.TimeLayout),
	}
	sendTemplateEmail(mentorSessionConfirmedTemplateID, session.Mentor.Name, session.Mentor.Email, dynamicTemplateData)
	sendTemplateEmail(menteeSessionConfirmedTemplateID, session.Mentee.Name, session.Mentee.Email, dynamicTemplateData)
}

func SendSessionRescheduledEmail(session *model.SessionResponse) {
	dynamicTemplateData := map[string]any{
		"mentorName":  session.Mentor.Name,
		"menteeName":  session.Mentee.Name,
		"sessionDate": session.NewSessionTimeStart.Format(utils.DateLayout),
		"sessionTime": session.NewSessionTimeStart.Format(utils.TimeLayout),
	}

	templateID := mentorSessionRescheduledTemplateID
	toName := session.Mentor.Name
	toEmail := session.Mentor.Email

	if session.SessionStatus == model.ReschedulingByMentee {
		templateID = menteeSessionRescheduledTemplateID
		toName = session.Mentee.Name
		toEmail = session.Mentee.Email
	} else {
		log.Printf("Wrong session status to send rescheduled email. Session id:%s, status:%s", session.SessionId, session.SessionStatus)
		return
	}
	sendTemplateEmail(templateID, toName, toEmail, dynamicTemplateData)
}

func SendNotificationBeforeSession(session *model.SessionNotification) {
	dynamicTemplateData := map[string]any{
		"mentorName":     session.MentorName,
		"menteeName":     session.MenteeName,
		"meetingLink":    session.MeetingLink,
		"paymentDetails": session.PaymentDetails,
	}

	sendTemplateEmail(sessionMenteeNotificationTemplateID, session.MenteeName, session.MenteeEmail, dynamicTemplateData)
	sendTemplateEmail(sessionMentorNotificationTemplateID, session.MentorName, session.MentorEmail, dynamicTemplateData)
}
