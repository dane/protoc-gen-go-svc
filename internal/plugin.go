package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/dane/protoc-gen-go-svc/internal/options"
)

const (
	FileName = "service.pb.go"
)

type Plugin struct {
	Verbose            bool
	PrivatePackageName string
}

type Package struct {
	ProtoName         protoreflect.FullName
	Name              protogen.GoPackageName
	ImportPath        protogen.GoImportPath
	ServiceImportPath protogen.GoImportPath
	Service           *protogen.Service
	Messages          []*protogen.Message
}

func (p *Plugin) Run(plugin *protogen.Plugin) error {
	privatePackageName := protoreflect.FullName(p.PrivatePackageName)

	var (
		servicePackageName string
		serviceImportPath  string
		goPackage          string
	)

	// Group service, package name, and import path as a Package. Grouping is
	// managed with a map for easy lookups later on.
	packages := make(map[protoreflect.FullName]*Package)
	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}

		// Create a package if one does not exist.
		key := file.Desc.Package()
		pkg, ok := packages[key]
		if !ok {
			pkg = &Package{
				Name:       file.GoPackageName,
				ProtoName:  file.Desc.Package(),
				ImportPath: file.GoImportPath,
			}

			if value := options.GoPackage(file); value != "" {
				if goPackage == "" {
					goPackage = value
				}

				if goPackage != value {
					return NewErrBadServiceImportPath(file, goPackage, value)
				}

				opt := strings.SplitN(value, ";", 2)
				if len(opt) != 2 {
					return NewErrInvalidServiceImportPath(file, value)
				}

				serviceImportPath = opt[0]
				servicePackageName = opt[1]

				pkg.ServiceImportPath = protogen.GoImportPath(serviceImportPath)
			}

			packages[key] = pkg
		}

		// Assign the service to the package if it is defined in this file.
		for _, service := range file.Services {
			// This assumes there is one service per package. Protobufs support
			// multiple services, based on the chaining process, a single
			// private service would need to implement all public service RPCs.
			pkg.Service = service
		}

		// Assign the messages found in this file to the pkg.
		pkg.Messages = append(pkg.Messages, file.Messages...)
	}

	// Ensure a private service is present.
	privatePackage, ok := packages[privatePackageName]
	if !ok {
		return NewErrPrivatePackageNotFound(privatePackageName)
	}

	// Convert map of packages to sorted slice. Exclude the private package from
	// the slice, sort, and prepend it after. This is to ensure the private
	// package is always first in the slice.
	var allPackages []*Package
	for _, pkg := range packages {
		if pkg == privatePackage {
			continue
		}

		allPackages = append(allPackages, pkg)
	}

	sort.Slice(allPackages, func(a, b int) bool {
		return allPackages[a].ProtoName > allPackages[b].ProtoName
	})

	allPackages = append([]*Package{privatePackage}, allPackages...)

	// Create services in order of private service then public services in
	// decending order.
	var svcChain []*Service

	if p.Verbose {
		defer func() {
			if len(svcChain) == 0 {
				return
			}

			// Output the service package name and the complete service in JSON
			// format for debugging purposes. The each service will contain the
			// subsequent services in the chain.
			svc := svcChain[len(svcChain)-1]
			fmt.Fprintf(os.Stderr, ">> %s\n", svc.ProtoPackageName)
			enc := json.NewEncoder(os.Stderr)
			enc.SetIndent("", "    ")
			_ = enc.Encode(svc)
		}()
	}

	for _, pkg := range allPackages {
		svc, err := NewService(
			pkg.ProtoName,
			pkg.Name,
			pkg.ImportPath,
			pkg.ServiceImportPath,
			pkg.Service,
			pkg.Messages,
			svcChain,
		)

		if err != nil {
			return err
		}

		svcChain = append(svcChain, svc)

		// Write service file.
		importPath := protogen.GoImportPath(path.Join(servicePackageName, svc.PackageName))
		fileName := path.Join(servicePackageName, svc.PackageName, FileName)

		file := plugin.NewGeneratedFile(fileName, importPath)
		if err := render(file, svc.ProtoPackageName, serviceTemplate, svc); err != nil {
			return err
		}

		// Write testing service file.
		importPath = protogen.GoImportPath(path.Join(servicePackageName, svc.PackageName, "testing"))
		fileName = path.Join(servicePackageName, svc.PackageName, "testing", FileName)

		if !svc.IsPrivate {
			file = plugin.NewGeneratedFile(fileName, importPath)
			if err := render(file, "testing", testingTemplate, svc); err != nil {
				return err
			}
		}
	}

	// Write services register wrapper file.
	importPath := protogen.GoImportPath(serviceImportPath)
	fileName := path.Join(serviceImportPath, FileName)
	file := plugin.NewGeneratedFile(fileName, importPath)
	return render(file, "register", registerTemplate, RegisterService{
		PackageName: servicePackageName,
		Private:     svcChain[0],
		Services:    svcChain[1:],
	})
}
