# protoc-gen-go-svc/example

This is an example service where all files, except for the private service `Impl` methods are auto-generated.

## Versioning Rules

The following are rules the protoc plugin must adhere to:

- Services are versioned in a format of v1 or date-based as `YYYYMMDD`
- Versioned services delegate to the next version in assending order
- The lastest version delegates to the private service
- RPC delegation assumes identical RPC name unless an override is provided
- Input message fields assume identically named fields in the delegation unless
  an override is provided
- Deprecated fields are written directly to the private service message
- Input fields introduced in successive service versions have default values for
  the cases where the RPC of a previous version was called
- Output message fields assume identically named fields in message propagation
  unless an override is provided

### Conversion Function Naming

convert.ToNextCreateRequest()
convert.ToPrivateCreateRequest()
convert.ToPublicCreateRequest()
