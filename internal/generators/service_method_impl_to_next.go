package generators

import "io"

func NewServiceMethodImplToNext(methodName, inputName, outputName, nextMethodName, nextInputName, privateInputName, privateOutputName string, deprecatdFields []DeprecatedField) Generator {
	return GeneratorFunc(func(w io.Writer) error {
		return execute("service_method_impl_to_next", templateServiceMethodImplToNext, w, serviceMethodImplToNext{
			MethodName:        methodName,
			InputName:         inputName,
			OutputName:        outputName,
			PrivateInputName:  privateInputName,
			PrivateOutputName: privateOutputName,
			NextMethodName:    nextMethodName,
			NextInputName:     nextInputName,
			DeprecatedFields:  deprecatdFields,
		})
	})
}

type serviceMethodImplToNext struct {
	MethodName        string
	InputName         string
	OutputName        string
	PrivateInputName  string
	PrivateOutputName string
	NextMethodName    string
	NextInputName     string
	DeprecatedFields  []DeprecatedField
}

type DeprecatedField struct {
	FieldName        string
	PrivateFieldName string
}
