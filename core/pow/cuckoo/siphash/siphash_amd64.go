//+build amd64  !noasm !appengine

// Copyright (c) 2019 Aidos Developer

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package siphash

//go:noescape
func SiphashPRF8192(v *[4]uint64, nonce *[8192]uint64, uorv uint64, result *[8192]uint64)

// noescape 的作用是：禁止逃逸，而且它必须指示一个只有声明没有主体的函数，
// 告诉编译器，下面的函数无论如何都不会逃逸，那么当函数返回时，其中的资源也会一并都被销毁。
//go:noescape
func SiphashPRF8192Seq(v *[4]uint64, nonce uint64, uorv uint64, result *[8192]uint64)
