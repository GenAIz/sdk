package dataset

type CreateFields struct {
	WorkspaceId *int64
	Description string
	Handle      string
	Name        string
}

type ListFields struct {
	DataSetId *int64
	Path      string
}

type UploadFields struct {
	DataSetId       *int64
	DestinationPath string
	FileName        string
	FilePath        string
}
