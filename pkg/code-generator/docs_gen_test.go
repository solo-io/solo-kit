package code_generator_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
)

var _ = Describe("DocsGen", func() {

	const (
		testProtoName         = "doc_gen_test.proto"
		testProjectConfigName = "solo-kit.json"
	)

	var tempDir string

	BeforeEach(func() {

		// Create temp directory and path variables
		workingDir, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		projectRoot := filepath.Join(workingDir, "../../")
		tempDir, err = ioutil.TempDir(projectRoot, "doc-gen-test-")
		Expect(err).NotTo(HaveOccurred())
		relativePathToTempDir := filepath.Join("github.com/solo-io/solo-kit", filepath.Base(tempDir))

		// Generate test proto file with two messages
		buf := &bytes.Buffer{}
		err = testProtoTemplate().Execute(buf, relativePathToTempDir)
		Expect(err).NotTo(HaveOccurred())
		err = ioutil.WriteFile(filepath.Join(tempDir, testProtoName), []byte(buf.String()), 0644)
		Expect(err).NotTo(HaveOccurred())

		// Generate project config
		buf = &bytes.Buffer{}
		err = projectConfigFile().Execute(buf, "docs")
		Expect(err).NotTo(HaveOccurred())
		err = ioutil.WriteFile(filepath.Join(tempDir, testProjectConfigName), []byte(buf.String()), 0644)
		Expect(err).NotTo(HaveOccurred())

		// Run code gen
		err = cmd.Run(tempDir, true, true, nil, nil)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tempDir)).NotTo(HaveOccurred())
	})

	It("docs for a message are generated based on the value of the skip_docs_gen option", func() {

		// Collect all the generated docs
		foundExpectedDoc := false
		err := filepath.Walk(tempDir+"/docs", func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {

				// No file must contain any reference to DoNotGenerateDocsForMe
				file, err := ioutil.ReadFile(path)
				Expect(err).NotTo(HaveOccurred())
				matched, err := regexp.Match("(?i)DoNotGenerateDocsForMe", file)
				Expect(err).NotTo(HaveOccurred())
				Expect(matched).To(BeFalse())

				// Verify that GenerateDocsForMe appears in at least one of the generated docs
				matched, err = regexp.Match("(?i)GenerateDocsForMe", file)
				Expect(err).NotTo(HaveOccurred())
				if matched {
					foundExpectedDoc = true
				}

			}
			return nil
		})
		Expect(foundExpectedDoc).To(BeTrue())
		Expect(err).NotTo(HaveOccurred())
	})

})

func testProtoTemplate() *template.Template {
	return template.Must(template.New("testProtoTemplate").Parse(`

syntax = "proto3";

package testing.solo.io;
option go_package = "{{.}}";

import "gogoproto/gogo.proto";
option (gogoproto.equal_all) = true;

import "github.com/solo-io/solo-kit/api/v1/metadata.proto";
import "github.com/solo-io/solo-kit/api/v1/status.proto";
import "github.com/solo-io/solo-kit/api/v1/solo-kit.proto";

message GenerateDocsForMe {
    option (core.solo.io.resource).short_name = "docs";
    option (core.solo.io.resource).plural_name = "generatedocsforme";
    core.solo.io.Metadata metadata = 1 [(gogoproto.nullable) = false];
    core.solo.io.Status status = 6 [(gogoproto.nullable) = false];

    // Some field
    string basic_field = 2;

}

message DoNotGenerateDocsForMe {
    option (core.solo.io.resource).short_name = "nodocs";
    option (core.solo.io.resource).plural_name = "donotgeneratedocsforme";
    option (core.solo.io.resource).skip_docs_gen = true;
    core.solo.io.Metadata metadata = 1 [(gogoproto.nullable) = false];
    core.solo.io.Status status = 6 [(gogoproto.nullable) = false];

    // Some field
    string basic_field = 2;
}

`))
}

func projectConfigFile() *template.Template {
	return template.Must(template.New("").Parse(`

{
  "title": "Solo-Kit Testing",
  "name": "testing.solo.io",
  "version": "v1",
  "docs_dir": "{{.}}"
}

`))
}
