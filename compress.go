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
	"regexp"
	"strings"

	"github.com/vicanso/cod"
)

var (
	defaultCompressRegexp = regexp.MustCompile("text|javascript|json")
)

const (
	defaultCompresMinLength = 1024
)

type (
	// Compressor compressor interface
	Compressor interface {
		// Accept accept check function
		Accept(c *cod.Context) (acceptable bool, encoding string)
		// Compress compress function
		Compress([]byte, int) ([]byte, error)
	}
	// Config compress config
	Config struct {
		// Level compress level
		Level int
		// MinLength min compress length
		MinLength int
		// Checker check the data is compressable
		Checker *regexp.Regexp
		// CompressorList compressor list
		CompressorList []Compressor
		// Skipper skipper function
		Skipper cod.Skipper
	}
)

// AcceptEncoding check request accept encoding
func AcceptEncoding(c *cod.Context, encoding string) (bool, string) {
	acceptEncoding := c.GetRequestHeader(cod.HeaderAcceptEncoding)
	if strings.Contains(acceptEncoding, encoding) {
		return true, encoding
	}
	return false, ""
}

// NewDefault create a default compress middleware, support gzip
func NewDefault() cod.Handler {
	return NewWithDefaultCompressor(Config{})
}

// NewWithDefaultCompressor create compress middleware with default compressor
func NewWithDefaultCompressor(config Config) cod.Handler {
	compressorList := make([]Compressor, 0)

	// 添加默认的 brotli 压缩
	br := new(BrCompressor)
	_, err := br.Compress([]byte("brotli"), 0)
	if err != nil {
		compressorList = append(compressorList, br)
	}

	// 添加默认的 gzip 压缩
	compressorList = append(compressorList, new(GzipCompressor))
	config.CompressorList = compressorList

	return New(config)
}

// New create a new compress middleware
func New(config Config) cod.Handler {
	minLength := config.MinLength
	if minLength == 0 {
		minLength = defaultCompresMinLength
	}
	skipper := config.Skipper
	if skipper == nil {
		skipper = cod.DefaultSkipper
	}
	checker := config.Checker
	if checker == nil {
		checker = defaultCompressRegexp
	}
	compressorList := config.CompressorList
	return func(c *cod.Context) (err error) {
		if skipper(c) || compressorList == nil {
			return c.Next()
		}
		err = c.Next()
		if err != nil {
			return
		}

		bodyBuf := c.BodyBuffer
		// 如果数据为空，直接跳过
		if bodyBuf == nil {
			return
		}

		// encoding 不为空，已做处理，无需要压缩
		if c.GetHeader(cod.HeaderContentEncoding) != "" {
			return
		}
		contentType := c.GetHeader(cod.HeaderContentType)
		buf := bodyBuf.Bytes()
		// 如果数据长度少于最小压缩长度或数据类型为非可压缩，则返回
		if len(buf) < minLength || !checker.MatchString(contentType) {
			return
		}

		for _, compressor := range compressorList {
			acceptable, encoding := compressor.Accept(c)
			if !acceptable {
				continue
			}
			newBuf, e := compressor.Compress(buf, config.Level)
			// 如果压缩成功，则使用压缩数据
			// 失败则忽略
			if e == nil {
				c.SetHeader(cod.HeaderContentEncoding, encoding)
				bodyBuf.Reset()
				bodyBuf.Write(newBuf)
				break
			}
		}
		return
	}
}
