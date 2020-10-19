package domain

//ExecutionLog represents the log w.r.t the REST API executions
type ExecutionLog struct {
	ID        uint
	UUID      string
	Command   string
	Params    string
	Status    string
	StartTime string
	EndTime   string
	Duration  string
}
