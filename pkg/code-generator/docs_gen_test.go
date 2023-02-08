package code_generator_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/solo-io/anyvendor/anyvendor"
	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/datafile"
	"github.com/solo-io/solo-kit/pkg/code-generator/sk_anyvendor"
	"github.com/solo-io/solo-kit/pkg/utils/modutils"
	"gopkg.in/yaml.v2"

	"github.com/solo-io/solo-kit/pkg/code-generator/docgen/options"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/code-generator/cmd"
)

var _ = Describe("DocsGen", func() {

	const (
		testProtoName         = "doc_gen_test.proto"
		testProtoNoDocsName   = "no_doc_gen_test.proto"
		testProjectConfigName = "solo-kit.json"
		hugoApiDir            = "api"
		hugoDataDir           = "docs/data"
		projectConfigDocsDir  = "docs/content"
		packageName           = "github.com/solo-io/solo-kit"
		outputDir             = "_output"
	)

	var (
		tempDir               string
		relativePathToTempDir string
		modRootDir            string
	)

	BeforeEach(func() {
		modPackageFile, err := modutils.GetCurrentModPackageFile()
		Expect(err).NotTo(HaveOccurred())
		modRootDir = filepath.Dir(modPackageFile)
		// Create temp directory and path variables
		workingDir, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		outputDir := filepath.Join(workingDir, outputDir)
		err = os.MkdirAll(outputDir, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
		tempDir, err = ioutil.TempDir(outputDir, "doc-gen-test-")
		Expect(err).NotTo(HaveOccurred())
		relativePathToTempDir = filepath.Join(packageName, outputDir, filepath.Base(tempDir))

		// Generate test proto file for which doc has to be generated
		buf := &bytes.Buffer{}
		err = testProtoTemplate().Execute(buf, relativePathToTempDir)
		Expect(err).NotTo(HaveOccurred())
		err = ioutil.WriteFile(filepath.Join(tempDir, testProtoName), []byte(buf.String()), 0644)
		Expect(err).NotTo(HaveOccurred())

		// Generate test proto file for which doc has not to be generated
		buf = &bytes.Buffer{}
		err = testProtoNoDocsTemplate().Execute(buf, relativePathToTempDir)
		Expect(err).NotTo(HaveOccurred())
		err = ioutil.WriteFile(filepath.Join(tempDir, testProtoNoDocsName), []byte(buf.String()), 0644)
		Expect(err).NotTo(HaveOccurred())

		// Generate project config
		buf = &bytes.Buffer{}
		err = projectConfigFile().Execute(buf, struct {
			Dir       string
			GoPackage string
		}{
			Dir:       projectConfigDocsDir,
			GoPackage: relativePathToTempDir,
		})
		Expect(err).NotTo(HaveOccurred())
		err = ioutil.WriteFile(filepath.Join(tempDir, testProjectConfigName), []byte(buf.String()), 0644)
		Expect(err).NotTo(HaveOccurred())

		genDocs := &cmd.DocsOptions{
			Output: options.Hugo,
			HugoOptions: &options.HugoOptions{
				DataDir: hugoDataDir,
				ApiDir:  hugoApiDir,
			},
		}

		// Run code gen
		opts := cmd.GenerateOptions{
			RelativeRoot:  filepath.Join(".", outputDir, filepath.Base(tempDir)),
			SkipGenMocks:  true,
			CompileProtos: true,
			GenDocs:       genDocs,
			ExternalImports: &sk_anyvendor.Imports{
				Local: []string{
					"test/**/*.proto",
					"api/**/*.proto",
					filepath.Join(strings.TrimPrefix(tempDir, modRootDir), anyvendor.ProtoMatchPattern),
					sk_anyvendor.SoloKitMatchPattern},
				External: map[string][]string{
					sk_anyvendor.ExtProtoMatcher.Package:           sk_anyvendor.ExtProtoMatcher.Patterns,
					sk_anyvendor.EnvoyValidateProtoMatcher.Package: sk_anyvendor.EnvoyValidateProtoMatcher.Patterns,
				},
			},
		}
		err = cmd.Generate(opts)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tempDir)).NotTo(HaveOccurred())
	})

	It("docs for a message are generated based on the value of the skip_docs_gen option", func() {

		// Traverse the generated doc directory tree
		foundExpectedDoc, foundUnexpectedDoc := false, false
		outPath, err := filepath.Abs(filepath.Base(tempDir))
		Expect(err).NotTo(HaveOccurred())
		err = filepath.Walk(filepath.Join(outPath, "docs"), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {

				// Verify that a doc file has been generated for GenerateDocsForMe
				if info.Name() == testProtoName+".sk.md" {
					foundExpectedDoc = true
				}

				// Verify that no doc file has been generated for DoNotGenerateDocsForMe
				if info.Name() == testProtoNoDocsName+".sk.md" {
					foundUnexpectedDoc = true
				}

				// No file must contain any reference to DoNotGenerateDocsForMe
				file, err := ioutil.ReadFile(path)
				Expect(err).NotTo(HaveOccurred())
				matched, err := regexp.Match("(?i)DoNotGenerateDocsForMe", file)
				Expect(err).NotTo(HaveOccurred())
				Expect(matched).To(BeFalse())
			}
			return nil
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(foundExpectedDoc).To(BeTrue())
		Expect(foundUnexpectedDoc).To(BeFalse())
		dataFile, err := ioutil.ReadFile(filepath.Join(outPath, hugoDataDir, options.HugoProtoDataFile))
		hugoProtoMap := &datafile.HugoProtobufData{}

		Expect(yaml.Unmarshal(dataFile, hugoProtoMap)).NotTo(HaveOccurred())
		apiSummary, ok := hugoProtoMap.Apis["testing.solo.io.GenerateDocsForMe"]
		Expect(ok).To(BeTrue())
		Expect(apiSummary).To(Equal(datafile.ApiSummary{
			RelativePath: filepath.Join(
				hugoApiDir,
				filepath.Join(packageName, strings.TrimPrefix(tempDir, modRootDir)),
				"doc_gen_test.proto.sk/#GenerateDocsForMe"),
			Package: "testing.solo.io",
		}))
		By("verify that data file's mapping matches Hugo's served url")
		servedFile := strings.Split(apiSummary.RelativePath, "/#")[0]
		diskFile := filepath.Join(outPath, projectConfigDocsDir, servedFile+".md")
		_, err = ioutil.ReadFile(diskFile)
		Expect(err).NotTo(HaveOccurred())

	})

})

