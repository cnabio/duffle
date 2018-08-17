package builder

// SummaryStatusCode is the enumeration of the possible status codes returned for a kubed up.
type SummaryStatusCode int

const (
	// SummaryUnknown means that the status of `kubed up` is in an unknown state.
	SummaryUnknown SummaryStatusCode = iota
	// SummaryLogging means that the status is currently gathering and exporting logs.
	SummaryLogging
	// SummaryStarted means that `kubed up` has begun.
	SummaryStarted
	// SummaryOngoing means that `kubed up` is ongoing and we are waiting for further information from the builder.
	SummaryOngoing
	// SummarySuccess means that `kubed up` has succeeded.
	SummarySuccess
	// SummaryFailure means that `kubed up` has failed. Usually this can be followed up by checking the build logs.
	SummaryFailure
)

// SummaryStatusCodeName is the relation between summary status code enums and their respective names.
var SummaryStatusCodeName = map[int]string{
	0: "UNKNOWN",
	1: "LOGGING",
	2: "STARTED",
	3: "ONGOING",
	4: "SUCCESS",
	5: "FAILURE",
}

// Summary is the message returned when executing a kubed up.
type Summary struct {
	// StageDesc describes the particular stage this summary
	// represents, e.g. "Build Docker Image." This is meant
	// to be a canonical summary of the stage's intent.
	StageDesc string `json:"stage_desc,omitempty"`
	// status_text indicates a string description of the progress
	// or completion of kubed up.
	StatusText string `json:"status_text,omitempty"`
	// status_code indicates the status of the progress or
	// completion of a kubed up.
	StatusCode SummaryStatusCode `json:"status_code,omitempty"`
	// build_id is the build identifier associated with this kubed up build.
	BuildID string `json:"build_id,omitempty"`
}
