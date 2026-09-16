package yarn

type Applications struct {
	Apps struct {
		App []Application `json:"app"`
	} `json:"apps"`
}

type Application struct {
	Id              string           `json:"id"`
	User            string           `json:"user"`
	QueueUser       string           `json:"queueUser"`
	Name            string           `json:"name"`
	Queue           string           `json:"queue"`
	State           ApplicationState `json:"state"`
	FinalStatus     string           `json:"finalStatus"` // UNDEFINED, SUCCEEDED, FAILED, KILLED
	Progress        float32          `json:"progress"`
	ApplicationType string           `json:"applicationType"`
	ApplicationTags string           `json:"applicationTags"`
	StartedTime     int64            `json:"startedTime"`   // epoch millis
	FinishedTime    int64            `json:"finishedTime"`  // epoch millis
	ElapsedTime     int64            `json:"elapsedTime"`   // milliseconds
	MemorySeconds   int64            `json:"memorySeconds"` // megabytes
	VcoreSeconds    int64            `json:"vcoreSeconds"`
}
