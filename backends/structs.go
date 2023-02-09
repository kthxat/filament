package backends

import (
	"errors"
	"reflect"

	"github.com/spf13/viper"
)

var ErrUnsupportedOperation = errors.New("unsupported operation")

type BackendDescriptor struct {
	ID, DisplayName string
	Type            reflect.Type
	New             func(params *BackendConstructionParams) (Backend, error)
}

type BackendConstructionParams struct {
	Config *viper.Viper
}
