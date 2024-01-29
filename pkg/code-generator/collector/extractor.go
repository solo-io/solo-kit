package collector

import (
	"sync"
	"time"

	"github.com/rotisserie/eris"
)

type ImportsFetcher func(file string) ([]string, error)

type ImportsExtractor interface {
	FetchImportsForFile(protoFile string, importsFetcher ImportsFetcher) ([]string, error)
}

// thread-safe
// The synchronizedImportsExtractor provides synchronized access to imports for a given proto file.
// It provides 2 useful features for tree traversal:
//  1. Imports for each file are cached, ensuring that if we attempt to access that file
//     during traversal again, we do not need to duplicate work.
//  2. If imports for a file are unknown, and simultaneous go routines attempt to load
//     the imports, only 1 will execute and the other will block, waiting for the result.
//     This reduces the number of times we open and parse files.
type synchronizedImportsExtractor struct {
	// cachedImports contains a map of fileImports, each indexed by their file name
	cachedImports sync.Map

	// protect access to activeRequests
	activeRequestsMu sync.RWMutex
	activeRequests   map[string]*importsFetchRequest
}

type fileImports struct {
	imports []string
	err     error
}

func NewSynchronizedImportsExtractor() ImportsExtractor {
	return &synchronizedImportsExtractor{
		activeRequests: map[string]*importsFetchRequest{},
	}
}

var (
	FetchImportsTimeout = func(filename string) error {
		return eris.Errorf("Timed out while fetching imports for proto file: [%s]", filename)
	}
)

func (i *synchronizedImportsExtractor) FetchImportsForFile(protoFile string, importsFetcher ImportsFetcher) ([]string, error) {
	// Attempt to load the imports from the cache
	cachedImports, ok := i.cachedImports.Load(protoFile)
	if ok {
		typedCachedImports := cachedImports.(*fileImports)
		return typedCachedImports.imports, typedCachedImports.err
	}

	i.activeRequestsMu.Lock()

	// If there's not a current active request for this file, create one.
	activeRequest := i.activeRequests[protoFile]
	if activeRequest == nil {
		activeRequest = newImportsFetchRequest()
		i.activeRequests[protoFile] = activeRequest

		// This goroutine has exclusive ownership over the current request.
		// It releases the resource by nil'ing the importRequest field
		// once the goroutine is done.
		go func() {
			// fetch the imports
			imports, err := importsFetcher(protoFile)

			// update the cache
			i.cachedImports.Store(protoFile, &fileImports{imports: imports, err: err})

			// Signal to waiting goroutines
			activeRequest.done(imports, err)

			// Free active request so a different request can run.
			i.activeRequestsMu.Lock()
			defer i.activeRequestsMu.Unlock()
			delete(i.activeRequests, protoFile)
		}()
	}

	inflightRequest := i.activeRequests[protoFile]
	i.activeRequestsMu.Unlock()

	select {
	case <-time.After(30 * time.Second):
		// We should never reach this. This can only occur if we deadlock on file imports
		// which only happens with cyclic dependencies
		// Perhaps a safer alternative to erroring is just to execute the importsFetcher:
		// 	return importsFetcher(protoFile)
		return nil, FetchImportsTimeout(protoFile)
	case <-inflightRequest.wait():
		// Wait for the existing request to complete, then return
		return inflightRequest.result()
	}
}

// importsFetchRequest is used to wait on some in-flight request from multiple goroutines.
// Derived from: https://github.com/coreos/go-oidc/blob/08563f61dbb316f8ef85b784d01da503f2480690/oidc/jwks.go#L53
type importsFetchRequest struct {
	doneCh  chan struct{}
	imports []string
	err     error
}

func newImportsFetchRequest() *importsFetchRequest {
	return &importsFetchRequest{doneCh: make(chan struct{})}
}

// wait returns a channel that multiple goroutines can receive on. Once it returns
// a value, the inflight request is done and result() can be inspected.
func (i *importsFetchRequest) wait() <-chan struct{} {
	return i.doneCh
}

// done can only be called by a single goroutine. It records the result of the
// inflight request and signals other goroutines that the result is safe to
// inspect.
func (i *importsFetchRequest) done(imports []string, err error) {
	i.imports = imports
	i.err = err
	close(i.doneCh)
}

// result cannot be called until the wait() channel has returned a value.
func (i *importsFetchRequest) result() ([]string, error) {
	return i.imports, i.err
}

// This is the old implementation, that does not include caching or locking
type primitiveImportsExtractor struct {
}

func (p *primitiveImportsExtractor) FetchImportsForFile(protoFile string, importsFetcher ImportsFetcher) ([]string, error) {
	return importsFetcher(protoFile)
}
