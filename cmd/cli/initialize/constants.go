package initialize

const generate_yaml = `projectname: {{.ProjectName}}
input: {{.Input}}
output: {{.Output}}
docs: {{.Docs}}
{{if .Root }}root: {{.Root}} {{end}}
env:
{{range $key, $value := .Env}}  - {{$value}} {{end}}
`

const resource_proto_common = `syntax = "proto3";
package {{ .ProjectName }}.api.v1;
option go_package = "{{ .Root }}/pkg/api/v1";
`

const resource_proto = `import "gogoproto/gogo.proto";
option (gogoproto.equal_all) = true;

import "github.com/solo-io/solo-kit/api/v1/metadata.proto";
import "github.com/solo-io/solo-kit/api/v1/status.proto";

/*
@solo-kit:resource
@solo-kit:resource.short_name={{ .ShortName }}
@solo-kit:resource.plural_name={{ .PluralName }}


// TODO: place your comments here
 */
message {{ .ResourceName }} {
    // The Resource-specific config is called a spec.
    {{ .ResourceName }}Spec spec = 1;

    // Status indicates the validation status of the resource. Status is read-only by clients, and set during validation
    core.solo.io.Status status = 2 [(gogoproto.nullable) = false, (gogoproto.moretags) = "testdiff:\"ignore\""];

    // Metadata contains the object metadata for this resource
    core.solo.io.Metadata metadata = 3 [(gogoproto.nullable) = false];
}

// TODO: describe the {{ .ResourceName }}Spec
message {{ .ResourceName }}Spec {
	// TODO: add fields
}
`
