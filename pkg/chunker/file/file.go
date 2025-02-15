/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package file

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/fsnotify.v1"

	"github.com/talos-systems/talos/pkg/chunker"
)

// Options is the functional options struct.
type Options struct {
	Size int
}

// Option is the functional option func.
type Option func(*Options)

// Size sets the chunk size of the Chunker.
func Size(s int) Option {
	return func(args *Options) {
		args.Size = s
	}
}

// File is a conecrete type that implements the chunker.Chunker interface.
type File struct {
	source  Source
	options *Options
}

// Source is an interface describing the source of a File.
type Source = *os.File

// NewChunker initializes a Chunker with default values.
func NewChunker(source Source, setters ...Option) chunker.Chunker {
	opts := &Options{
		Size: 1024,
	}

	for _, setter := range setters {
		setter(opts)
	}

	return &File{
		source,
		opts,
	}
}

// Read implements ChunkReader.
//
// nolint: gocyclo
func (c *File) Read(ctx context.Context) <-chan []byte {
	// Create a buffered channel of length 1.
	ch := make(chan []byte, 1)

	filename := c.source.Name()

	go func(ch chan []byte) {
		defer close(ch)

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Printf("failed to watch: %v\n", err)
			return
		}
		// nolint: errcheck
		defer watcher.Close()

		if err = watcher.Add(filepath.Dir(filename)); err != nil {
			log.Printf("failed to watch add: %v\n", err)
			return
		}
		offset, err := c.source.Seek(0, io.SeekStart)
		if err != nil {
			log.Printf("failed to seek: %v\n", err)
			return
		}

		buf := make([]byte, c.options.Size)

		for {
			for {
				n, err := c.source.ReadAt(buf, offset)
				if err != nil && err != io.EOF {
					log.Printf("read error: %s\n", err.Error())
					return
				}

				offset += int64(n)

				if n > 0 {
					// Copy the buffer since we will modify it in the next loop.
					b := make([]byte, n)
					copy(b, buf[:n])

				DELIVER:
					select {
					case <-ctx.Done():
						return
					case event := <-watcher.Events:
						// drain events while waiting for the buffer to be delivered
						// otherwise inotify() queue might overflow
						if event.Name == filename && event.Op == fsnotify.Write {
							// clear EOF condition (if there was one) to make sure
							// we read more data
							err = nil
						}
						goto DELIVER
					case ch <- b:
					}
				}

				if err == io.EOF || n == 0 {
					break
				}
			}

		WATCH:
			select {
			case <-ctx.Done():
				return
			case event := <-watcher.Events:
				if event.Name != filename {
					// ignore events for other files
					goto WATCH
				}
				switch event.Op {
				case fsnotify.Write:
					// new data, run one more loop copying data back to the client
				case fsnotify.Remove:
					log.Printf("file was removed while watching: %s", filename)
					return
				default:
					log.Printf("ignoring fsnotify event: %v\n", event)
					goto WATCH
				}
			case err := <-watcher.Errors:
				log.Printf("failed to watch: %v\n", err)
				return
			}
		}
	}(ch)

	return ch
}
