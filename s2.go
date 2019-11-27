// Copyright 2018 tree xie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compress

import (
	"io"

	"github.com/klauspost/compress/s2"
	"github.com/vicanso/elton"
)

const (
	s2Encoding = "s2"
)

type (
	// S2Compressor s2 compressor
	S2Compressor struct{}
)

// Accept check accept encoding
func (*S2Compressor) Accept(c *elton.Context) (acceptable bool, encoding string) {
	return AcceptEncoding(c, s2Encoding)
}

func s2IsBetterCompress(level int) bool {
	if level == 0 || level > 3 {
		return true
	}
	return false
}

// Compress s2 compress
func (*S2Compressor) Compress(buf []byte, level int) ([]byte, error) {
	var dst []byte
	fn := s2.Encode
	if s2IsBetterCompress(level) {
		fn = s2.EncodeBetter
	}
	data := fn(dst, buf)
	return data, nil
}

// Pipe s2 pipe
func (*S2Compressor) Pipe(c *elton.Context, level int) (err error) {
	r := c.Body.(io.Reader)
	closer, ok := c.Body.(io.Closer)
	if ok {
		defer closer.Close()
	}
	var w *s2.Writer
	if s2IsBetterCompress(level) {
		w = s2.NewWriter(c.Response, s2.WriterBetterCompression())
	} else {
		w = s2.NewWriter(c.Response)
	}
	defer w.Close()
	_, err = io.Copy(w, r)
	return
}
