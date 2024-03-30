package irisutils

// Standard response struct
type Resp struct {
	Result bool `json:"result"`
	Data   any  `json:"data"`
}
