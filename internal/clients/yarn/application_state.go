package yarn

type ApplicationState string

const (
	NEW        ApplicationState = "NEW"
	NEW_SAVING ApplicationState = "NEW_SAVING"
	SUBMITTED  ApplicationState = "SUBMITTED"
	ACCEPTED   ApplicationState = "ACCEPTED"
	RUNNING    ApplicationState = "RUNNING"
	FINISHED   ApplicationState = "FINISHED"
	FAILED     ApplicationState = "FAILED"
	KILLED     ApplicationState = "KILLED"
)
