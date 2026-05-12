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

type Applications struct {
	Apps struct{ App []Application }
}

type Application struct {
	Id              string
	User            string
	Name            string
	Queue           string
	State           ApplicationState
	Progress        float64
	ApplicationType string
	ApplicationTags string
	StartedTime     int64
	FinishedTime    int64
	ElapsedTime     int64
	// MemorySeconds   int64
	// VcoreSeconds    int64
}
