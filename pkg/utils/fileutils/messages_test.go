package fileutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/solo-io/solo-kit/test/mocks/v1"

	"io/ioutil"
	"os"

	. "github.com/solo-io/solo-kit/pkg/utils/fileutils"
)

var _ = Describe("Messages", func() {
	var filename string
	BeforeEach(func() {
		f, err := ioutil.TempFile("", "messages_test")
		Expect(err).NotTo(HaveOccurred())
		filename = f.Name()
	})
	AfterEach(func() {
		os.RemoveAll(filename)
	})
	It("Writes and reads proto messages into files", func() {
		input := &v1.MockResource{
			Data: "hi",
		}
		err := WriteToFile(filename, input)
		Expect(err).NotTo(HaveOccurred())

		b, err := ioutil.ReadFile(filename)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(b)).To(Equal(`data.json: hi
`))

		var output v1.MockResource
		err = ReadFileInto(filename, &output)
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(Equal(*input))
	})
})
