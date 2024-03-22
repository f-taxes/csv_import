package global

type ImportFile struct {
	Account  string
	File     string
	RealName string
	Schema   string
}

type Processor interface {
	Parse(contents []byte, account, fileName string)
}

type SchemaProcessor struct {
	Label     string    `json:"label"`
	Name      string    `json:"name"`
	Processor Processor `json:"-"`
}
