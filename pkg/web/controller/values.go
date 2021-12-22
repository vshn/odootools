package controller

// Values is an arbitrary tree of data to be passed to Template rendering.
type Values map[string]interface{}

func AsError(err error) Values {
	return Values{
		"Error": err.Error(),
	}
}
