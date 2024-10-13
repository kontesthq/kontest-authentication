package kafka_utils

type KafkaTopic struct {
	envVarName   string
	DefaultValue string
}

var (
	UserRegistrationEventTopic     = KafkaTopic{"KAFKA_USER_REGISTRATION_EVENT_TOPIC", "user-registration-event"}
	AccountDeletionEventTopic      = KafkaTopic{"KAFKA_ACCOUNT_DELETION_EVENT_TOPIC", "account-deletion-event"}
	AccountDeletionEmailEventTopic = KafkaTopic{"KAFKA_ACCOUNT_DELETION_EMAIL_EVENT_TOPIC", "account-deletion-email-event"}
	PasswordChangeEmailEventTopic  = KafkaTopic{"KAFKA_PASSWORD_CHANGE_EMAIL_EVENT_TOPIC", "password-change-email-event"}
)
