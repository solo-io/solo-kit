package file_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// TODO: fix tests
func TestFile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "File Suite")
}

// var (
// 	tmpDir string
// )
//
// var _ = SynchronizedBeforeSuite(func() []byte {
// 	dbRunner = db.NewRunner()
// 	err := dbRunner.Start()
// 	Ω(err).ShouldNot(HaveOccurred())
// 	return []byte(dbRunner.URL)
// }, func(data []byte) {})
//
// var _ = SynchronizedAfterSuite(func() {
// 	dbClient.Cleanup()
// }, func() {
// 	dbRunner.Stop()
// })
