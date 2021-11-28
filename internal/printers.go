package internal

import (
	"fmt"
	"os"
)

func printSvcYAML(svc *Service) {
	fmt.Fprintf(os.Stderr, "- package: %s\n", svc.PackageName)
	fmt.Fprintf(os.Stderr, "  service: %s\n", svc.Name)

	// Print messages.
	fmt.Fprintf(os.Stderr, "  messages:\n")
	for _, msg := range svc.Messages {
		fmt.Fprintf(os.Stderr, "  - name: %s\n", msg.Name)
		fmt.Fprintf(os.Stderr, "    isExternal: %v\n", msg.IsExternal)
		fmt.Fprintf(os.Stderr, "    isOneOf: %v\n", msg.IsOneOf)

		// Print fields.
		fmt.Fprintf(os.Stderr, "    fields:\n")
		for _, f := range msg.Fields {
			fmt.Fprintf(os.Stderr, "    - name: %s\n", f.Name)
			fmt.Fprintf(os.Stderr, "      type: %v\n", f.Type)
			fmt.Fprintf(os.Stderr, "      isMatch: %v\n", f.IsMatch)
			fmt.Fprintf(os.Stderr, "      isMessage: %v\n", f.IsMessage)
			fmt.Fprintf(os.Stderr, "      isEnum: %v\n", f.IsEnum)

			if f.IsEnum {
				fmt.Fprintf(os.Stderr, "      enumValues:\n")
				for _, v := range f.EnumValues {
					fmt.Fprintf(os.Stderr, "      - %s\n", v.Name)
				}
			}
		}
	}

	fmt.Fprintf(os.Stderr, "  methods:\n")
	for _, m := range svc.Methods {
		fmt.Fprintf(os.Stderr, "  - name: %s\n", m.Name)
	}
}

func printYAML(allPackages []*Package) {
	for _, pkg := range allPackages {
		fmt.Fprintf(os.Stderr, "- package: %s\n", pkg.Name)
		fmt.Fprintf(os.Stderr, "  protoPackage: %s\n", pkg.ProtoName)
		fmt.Fprintf(os.Stderr, "  service: %s\n", pkg.Service.GoName)

		fmt.Fprintln(os.Stderr, "  messages:")
		for _, message := range pkg.Messages {
			fmt.Fprintf(os.Stderr, "  - name: %s\n", message.GoIdent.GoName)
			fmt.Fprintln(os.Stderr, "    fields:")
			for _, field := range message.Fields {
				fmt.Fprintf(os.Stderr, "    - name: %s\n", field.GoName)
				fmt.Fprintf(os.Stderr, "      kind: %s\n", field.Desc.Kind())

				// Enums are straightforward
				fmt.Fprintf(os.Stderr, "      isEnum: %v\n", field.Enum != nil)
				if field.Enum != nil {
					fmt.Fprintf(os.Stderr, "      enumValues:\n")
					for _, value := range field.Enum.Values {
						fmt.Fprintf(os.Stderr, "      - value: %s\n", value.GoIdent.GoName)
					}
				}

				fmt.Fprintf(os.Stderr, "      isMessage: %v\n", field.Message != nil)
				if field.Message != nil {
					fmt.Fprintf(os.Stderr, "      message:\n")
					fmt.Fprintf(os.Stderr, "        name: %s\n", field.Message.GoIdent.GoName)
					fmt.Fprintf(os.Stderr, "        import: %s\n", string(field.Message.GoIdent.GoImportPath))
					fmt.Fprintf(os.Stderr, "        isOneOf: %v\n", len(field.Message.Oneofs) > 0)
					if len(field.Message.Oneofs) > 0 {
						// Oneofs are complicated because they are a field that
						// is a message. The message has "oneofs" and each oneof
						// has fields that are messages.
						fmt.Fprintf(os.Stderr, "        oneOfs:\n")
						for _, oneof := range field.Message.Oneofs {
							fmt.Fprintf(os.Stderr, "        - name: %s:\n", oneof.GoName)

							fmt.Fprintf(os.Stderr, "          fields:\n")
							for _, field := range oneof.Fields {
								fmt.Fprintf(os.Stderr, "          - name: %s\n", field.GoName)
								fmt.Fprintf(os.Stderr, "            type: %s\n", field.GoIdent.GoName)
								fmt.Fprintf(os.Stderr, "            message: %s\n", field.Message.GoIdent.GoName)
							}
						}
					}
				}
			}
		}
	}
}
