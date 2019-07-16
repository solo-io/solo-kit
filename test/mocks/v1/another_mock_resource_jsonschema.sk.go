// Code generated by solo-kit. DO NOT EDIT.

package v1

var AnotherMockResourceJsonSchema = `
{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "$ref": "#/definitions/testing.solo.io.v1.AnotherMockResource",
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
        "core.solo.io.v1.Status": {
            "$schema": "http://json-schema.org/draft-04/schema#",
            "properties": {
                "reason": {
                    "type": "string"
                },
                "reportedBy": {
                    "type": "string"
                },
                "state": {
                    "enum": [
                        "Pending",
                        0,
                        "Accepted",
                        1,
                        "Rejected",
                        2
                    ],
                    "oneOf": [
                        {
                            "type": "string"
                        },
                        {
                            "type": "integer"
                        }
                    ]
                },
                "subresourceStatuses": {
                    "additionalProperties": true,
                    "type": "object",
                    "title": "core.solo.io.v1.Status.SubresourceStatusesEntry"
                }
            },
            "type": "object",
            "title": "core.solo.io.v1.Status"
        },
        "testing.solo.io.v1.AnotherMockResource": {
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
                        "basicField": {
                            "type": "string"
                        }
                    },
                    "type": "object",
                    "title": "spec"
                },
                "status": {
                    "$ref": "#/definitions/core.solo.io.v1.Status",
                    "additionalProperties": true,
                    "type": "object",
                    "title": "core.solo.io.v1.Status"
                }
            },
            "title": "testing.solo.io.v1.AnotherMockResource"
        }
    }
}
`
