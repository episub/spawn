package opa

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/loader"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/storage/inmem"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/radovskyb/watcher"
)

var unsafeCompiler = ast.NewCompiler()
var unsafeDocuments = map[string]interface{}{}
var unsafeStore storage.Store
var unsafeQueries map[string]ast.Body
var regoStore storage.Store
var loaded bool
var mutex = &sync.RWMutex{}
var dMutex = &sync.RWMutex{}
var queryMutex = &sync.RWMutex{}

// GetCompiler Returns compiler object in thread-safe manner since we sometimes update the compiler in a separate thread
func GetCompiler(ctx context.Context) *ast.Compiler {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetCompiler")
	defer span.Finish()

	var copy *ast.Compiler

	mutex.RLock()
	copy = unsafeCompiler
	mutex.RUnlock()
	return copy
}

// GetStore Returns a new in memory storage with documents, using a mutext to ensure safety
func GetStore(ctx context.Context) storage.Store {
	span, ctx := opentracing.StartSpanFromContext(ctx, "GetStore")
	defer span.Finish()

	var s storage.Store

	dMutex.RLock()
	s = unsafeStore
	dMutex.RUnlock()
	return s
}

// LoadBundle Loads bundle from specified path, and checks for any changes to the folder to load again in future
func LoadBundle(path string) error {
	// This function could be modified to allow re-running, but not worth it here for a file system monitoring setup.  When loading bundle from a remote server, this function will be rewritten
	if loaded {
		return fmt.Errorf("Bundle already loaded.  Cannot load twice.")
	}

	err := loadCompiler(path)

	if err != nil {
		return err
	}

	loaded = true
	go func() {
		// All good, so let's set this up to reload automatically if files change:
		w := watcher.New()

		// SetMaxEvents to 1 to allow at most 1 event's to be received
		// on the Event channel per watching cycle.
		//
		// If SetMaxEvents is not set, the default is to send all events.
		w.SetMaxEvents(1)

		w.FilterOps(watcher.Rename, watcher.Move, watcher.Write, watcher.Create, watcher.Remove, watcher.Chmod)

		go func() {
			log.Printf("Watching %s for changes", path)
			for {
				select {
				case event := <-w.Event:
					fmt.Println(event) // Print the event's info.

					log.Printf("Reloading compiler")
					err := loadCompiler(path)

					if err != nil {
						log.Printf("Error reloading compiler: %s", err)
					}
				case err := <-w.Error:
					log.Fatalln(err)
				case <-w.Closed:
					return
				}
			}
		}()

		// Watch test_folder recursively for changes.
		if err := w.AddRecursive(path); err != nil {
			log.Fatalln(err)
		}

		// Start the watching process - it'll check for changes every 100ms.
		if err := w.Start(time.Second * 1); err != nil {
			log.Fatalln(err)
		}
	}()

	return nil
}

func loadCompiler(path string) error {
	log.Printf("Loading path %s", path)
	newCompiler := ast.NewCompiler()

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(pwd)

	log.Printf("Current dir: %s", pwd)
	err = os.Chdir(path)
	if err != nil {
		return err
	}

	result, err := loader.Filtered([]string{"."}, nil)
	if err != nil {
		return fmt.Errorf("Error loading all path: %s", err)
	}

	// Create map from all values for compiling:
	modules := make(map[string]*ast.Module)
	for k, v := range result.Modules {
		log.Printf("* %s", k)
		modules[k] = v.Parsed
	}

	// Compile the loaded modules:
	newCompiler.Compile(modules)

	if newCompiler.Failed() {
		return newCompiler.Errors
	}

	setCompiler(newCompiler, result.Documents)

	return nil
}

func getCompiledQuery(query string) ast.Body {
	var compiled ast.Body

	// Check if already compiled:
	queryMutex.RLock()
	if compiled, ok := unsafeQueries[query]; ok {
		compiled = unsafeQueries[query]
		queryMutex.RUnlock()
		return compiled
	}
	queryMutex.RUnlock()

	// We must compile ourselves:
	queryMutex.Lock()
	compiled = ast.MustParseBody(query)
	unsafeQueries[query] = compiled
	queryMutex.Unlock()

	return compiled
}

func setCompiler(compiler *ast.Compiler, documents map[string]interface{}) {
	mutex.Lock()
	unsafeCompiler = compiler
	mutex.Unlock()
	dMutex.Lock()
	unsafeDocuments = documents
	unsafeStore = inmem.NewFromObject(unsafeDocuments)
	dMutex.Unlock()
	queryMutex.Lock()
	unsafeQueries = make(map[string]ast.Body)
	queryMutex.Unlock()
}
