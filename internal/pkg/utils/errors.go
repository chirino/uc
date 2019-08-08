package utils

import (
	"fmt"
	"os"
)

type errors []error

func Errors(e ...error) error {
	switch len(e) {
	case 0:
		return nil
	case 1:
		return e[0]
	}
	result := errors(e)
	return &result
}

func (me *errors) Error() (err string) {
	err = fmt.Sprintf("%d errors: ", len(*me))
	for i, e := range *me {
		if i != 0 {
			err += ", "
		}
		err += fmt.Sprintf("#%d: %s", i+1, e.Error())
	}
	return err
}

func ExitOnError(err error) {
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
