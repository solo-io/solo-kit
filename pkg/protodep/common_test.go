package protodep

import (
	"os"
	"path/filepath"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/errors"
	mock_protodep "github.com/solo-io/solo-kit/pkg/protodep/mocks"
	"github.com/spf13/afero"
)

//go:generate mockgen -package mock_protodep -destination ./mocks/afero.go github.com/spf13/afero Fs,File
//go:generate mockgen -package mock_protodep -destination ./mocks/fileinfo.go os FileInfo
//go:generate mockgen -package mock_protodep -destination ./mocks/copier.go -source ./common.go

var _ = Describe("common", func() {
	Context("copier", func() {
		Context("mocks", func() {
			var (
				ctrl         *gomock.Controller
				mockFs       *mock_protodep.MockFs
				mockFileInfo *mock_protodep.MockFileInfo
				mockFile     *mock_protodep.MockFile
				cp           *copier
			)
			BeforeEach(func() {
				ctrl = gomock.NewController(GinkgoT())
				mockFs = mock_protodep.NewMockFs(ctrl)
				mockFileInfo = mock_protodep.NewMockFileInfo(ctrl)
				mockFile = mock_protodep.NewMockFile(ctrl)
				cp = NewCopier(mockFs)
			})
			It("will return error if mkdir fails", func() {
				src, dst := "src/src.go", "dst/dstgo."
				fakeErr := errors.New("hello")
				mockFs.EXPECT().MkdirAll(filepath.Dir(dst), os.ModePerm).Return(fakeErr)
				_, err := cp.Copy(src, dst)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(fakeErr))
			})
			It("will return error if Stat fails", func() {
				src, dst := "src/src.go", "dst/dst.go"
				fakeErr := errors.New("hello")
				mockFs.EXPECT().MkdirAll(filepath.Dir(dst), os.ModePerm).Return(nil)
				mockFs.EXPECT().Stat(src).Return(nil, fakeErr)
				_, err := cp.Copy(src, dst)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(fakeErr))
			})
			It("will return error if fileinfo returns error fails", func() {
				src, dst := "src/src.go", "dst/dst.go"
				mockFs.EXPECT().MkdirAll(filepath.Dir(dst), os.ModePerm).Return(nil)
				mockFs.EXPECT().Stat(src).Return(mockFileInfo, nil)
				mockFileInfo.EXPECT().Mode().Return(os.ModeIrregular)
				_, err := cp.Copy(src, dst)
				Expect(err).To(HaveOccurred())
				isErr := eris.Is(err, IrregularFileError(src))
				Expect(isErr).To(BeTrue())
			})
			It("will return error if open fails", func() {
				src, dst := "src/src.go", "dst/dst.go"
				mockFs.EXPECT().MkdirAll(filepath.Dir(dst), os.ModePerm).Return(nil)
				mockFs.EXPECT().Stat(src).Return(mockFileInfo, nil)
				mockFileInfo.EXPECT().Mode().Return(os.ModePerm)
				fakeErr := errors.New("hello")
				mockFs.EXPECT().Open(src).Return(nil, fakeErr)
				_, err := cp.Copy(src, dst)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(fakeErr))
			})
			It("will return error if create fails", func() {
				src, dst := "src/src.go", "dst/dst.go"
				mockFs.EXPECT().MkdirAll(filepath.Dir(dst), os.ModePerm).Return(nil)
				mockFs.EXPECT().Stat(src).Return(mockFileInfo, nil)
				mockFileInfo.EXPECT().Mode().Return(os.ModePerm)
				fakeErr := errors.New("hello")
				mockFs.EXPECT().Open(src).Return(mockFile, nil)
				mockFile.EXPECT().Close().Return(nil)
				mockFs.EXPECT().Create(dst).Return(nil, fakeErr)
				_, err := cp.Copy(src, dst)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(fakeErr))
			})
		})
		Context("real copy", func() {
			It("Can copy", func() {
				fs := afero.NewOsFs()
				cp := &copier{fs: fs}
				tmpFile, err := afero.TempFile(fs, "", "")
				Expect(err).NotTo(HaveOccurred())
				defer fs.Remove(tmpFile.Name())
				tmpDir, err := afero.TempDir(fs, "", "")
				Expect(err).NotTo(HaveOccurred())
				defer fs.Remove(tmpDir)
				dstFileName := "test"
				_, err = cp.Copy(tmpFile.Name(), filepath.Join(tmpDir, dstFileName))
				Expect(err).NotTo(HaveOccurred())
				_, err = fs.Stat(filepath.Join(tmpDir, dstFileName))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
