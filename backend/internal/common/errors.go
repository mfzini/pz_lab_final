package common

type ApiError struct {
	Err        string
	StatusCode int
}

func (self *ApiError) Error() string {
	return self.Err
}
