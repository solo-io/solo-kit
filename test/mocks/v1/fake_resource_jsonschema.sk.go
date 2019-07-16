// Code generated by solo-kit. DO NOT EDIT.

package v1

var FakeResourceJsonSchema = `
{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "$ref": "#/definitions/testing.solo.io.v1.FakeResource",
    "definitions": {
        "core.solo.io.v1.Metadata": {
            "$schema": "http://json-schema.org/draft-04/schema#",
            "properties": {
                "annotations": {
                    "additionalProperties": true,
                    "type": "object",
                    "title": "core.solo.io.v1.Metadata.AnnotationsEntry"
                },
                "cluster": {
                    "type": "string"
                },
                "labels": {
                    "additionalProperties": true,
                    "type": "object",
                    "title": "core.solo.io.v1.Metadata.LabelsEntry"
                },
                "name": {
                    "type": "string"
                },
                "namespace": {
                    "type": "string"
                },
                "resourceVersion": {
                    "type": "string"
                }
            },
            "type": "object",
            "title": "core.solo.io.v1.Metadata"
        },
        "testing.solo.io.v1.FakeResource": {
            "properties": {
                "metadata": {
                    "$ref": "#/definitions/core.solo.io.v1.Metadata",
                    "additionalProperties": true,
                    "type": "object",
                    "title": "core.solo.io.v1.Metadata"
                },
                "spec": {
                    "$schema": "http://json-schema.org/draft-04/schema#",
                    "properties": {
                        "count": {
                            "type": "integer"
                        }
                    },
                    "type": "object",
                    "title": "spec"
                }
            },
            "title": "testing.solo.io.v1.FakeResource"
        }
    }
}
`