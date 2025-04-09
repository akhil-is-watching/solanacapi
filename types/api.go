package types

type CreateProjectRequest struct {
	ProjectName string `json:"projectName"`
}

type CompileRequest struct {
	ProjectName  string     `json:"projectName"`
	ProgramFiles [][]string `json:"programFiles"`
	TestFiles    [][]string `json:"testFiles"`
	ConfigFiles  [][]string `json:"configFiles"`
}

type CompileResponse struct {
	StdOut string `json:"stdout"`
	Error  string `json:"error"`
}
