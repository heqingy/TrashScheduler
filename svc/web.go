package svc

type incomeSMSBody struct {
	ToCountry           *string
	ToState             *string
	SmsMessageSid       *string
	NumMedia            *string
	ToCity              *string
	FromZip             *string
	SmsSid              *string
	FromState           *string
	SmsStatus           *string
	FromCity            *string
	FromCountry         *string
	To                  *string
	MessagingServiceSid *string
	ToZip               *string
	NumSegments         *string
	MessageSid          *string
	AccountSid          *string
	ApiVersion          *string

	// We only care these two
	From string
	Body string
}