func testProtoTemplate() *template.Template {
	return template.Must(template.New("testProtoTemplate").Parse(`

syntax = "proto3";

package testing.solo.io;
option go_package = "{{.}}";



import "github.com/solo-io/solo-kit/api/v1/metadata.proto";
import "github.com/solo-io/solo-kit/api/v1/status.proto";
import "github.com/solo-io/solo-kit/api/v1/solo-kit.proto";

message GenerateDocsForMe {
    option (core.solo.io.resource).short_name = "docs";
    option (core.solo.io.resource).plural_name = "generatedocsforme";
    core.solo.io.Metadata metadata = 1;
    core.solo.io.Status status = 6;

    // Some field
    string basic_field = 2;

}

`))
}

func testProtoNoDocsTemplate() *template.Template {
	return template.Must(template.New("testProtoTemplate").Parse(`

syntax = "proto3";

package testing.solo.io;
option go_package = "{{.}}";



import "github.com/solo-io/solo-kit/api/v1/metadata.proto";
import "github.com/solo-io/solo-kit/api/v1/status.proto";
import "github.com/solo-io/solo-kit/api/v1/solo-kit.proto";

message DoNotGenerateDocsForMe {
    option (core.solo.io.resource).short_name = "nodocs";
    option (core.solo.io.resource).plural_name = "donotgeneratedocsforme";
    option (core.solo.io.resource).skip_docs_gen = true;
    core.solo.io.Metadata metadata = 1;
    core.solo.io.Status status = 6;

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
  "docs_dir": "{{.Dir}}/api",
  "go_package": "{{.GoPackage}}"
}

`))
}
