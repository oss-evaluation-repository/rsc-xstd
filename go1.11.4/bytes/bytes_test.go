// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes_test

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	. "rsc.io/xstd/go1.11.4/bytes"
	"rsc.io/xstd/go1.11.4/internal/testenv"
	"rsc.io/xstd/go1.11.4/strings"
	"rsc.io/xstd/go1.11.4/unicode"
	"rsc.io/xstd/go1.11.4/unicode/utf8"
)

func eq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sliceOfString(s [][]byte) []string {
	result := make([]string, len(s))
	for i, v := range s {
		result[i] = string(v)
	}
	return result
}

// For ease of reading, the test cases use strings that are converted to byte
// slices before invoking the functions.

var abcd = "abcd"
var faces = "☺☻☹"
var commas = "1,2,3,4"
var dots = "1....2....3....4"

type BinOpTest struct {
	a string
	b string
	i int
}

func TestEqual(t *testing.T) {
	for _, tt := range compareTests {
		eql := Equal(tt.a, tt.b)
		if eql != (tt.i == 0) {
			t.Errorf(`Equal(%q, %q) = %v`, tt.a, tt.b, eql)
		}
		eql = EqualPortable(tt.a, tt.b)
		if eql != (tt.i == 0) {
			t.Errorf(`EqualPortable(%q, %q) = %v`, tt.a, tt.b, eql)
		}
	}
}

func TestEqualExhaustive(t *testing.T) {
	var size = 128
	if testing.Short() {
		size = 32
	}
	a := make([]byte, size)
	b := make([]byte, size)
	b_init := make([]byte, size)
	// randomish but deterministic data
	for i := 0; i < size; i++ {
		a[i] = byte(17 * i)
		b_init[i] = byte(23*i + 100)
	}

	for len := 0; len <= size; len++ {
		for x := 0; x <= size-len; x++ {
			for y := 0; y <= size-len; y++ {
				copy(b, b_init)
				copy(b[y:y+len], a[x:x+len])
				if !Equal(a[x:x+len], b[y:y+len]) || !Equal(b[y:y+len], a[x:x+len]) {
					t.Errorf("Equal(%d, %d, %d) = false", len, x, y)
				}
			}
		}
	}
}

// make sure Equal returns false for minimally different strings. The data
// is all zeros except for a single one in one location.
func TestNotEqual(t *testing.T) {
	var size = 128
	if testing.Short() {
		size = 32
	}
	a := make([]byte, size)
	b := make([]byte, size)

	for len := 0; len <= size; len++ {
		for x := 0; x <= size-len; x++ {
			for y := 0; y <= size-len; y++ {
				for diffpos := x; diffpos < x+len; diffpos++ {
					a[diffpos] = 1
					if Equal(a[x:x+len], b[y:y+len]) || Equal(b[y:y+len], a[x:x+len]) {
						t.Errorf("NotEqual(%d, %d, %d, %d) = true", len, x, y, diffpos)
					}
					a[diffpos] = 0
				}
			}
		}
	}
}

var indexTests = []BinOpTest{
	{"", "", 0},
	{"", "a", -1},
	{"", "foo", -1},
	{"fo", "foo", -1},
	{"foo", "baz", -1},
	{"foo", "foo", 0},
	{"oofofoofooo", "f", 2},
	{"oofofoofooo", "foo", 4},
	{"barfoobarfoo", "foo", 3},
	{"foo", "", 0},
	{"foo", "o", 1},
	{"abcABCabc", "A", 3},
	// cases with one byte strings - test IndexByte and special case in Index()
	{"", "a", -1},
	{"x", "a", -1},
	{"x", "x", 0},
	{"abc", "a", 0},
	{"abc", "b", 1},
	{"abc", "c", 2},
	{"abc", "x", -1},
	{"barfoobarfooyyyzzzyyyzzzyyyzzzyyyxxxzzzyyy", "x", 33},
	{"foofyfoobarfoobar", "y", 4},
	{"oooooooooooooooooooooo", "r", -1},
	// test fallback to Rabin-Karp.
	{"oxoxoxoxoxoxoxoxoxoxoxoy", "oy", 22},
	{"oxoxoxoxoxoxoxoxoxoxoxox", "oy", -1},
}

var lastIndexTests = []BinOpTest{
	{"", "", 0},
	{"", "a", -1},
	{"", "foo", -1},
	{"fo", "foo", -1},
	{"foo", "foo", 0},
	{"foo", "f", 0},
	{"oofofoofooo", "f", 7},
	{"oofofoofooo", "foo", 7},
	{"barfoobarfoo", "foo", 9},
	{"foo", "", 3},
	{"foo", "o", 2},
	{"abcABCabc", "A", 3},
	{"abcABCabc", "a", 6},
}

var indexAnyTests = []BinOpTest{
	{"", "", -1},
	{"", "a", -1},
	{"", "abc", -1},
	{"a", "", -1},
	{"a", "a", 0},
	{"aaa", "a", 0},
	{"abc", "xyz", -1},
	{"abc", "xcz", 2},
	{"ab☺c", "x☺yz", 2},
	{"a☺b☻c☹d", "cx", len("a☺b☻")},
	{"a☺b☻c☹d", "uvw☻xyz", len("a☺b")},
	{"aRegExp*", ".(|)*+?^$[]", 7},
	{dots + dots + dots, " ", -1},
	{"012abcba210", "\xffb", 4},
	{"012\x80bcb\x80210", "\xffb", 3},
}

var lastIndexAnyTests = []BinOpTest{
	{"", "", -1},
	{"", "a", -1},
	{"", "abc", -1},
	{"a", "", -1},
	{"a", "a", 0},
	{"aaa", "a", 2},
	{"abc", "xyz", -1},
	{"abc", "ab", 1},
	{"ab☺c", "x☺yz", 2},
	{"a☺b☻c☹d", "cx", len("a☺b☻")},
	{"a☺b☻c☹d", "uvw☻xyz", len("a☺b")},
	{"a.RegExp*", ".(|)*+?^$[]", 8},
	{dots + dots + dots, " ", -1},
	{"012abcba210", "\xffb", 6},
	{"012\x80bcb\x80210", "\xffb", 7},
}

// Execute f on each test case.  funcName should be the name of f; it's used
// in failure reports.
func runIndexTests(t *testing.T, f func(s, sep []byte) int, funcName string, testCases []BinOpTest) {
	for _, test := range testCases {
		a := []byte(test.a)
		b := []byte(test.b)
		actual := f(a, b)
		if actual != test.i {
			t.Errorf("%s(%q,%q) = %v; want %v", funcName, a, b, actual, test.i)
		}
	}
}

func runIndexAnyTests(t *testing.T, f func(s []byte, chars string) int, funcName string, testCases []BinOpTest) {
	for _, test := range testCases {
		a := []byte(test.a)
		actual := f(a, test.b)
		if actual != test.i {
			t.Errorf("%s(%q,%q) = %v; want %v", funcName, a, test.b, actual, test.i)
		}
	}
}

func TestIndex(t *testing.T)     { runIndexTests(t, Index, "Index", indexTests) }
func TestLastIndex(t *testing.T) { runIndexTests(t, LastIndex, "LastIndex", lastIndexTests) }
func TestIndexAny(t *testing.T)  { runIndexAnyTests(t, IndexAny, "IndexAny", indexAnyTests) }
func TestLastIndexAny(t *testing.T) {
	runIndexAnyTests(t, LastIndexAny, "LastIndexAny", lastIndexAnyTests)
}

func TestIndexByte(t *testing.T) {
	for _, tt := range indexTests {
		if len(tt.b) != 1 {
			continue
		}
		a := []byte(tt.a)
		b := tt.b[0]
		pos := IndexByte(a, b)
		if pos != tt.i {
			t.Errorf(`IndexByte(%q, '%c') = %v`, tt.a, b, pos)
		}
		posp := IndexBytePortable(a, b)
		if posp != tt.i {
			t.Errorf(`indexBytePortable(%q, '%c') = %v`, tt.a, b, posp)
		}
	}
}

func TestLastIndexByte(t *testing.T) {
	testCases := []BinOpTest{
		{"", "q", -1},
		{"abcdef", "q", -1},
		{"abcdefabcdef", "a", len("abcdef")},      // something in the middle
		{"abcdefabcdef", "f", len("abcdefabcde")}, // last byte
		{"zabcdefabcdef", "z", 0},                 // first byte
		{"a☺b☻c☹d", "b", len("a☺")},               // non-ascii
	}
	for _, test := range testCases {
		actual := LastIndexByte([]byte(test.a), test.b[0])
		if actual != test.i {
			t.Errorf("LastIndexByte(%q,%c) = %v; want %v", test.a, test.b[0], actual, test.i)
		}
	}
}

// test a larger buffer with different sizes and alignments
func TestIndexByteBig(t *testing.T) {
	var n = 1024
	if testing.Short() {
		n = 128
	}
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		// different start alignments
		b1 := b[i:]
		for j := 0; j < len(b1); j++ {
			b1[j] = 'x'
			pos := IndexByte(b1, 'x')
			if pos != j {
				t.Errorf("IndexByte(%q, 'x') = %v", b1, pos)
			}
			b1[j] = 0
			pos = IndexByte(b1, 'x')
			if pos != -1 {
				t.Errorf("IndexByte(%q, 'x') = %v", b1, pos)
			}
		}
		// different end alignments
		b1 = b[:i]
		for j := 0; j < len(b1); j++ {
			b1[j] = 'x'
			pos := IndexByte(b1, 'x')
			if pos != j {
				t.Errorf("IndexByte(%q, 'x') = %v", b1, pos)
			}
			b1[j] = 0
			pos = IndexByte(b1, 'x')
			if pos != -1 {
				t.Errorf("IndexByte(%q, 'x') = %v", b1, pos)
			}
		}
		// different start and end alignments
		b1 = b[i/2 : n-(i+1)/2]
		for j := 0; j < len(b1); j++ {
			b1[j] = 'x'
			pos := IndexByte(b1, 'x')
			if pos != j {
				t.Errorf("IndexByte(%q, 'x') = %v", b1, pos)
			}
			b1[j] = 0
			pos = IndexByte(b1, 'x')
			if pos != -1 {
				t.Errorf("IndexByte(%q, 'x') = %v", b1, pos)
			}
		}
	}
}

// test a small index across all page offsets
func TestIndexByteSmall(t *testing.T) {
	b := make([]byte, 5015) // bigger than a page
	// Make sure we find the correct byte even when straddling a page.
	for i := 0; i <= len(b)-15; i++ {
		for j := 0; j < 15; j++ {
			b[i+j] = byte(100 + j)
		}
		for j := 0; j < 15; j++ {
			p := IndexByte(b[i:i+15], byte(100+j))
			if p != j {
				t.Errorf("IndexByte(%q, %d) = %d", b[i:i+15], 100+j, p)
			}
		}
		for j := 0; j < 15; j++ {
			b[i+j] = 0
		}
	}
	// Make sure matches outside the slice never trigger.
	for i := 0; i <= len(b)-15; i++ {
		for j := 0; j < 15; j++ {
			b[i+j] = 1
		}
		for j := 0; j < 15; j++ {
			p := IndexByte(b[i:i+15], byte(0))
			if p != -1 {
				t.Errorf("IndexByte(%q, %d) = %d", b[i:i+15], 0, p)
			}
		}
		for j := 0; j < 15; j++ {
			b[i+j] = 0
		}
	}
}

func TestIndexRune(t *testing.T) {
	tests := []struct {
		in   string
		rune rune
		want int
	}{
		{"", 'a', -1},
		{"", '☺', -1},
		{"foo", '☹', -1},
		{"foo", 'o', 1},
		{"foo☺bar", '☺', 3},
		{"foo☺☻☹bar", '☹', 9},
		{"a A x", 'A', 2},
		{"some_text=some_value", '=', 9},
		{"☺a", 'a', 3},
		{"a☻☺b", '☺', 4},

		// RuneError should match any invalid UTF-8 byte sequence.
		{"�", '�', 0},
		{"\xff", '�', 0},
		{"☻x�", '�', len("☻x")},
		{"☻x\xe2\x98", '�', len("☻x")},
		{"☻x\xe2\x98�", '�', len("☻x")},
		{"☻x\xe2\x98x", '�', len("☻x")},

		// Invalid rune values should never match.
		{"a☺b☻c☹d\xe2\x98�\xff�\xed\xa0\x80", -1, -1},
		{"a☺b☻c☹d\xe2\x98�\xff�\xed\xa0\x80", 0xD800, -1}, // Surrogate pair
		{"a☺b☻c☹d\xe2\x98�\xff�\xed\xa0\x80", utf8.MaxRune + 1, -1},
	}
	for _, tt := range tests {
		if got := IndexRune([]byte(tt.in), tt.rune); got != tt.want {
			t.Errorf("IndexRune(%q, %d) = %v; want %v", tt.in, tt.rune, got, tt.want)
		}
	}

	haystack := []byte("test世界")
	allocs := testing.AllocsPerRun(1000, func() {
		if i := IndexRune(haystack, 's'); i != 2 {
			t.Fatalf("'s' at %d; want 2", i)
		}
		if i := IndexRune(haystack, '世'); i != 4 {
			t.Fatalf("'世' at %d; want 4", i)
		}
	})
	if allocs != 0 {
		t.Errorf("expected no allocations, got %f", allocs)
	}
}

// test count of a single byte across page offsets
func TestCountByte(t *testing.T) {
	b := make([]byte, 5015) // bigger than a page
	windows := []int{1, 2, 3, 4, 15, 16, 17, 31, 32, 33, 63, 64, 65, 128}
	testCountWindow := func(i, window int) {
		for j := 0; j < window; j++ {
			b[i+j] = byte(100)
			p := Count(b[i:i+window], []byte{100})
			if p != j+1 {
				t.Errorf("TestCountByte.Count(%q, 100) = %d", b[i:i+window], p)
			}
		}
	}

	maxWnd := windows[len(windows)-1]

	for i := 0; i <= 2*maxWnd; i++ {
		for _, window := range windows {
			if window > len(b[i:]) {
				window = len(b[i:])
			}
			testCountWindow(i, window)
			for j := 0; j < window; j++ {
				b[i+j] = byte(0)
			}
		}
	}
	for i := 4096 - (maxWnd + 1); i < len(b); i++ {
		for _, window := range windows {
			if window > len(b[i:]) {
				window = len(b[i:])
			}
			testCountWindow(i, window)
			for j := 0; j < window; j++ {
				b[i+j] = byte(0)
			}
		}
	}
}

// Make sure we don't count bytes outside our window
func TestCountByteNoMatch(t *testing.T) {
	b := make([]byte, 5015)
	windows := []int{1, 2, 3, 4, 15, 16, 17, 31, 32, 33, 63, 64, 65, 128}
	for i := 0; i <= len(b); i++ {
		for _, window := range windows {
			if window > len(b[i:]) {
				window = len(b[i:])
			}
			// Fill the window with non-match
			for j := 0; j < window; j++ {
				b[i+j] = byte(100)
			}
			// Try to find something that doesn't exist
			p := Count(b[i:i+window], []byte{0})
			if p != 0 {
				t.Errorf("TestCountByteNoMatch(%q, 0) = %d", b[i:i+window], p)
			}
			for j := 0; j < window; j++ {
				b[i+j] = byte(0)
			}
		}
	}
}

var bmbuf []byte

func valName(x int) string {
	if s := x >> 20; s<<20 == x {
		return fmt.Sprintf("%dM", s)
	}
	if s := x >> 10; s<<10 == x {
		return fmt.Sprintf("%dK", s)
	}
	return fmt.Sprint(x)
}

func benchBytes(b *testing.B, sizes []int, f func(b *testing.B, n int)) {
	for _, n := range sizes {
		if isRaceBuilder && n > 4<<10 {
			continue
		}
		b.Run(valName(n), func(b *testing.B) {
			if len(bmbuf) < n {
				bmbuf = make([]byte, n)
			}
			b.SetBytes(int64(n))
			f(b, n)
		})
	}
}

var indexSizes = []int{10, 32, 4 << 10, 4 << 20, 64 << 20}

var isRaceBuilder = strings.HasSuffix(testenv.Builder(), "-race")

func BenchmarkIndexByte(b *testing.B) {
	benchBytes(b, indexSizes, bmIndexByte(IndexByte))
}

func BenchmarkIndexBytePortable(b *testing.B) {
	benchBytes(b, indexSizes, bmIndexByte(IndexBytePortable))
}

func bmIndexByte(index func([]byte, byte) int) func(b *testing.B, n int) {
	return func(b *testing.B, n int) {
		buf := bmbuf[0:n]
		buf[n-1] = 'x'
		for i := 0; i < b.N; i++ {
			j := index(buf, 'x')
			if j != n-1 {
				b.Fatal("bad index", j)
			}
		}
		buf[n-1] = '\x00'
	}
}

func BenchmarkIndexRune(b *testing.B) {
	benchBytes(b, indexSizes, bmIndexRune(IndexRune))
}

func BenchmarkIndexRuneASCII(b *testing.B) {
	benchBytes(b, indexSizes, bmIndexRuneASCII(IndexRune))
}

func bmIndexRuneASCII(index func([]byte, rune) int) func(b *testing.B, n int) {
	return func(b *testing.B, n int) {
		buf := bmbuf[0:n]
		buf[n-1] = 'x'
		for i := 0; i < b.N; i++ {
			j := index(buf, 'x')
			if j != n-1 {
				b.Fatal("bad index", j)
			}
		}
		buf[n-1] = '\x00'
	}
}

func bmIndexRune(index func([]byte, rune) int) func(b *testing.B, n int) {
	return func(b *testing.B, n int) {
		buf := bmbuf[0:n]
		utf8.EncodeRune(buf[n-3:], '世')
		for i := 0; i < b.N; i++ {
			j := index(buf, '世')
			if j != n-3 {
				b.Fatal("bad index", j)
			}
		}
		buf[n-3] = '\x00'
		buf[n-2] = '\x00'
		buf[n-1] = '\x00'
	}
}

func BenchmarkEqual(b *testing.B) {
	b.Run("0", func(b *testing.B) {
		var buf [4]byte
		buf1 := buf[0:0]
		buf2 := buf[1:1]
		for i := 0; i < b.N; i++ {
			eq := Equal(buf1, buf2)
			if !eq {
				b.Fatal("bad equal")
			}
		}
	})

	sizes := []int{1, 6, 9, 15, 16, 20, 32, 4 << 10, 4 << 20, 64 << 20}
	benchBytes(b, sizes, bmEqual(Equal))
}

func BenchmarkEqualPort(b *testing.B) {
	sizes := []int{1, 6, 32, 4 << 10, 4 << 20, 64 << 20}
	benchBytes(b, sizes, bmEqual(EqualPortable))
}

func bmEqual(equal func([]byte, []byte) bool) func(b *testing.B, n int) {
	return func(b *testing.B, n int) {
		if len(bmbuf) < 2*n {
			bmbuf = make([]byte, 2*n)
		}
		buf1 := bmbuf[0:n]
		buf2 := bmbuf[n : 2*n]
		buf1[n-1] = 'x'
		buf2[n-1] = 'x'
		for i := 0; i < b.N; i++ {
			eq := equal(buf1, buf2)
			if !eq {
				b.Fatal("bad equal")
			}
		}
		buf1[n-1] = '\x00'
		buf2[n-1] = '\x00'
	}
}

func BenchmarkIndex(b *testing.B) {
	benchBytes(b, indexSizes, func(b *testing.B, n int) {
		buf := bmbuf[0:n]
		buf[n-1] = 'x'
		for i := 0; i < b.N; i++ {
			j := Index(buf, buf[n-7:])
			if j != n-7 {
				b.Fatal("bad index", j)
			}
		}
		buf[n-1] = '\x00'
	})
}

func BenchmarkIndexEasy(b *testing.B) {
	benchBytes(b, indexSizes, func(b *testing.B, n int) {
		buf := bmbuf[0:n]
		buf[n-1] = 'x'
		buf[n-7] = 'x'
		for i := 0; i < b.N; i++ {
			j := Index(buf, buf[n-7:])
			if j != n-7 {
				b.Fatal("bad index", j)
			}
		}
		buf[n-1] = '\x00'
		buf[n-7] = '\x00'
	})
}

func BenchmarkCount(b *testing.B) {
	benchBytes(b, indexSizes, func(b *testing.B, n int) {
		buf := bmbuf[0:n]
		buf[n-1] = 'x'
		for i := 0; i < b.N; i++ {
			j := Count(buf, buf[n-7:])
			if j != 1 {
				b.Fatal("bad count", j)
			}
		}
		buf[n-1] = '\x00'
	})
}

func BenchmarkCountEasy(b *testing.B) {
	benchBytes(b, indexSizes, func(b *testing.B, n int) {
		buf := bmbuf[0:n]
		buf[n-1] = 'x'
		buf[n-7] = 'x'
		for i := 0; i < b.N; i++ {
			j := Count(buf, buf[n-7:])
			if j != 1 {
				b.Fatal("bad count", j)
			}
		}
		buf[n-1] = '\x00'
		buf[n-7] = '\x00'
	})
}

func BenchmarkCountSingle(b *testing.B) {
	benchBytes(b, indexSizes, func(b *testing.B, n int) {
		buf := bmbuf[0:n]
		step := 8
		for i := 0; i < len(buf); i += step {
			buf[i] = 1
		}
		expect := (len(buf) + (step - 1)) / step
		for i := 0; i < b.N; i++ {
			j := Count(buf, []byte{1})
			if j != expect {
				b.Fatal("bad count", j, expect)
			}
		}
		for i := 0; i < len(buf); i++ {
			buf[i] = 0
		}
	})
}

type ExplodeTest struct {
	s string
	n int
	a []string
}

var explodetests = []ExplodeTest{
	{"", -1, []string{}},
	{abcd, -1, []string{"a", "b", "c", "d"}},
	{faces, -1, []string{"☺", "☻", "☹"}},
	{abcd, 2, []string{"a", "bcd"}},
}

func TestExplode(t *testing.T) {
	for _, tt := range explodetests {
		a := SplitN([]byte(tt.s), nil, tt.n)
		result := sliceOfString(a)
		if !eq(result, tt.a) {
			t.Errorf(`Explode("%s", %d) = %v; want %v`, tt.s, tt.n, result, tt.a)
			continue
		}
		s := Join(a, []byte{})
		if string(s) != tt.s {
			t.Errorf(`Join(Explode("%s", %d), "") = "%s"`, tt.s, tt.n, s)
		}
	}
}

type SplitTest struct {
	s   string
	sep string
	n   int
	a   []string
}

var splittests = []SplitTest{
	{abcd, "a", 0, nil},
	{abcd, "a", -1, []string{"", "bcd"}},
	{abcd, "z", -1, []string{"abcd"}},
	{abcd, "", -1, []string{"a", "b", "c", "d"}},
	{commas, ",", -1, []string{"1", "2", "3", "4"}},
	{dots, "...", -1, []string{"1", ".2", ".3", ".4"}},
	{faces, "☹", -1, []string{"☺☻", ""}},
	{faces, "~", -1, []string{faces}},
	{faces, "", -1, []string{"☺", "☻", "☹"}},
	{"1 2 3 4", " ", 3, []string{"1", "2", "3 4"}},
	{"1 2", " ", 3, []string{"1", "2"}},
	{"123", "", 2, []string{"1", "23"}},
	{"123", "", 17, []string{"1", "2", "3"}},
}

func TestSplit(t *testing.T) {
	for _, tt := range splittests {
		a := SplitN([]byte(tt.s), []byte(tt.sep), tt.n)

		// Appending to the results should not change future results.
		var x []byte
		for _, v := range a {
			x = append(v, 'z')
		}

		result := sliceOfString(a)
		if !eq(result, tt.a) {
			t.Errorf(`Split(%q, %q, %d) = %v; want %v`, tt.s, tt.sep, tt.n, result, tt.a)
			continue
		}
		if tt.n == 0 {
			continue
		}

		if want := tt.a[len(tt.a)-1] + "z"; string(x) != want {
			t.Errorf("last appended result was %s; want %s", x, want)
		}

		s := Join(a, []byte(tt.sep))
		if string(s) != tt.s {
			t.Errorf(`Join(Split(%q, %q, %d), %q) = %q`, tt.s, tt.sep, tt.n, tt.sep, s)
		}
		if tt.n < 0 {
			b := Split([]byte(tt.s), []byte(tt.sep))
			if !reflect.DeepEqual(a, b) {
				t.Errorf("Split disagrees withSplitN(%q, %q, %d) = %v; want %v", tt.s, tt.sep, tt.n, b, a)
			}
		}
		if len(a) > 0 {
			in, out := a[0], s
			if cap(in) == cap(out) && &in[:1][0] == &out[:1][0] {
				t.Errorf("Join(%#v, %q) didn't copy", a, tt.sep)
			}
		}
	}
}

var splitaftertests = []SplitTest{
	{abcd, "a", -1, []string{"a", "bcd"}},
	{abcd, "z", -1, []string{"abcd"}},
	{abcd, "", -1, []string{"a", "b", "c", "d"}},
	{commas, ",", -1, []string{"1,", "2,", "3,", "4"}},
	{dots, "...", -1, []string{"1...", ".2...", ".3...", ".4"}},
	{faces, "☹", -1, []string{"☺☻☹", ""}},
	{faces, "~", -1, []string{faces}},
	{faces, "", -1, []string{"☺", "☻", "☹"}},
	{"1 2 3 4", " ", 3, []string{"1 ", "2 ", "3 4"}},
	{"1 2 3", " ", 3, []string{"1 ", "2 ", "3"}},
	{"1 2", " ", 3, []string{"1 ", "2"}},
	{"123", "", 2, []string{"1", "23"}},
	{"123", "", 17, []string{"1", "2", "3"}},
}

func TestSplitAfter(t *testing.T) {
	for _, tt := range splitaftertests {
		a := SplitAfterN([]byte(tt.s), []byte(tt.sep), tt.n)

		// Appending to the results should not change future results.
		var x []byte
		for _, v := range a {
			x = append(v, 'z')
		}

		result := sliceOfString(a)
		if !eq(result, tt.a) {
			t.Errorf(`Split(%q, %q, %d) = %v; want %v`, tt.s, tt.sep, tt.n, result, tt.a)
			continue
		}

		if want := tt.a[len(tt.a)-1] + "z"; string(x) != want {
			t.Errorf("last appended result was %s; want %s", x, want)
		}

		s := Join(a, nil)
		if string(s) != tt.s {
			t.Errorf(`Join(Split(%q, %q, %d), %q) = %q`, tt.s, tt.sep, tt.n, tt.sep, s)
		}
		if tt.n < 0 {
			b := SplitAfter([]byte(tt.s), []byte(tt.sep))
			if !reflect.DeepEqual(a, b) {
				t.Errorf("SplitAfter disagrees withSplitAfterN(%q, %q, %d) = %v; want %v", tt.s, tt.sep, tt.n, b, a)
			}
		}
	}
}

type FieldsTest struct {
	s string
	a []string
}

var fieldstests = []FieldsTest{
	{"", []string{}},
	{" ", []string{}},
	{" \t ", []string{}},
	{"  abc  ", []string{"abc"}},
	{"1 2 3 4", []string{"1", "2", "3", "4"}},
	{"1  2  3  4", []string{"1", "2", "3", "4"}},
	{"1\t\t2\t\t3\t4", []string{"1", "2", "3", "4"}},
	{"1\u20002\u20013\u20024", []string{"1", "2", "3", "4"}},
	{"\u2000\u2001\u2002", []string{}},
	{"\n™\t™\n", []string{"™", "™"}},
	{faces, []string{faces}},
}

func TestFields(t *testing.T) {
	for _, tt := range fieldstests {
		b := []byte(tt.s)
		a := Fields(b)

		// Appending to the results should not change future results.
		var x []byte
		for _, v := range a {
			x = append(v, 'z')
		}

		result := sliceOfString(a)
		if !eq(result, tt.a) {
			t.Errorf("Fields(%q) = %v; want %v", tt.s, a, tt.a)
			continue
		}

		if string(b) != tt.s {
			t.Errorf("slice changed to %s; want %s", string(b), tt.s)
		}
		if len(tt.a) > 0 {
			if want := tt.a[len(tt.a)-1] + "z"; string(x) != want {
				t.Errorf("last appended result was %s; want %s", x, want)
			}
		}
	}
}

func TestFieldsFunc(t *testing.T) {
	for _, tt := range fieldstests {
		a := FieldsFunc([]byte(tt.s), unicode.IsSpace)
		result := sliceOfString(a)
		if !eq(result, tt.a) {
			t.Errorf("FieldsFunc(%q, unicode.IsSpace) = %v; want %v", tt.s, a, tt.a)
			continue
		}
	}
	pred := func(c rune) bool { return c == 'X' }
	var fieldsFuncTests = []FieldsTest{
		{"", []string{}},
		{"XX", []string{}},
		{"XXhiXXX", []string{"hi"}},
		{"aXXbXXXcX", []string{"a", "b", "c"}},
	}
	for _, tt := range fieldsFuncTests {
		b := []byte(tt.s)
		a := FieldsFunc(b, pred)

		// Appending to the results should not change future results.
		var x []byte
		for _, v := range a {
			x = append(v, 'z')
		}

		result := sliceOfString(a)
		if !eq(result, tt.a) {
			t.Errorf("FieldsFunc(%q) = %v, want %v", tt.s, a, tt.a)
		}

		if string(b) != tt.s {
			t.Errorf("slice changed to %s; want %s", b, tt.s)
		}
		if len(tt.a) > 0 {
			if want := tt.a[len(tt.a)-1] + "z"; string(x) != want {
				t.Errorf("last appended result was %s; want %s", x, want)
			}
		}
	}
}

// Test case for any function which accepts and returns a byte slice.
// For ease of creation, we write the byte slices as strings.
type StringTest struct {
	in, out string
}

var upperTests = []StringTest{
	{"", ""},
	{"abc", "ABC"},
	{"AbC123", "ABC123"},
	{"azAZ09_", "AZAZ09_"},
	{"\u0250\u0250\u0250\u0250\u0250", "\u2C6F\u2C6F\u2C6F\u2C6F\u2C6F"}, // grows one byte per char
}

var lowerTests = []StringTest{
	{"", ""},
	{"abc", "abc"},
	{"AbC123", "abc123"},
	{"azAZ09_", "azaz09_"},
	{"\u2C6D\u2C6D\u2C6D\u2C6D\u2C6D", "\u0251\u0251\u0251\u0251\u0251"}, // shrinks one byte per char
}

const space = "\t\v\r\f\n\u0085\u00a0\u2000\u3000"

var trimSpaceTests = []StringTest{
	{"", ""},
	{"abc", "abc"},
	{space + "abc" + space, "abc"},
	{" ", ""},
	{" \t\r\n \t\t\r\r\n\n ", ""},
	{" \t\r\n x\t\t\r\r\n\n ", "x"},
	{" \u2000\t\r\n x\t\t\r\r\ny\n \u3000", "x\t\t\r\r\ny"},
	{"1 \t\r\n2", "1 \t\r\n2"},
	{" x\x80", "x\x80"},
	{" x\xc0", "x\xc0"},
	{"x \xc0\xc0 ", "x \xc0\xc0"},
	{"x \xc0", "x \xc0"},
	{"x \xc0 ", "x \xc0"},
	{"x \xc0\xc0 ", "x \xc0\xc0"},
	{"x ☺\xc0\xc0 ", "x ☺\xc0\xc0"},
	{"x ☺ ", "x ☺"},
}

// Execute f on each test case.  funcName should be the name of f; it's used
// in failure reports.
func runStringTests(t *testing.T, f func([]byte) []byte, funcName string, testCases []StringTest) {
	for _, tc := range testCases {
		actual := string(f([]byte(tc.in)))
		if actual != tc.out {
			t.Errorf("%s(%q) = %q; want %q", funcName, tc.in, actual, tc.out)
		}
	}
}

func tenRunes(r rune) string {
	runes := make([]rune, 10)
	for i := range runes {
		runes[i] = r
	}
	return string(runes)
}

// User-defined self-inverse mapping function
func rot13(r rune) rune {
	const step = 13
	if r >= 'a' && r <= 'z' {
		return ((r - 'a' + step) % 26) + 'a'
	}
	if r >= 'A' && r <= 'Z' {
		return ((r - 'A' + step) % 26) + 'A'
	}
	return r
}

func TestMap(t *testing.T) {
	// Run a couple of awful growth/shrinkage tests
	a := tenRunes('a')

	// 1.  Grow. This triggers two reallocations in Map.
	maxRune := func(r rune) rune { return unicode.MaxRune }
	m := Map(maxRune, []byte(a))
	expect := tenRunes(unicode.MaxRune)
	if string(m) != expect {
		t.Errorf("growing: expected %q got %q", expect, m)
	}

	// 2. Shrink
	minRune := func(r rune) rune { return 'a' }
	m = Map(minRune, []byte(tenRunes(unicode.MaxRune)))
	expect = a
	if string(m) != expect {
		t.Errorf("shrinking: expected %q got %q", expect, m)
	}

	// 3. Rot13
	m = Map(rot13, []byte("a to zed"))
	expect = "n gb mrq"
	if string(m) != expect {
		t.Errorf("rot13: expected %q got %q", expect, m)
	}

	// 4. Rot13^2
	m = Map(rot13, Map(rot13, []byte("a to zed")))
	expect = "a to zed"
	if string(m) != expect {
		t.Errorf("rot13: expected %q got %q", expect, m)
	}

	// 5. Drop
	dropNotLatin := func(r rune) rune {
		if unicode.Is(unicode.Latin, r) {
			return r
		}
		return -1
	}
	m = Map(dropNotLatin, []byte("Hello, 세계"))
	expect = "Hello"
	if string(m) != expect {
		t.Errorf("drop: expected %q got %q", expect, m)
	}

	// 6. Invalid rune
	invalidRune := func(r rune) rune {
		return utf8.MaxRune + 1
	}
	m = Map(invalidRune, []byte("x"))
	expect = "\uFFFD"
	if string(m) != expect {
		t.Errorf("invalidRune: expected %q got %q", expect, m)
	}
}

func TestToUpper(t *testing.T) { runStringTests(t, ToUpper, "ToUpper", upperTests) }

func TestToLower(t *testing.T) { runStringTests(t, ToLower, "ToLower", lowerTests) }

func TestTrimSpace(t *testing.T) { runStringTests(t, TrimSpace, "TrimSpace", trimSpaceTests) }

type RepeatTest struct {
	in, out string
	count   int
}

var RepeatTests = []RepeatTest{
	{"", "", 0},
	{"", "", 1},
	{"", "", 2},
	{"-", "", 0},
	{"-", "-", 1},
	{"-", "----------", 10},
	{"abc ", "abc abc abc ", 3},
}

func TestRepeat(t *testing.T) {
	for _, tt := range RepeatTests {
		tin := []byte(tt.in)
		tout := []byte(tt.out)
		a := Repeat(tin, tt.count)
		if !Equal(a, tout) {
			t.Errorf("Repeat(%q, %d) = %q; want %q", tin, tt.count, a, tout)
			continue
		}
	}
}

func repeat(b []byte, count int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("%s", v)
			}
		}
	}()

	Repeat(b, count)

	return
}

// See Issue golang.org/issue/16237
func TestRepeatCatchesOverflow(t *testing.T) {
	tests := [...]struct {
		s      string
		count  int
		errStr string
	}{
		0: {"--", -2147483647, "negative"},
		1: {"", int(^uint(0) >> 1), ""},
		2: {"-", 10, ""},
		3: {"gopher", 0, ""},
		4: {"-", -1, "negative"},
		5: {"--", -102, "negative"},
		6: {string(make([]byte, 255)), int((^uint(0))/255 + 1), "overflow"},
	}

	for i, tt := range tests {
		err := repeat([]byte(tt.s), tt.count)
		if tt.errStr == "" {
			if err != nil {
				t.Errorf("#%d panicked %v", i, err)
			}
			continue
		}

		if err == nil || !strings.Contains(err.Error(), tt.errStr) {
			t.Errorf("#%d expected %q got %q", i, tt.errStr, err)
		}
	}
}

func runesEqual(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i, r := range a {
		if r != b[i] {
			return false
		}
	}
	return true
}

type RunesTest struct {
	in    string
	out   []rune
	lossy bool
}

var RunesTests = []RunesTest{
	{"", []rune{}, false},
	{" ", []rune{32}, false},
	{"ABC", []rune{65, 66, 67}, false},
	{"abc", []rune{97, 98, 99}, false},
	{"\u65e5\u672c\u8a9e", []rune{26085, 26412, 35486}, false},
	{"ab\x80c", []rune{97, 98, 0xFFFD, 99}, true},
	{"ab\xc0c", []rune{97, 98, 0xFFFD, 99}, true},
}

func TestRunes(t *testing.T) {
	for _, tt := range RunesTests {
		tin := []byte(tt.in)
		a := Runes(tin)
		if !runesEqual(a, tt.out) {
			t.Errorf("Runes(%q) = %v; want %v", tin, a, tt.out)
			continue
		}
		if !tt.lossy {
			// can only test reassembly if we didn't lose information
			s := string(a)
			if s != tt.in {
				t.Errorf("string(Runes(%q)) = %x; want %x", tin, s, tin)
			}
		}
	}
}

type TrimTest struct {
	f            string
	in, arg, out string
}

var trimTests = []TrimTest{
	{"Trim", "abba", "a", "bb"},
	{"Trim", "abba", "ab", ""},
	{"TrimLeft", "abba", "ab", ""},
	{"TrimRight", "abba", "ab", ""},
	{"TrimLeft", "abba", "a", "bba"},
	{"TrimRight", "abba", "a", "abb"},
	{"Trim", "<tag>", "<>", "tag"},
	{"Trim", "* listitem", " *", "listitem"},
	{"Trim", `"quote"`, `"`, "quote"},
	{"Trim", "\u2C6F\u2C6F\u0250\u0250\u2C6F\u2C6F", "\u2C6F", "\u0250\u0250"},
	{"Trim", "\x80test\xff", "\xff", "test"},
	{"Trim", " Ġ ", " ", "Ġ"},
	{"Trim", " Ġİ0", "0 ", "Ġİ"},
	//empty string tests
	{"Trim", "abba", "", "abba"},
	{"Trim", "", "123", ""},
	{"Trim", "", "", ""},
	{"TrimLeft", "abba", "", "abba"},
	{"TrimLeft", "", "123", ""},
	{"TrimLeft", "", "", ""},
	{"TrimRight", "abba", "", "abba"},
	{"TrimRight", "", "123", ""},
	{"TrimRight", "", "", ""},
	{"TrimRight", "☺\xc0", "☺", "☺\xc0"},
	{"TrimPrefix", "aabb", "a", "abb"},
	{"TrimPrefix", "aabb", "b", "aabb"},
	{"TrimSuffix", "aabb", "a", "aabb"},
	{"TrimSuffix", "aabb", "b", "aab"},
}

func TestTrim(t *testing.T) {
	for _, tc := range trimTests {
		name := tc.f
		var f func([]byte, string) []byte
		var fb func([]byte, []byte) []byte
		switch name {
		case "Trim":
			f = Trim
		case "TrimLeft":
			f = TrimLeft
		case "TrimRight":
			f = TrimRight
		case "TrimPrefix":
			fb = TrimPrefix
		case "TrimSuffix":
			fb = TrimSuffix
		default:
			t.Errorf("Undefined trim function %s", name)
		}
		var actual string
		if f != nil {
			actual = string(f([]byte(tc.in), tc.arg))
		} else {
			actual = string(fb([]byte(tc.in), []byte(tc.arg)))
		}
		if actual != tc.out {
			t.Errorf("%s(%q, %q) = %q; want %q", name, tc.in, tc.arg, actual, tc.out)
		}
	}
}

type predicate struct {
	f    func(r rune) bool
	name string
}

var isSpace = predicate{unicode.IsSpace, "IsSpace"}
var isDigit = predicate{unicode.IsDigit, "IsDigit"}
var isUpper = predicate{unicode.IsUpper, "IsUpper"}
var isValidRune = predicate{
	func(r rune) bool {
		return r != utf8.RuneError
	},
	"IsValidRune",
}

type TrimFuncTest struct {
	f       predicate
	in, out string
}

func not(p predicate) predicate {
	return predicate{
		func(r rune) bool {
			return !p.f(r)
		},
		"not " + p.name,
	}
}

var trimFuncTests = []TrimFuncTest{
	{isSpace, space + " hello " + space, "hello"},
	{isDigit, "\u0e50\u0e5212hello34\u0e50\u0e51", "hello"},
	{isUpper, "\u2C6F\u2C6F\u2C6F\u2C6FABCDhelloEF\u2C6F\u2C6FGH\u2C6F\u2C6F", "hello"},
	{not(isSpace), "hello" + space + "hello", space},
	{not(isDigit), "hello\u0e50\u0e521234\u0e50\u0e51helo", "\u0e50\u0e521234\u0e50\u0e51"},
	{isValidRune, "ab\xc0a\xc0cd", "\xc0a\xc0"},
	{not(isValidRune), "\xc0a\xc0", "a"},
}

func TestTrimFunc(t *testing.T) {
	for _, tc := range trimFuncTests {
		actual := string(TrimFunc([]byte(tc.in), tc.f.f))
		if actual != tc.out {
			t.Errorf("TrimFunc(%q, %q) = %q; want %q", tc.in, tc.f.name, actual, tc.out)
		}
	}
}

type IndexFuncTest struct {
	in          string
	f           predicate
	first, last int
}

var indexFuncTests = []IndexFuncTest{
	{"", isValidRune, -1, -1},
	{"abc", isDigit, -1, -1},
	{"0123", isDigit, 0, 3},
	{"a1b", isDigit, 1, 1},
	{space, isSpace, 0, len(space) - 3}, // last rune in space is 3 bytes
	{"\u0e50\u0e5212hello34\u0e50\u0e51", isDigit, 0, 18},
	{"\u2C6F\u2C6F\u2C6F\u2C6FABCDhelloEF\u2C6F\u2C6FGH\u2C6F\u2C6F", isUpper, 0, 34},
	{"12\u0e50\u0e52hello34\u0e50\u0e51", not(isDigit), 8, 12},

	// tests of invalid UTF-8
	{"\x801", isDigit, 1, 1},
	{"\x80abc", isDigit, -1, -1},
	{"\xc0a\xc0", isValidRune, 1, 1},
	{"\xc0a\xc0", not(isValidRune), 0, 2},
	{"\xc0☺\xc0", not(isValidRune), 0, 4},
	{"\xc0☺\xc0\xc0", not(isValidRune), 0, 5},
	{"ab\xc0a\xc0cd", not(isValidRune), 2, 4},
	{"a\xe0\x80cd", not(isValidRune), 1, 2},
}

func TestIndexFunc(t *testing.T) {
	for _, tc := range indexFuncTests {
		first := IndexFunc([]byte(tc.in), tc.f.f)
		if first != tc.first {
			t.Errorf("IndexFunc(%q, %s) = %d; want %d", tc.in, tc.f.name, first, tc.first)
		}
		last := LastIndexFunc([]byte(tc.in), tc.f.f)
		if last != tc.last {
			t.Errorf("LastIndexFunc(%q, %s) = %d; want %d", tc.in, tc.f.name, last, tc.last)
		}
	}
}

type ReplaceTest struct {
	in       string
	old, new string
	n        int
	out      string
}

var ReplaceTests = []ReplaceTest{
	{"hello", "l", "L", 0, "hello"},
	{"hello", "l", "L", -1, "heLLo"},
	{"hello", "x", "X", -1, "hello"},
	{"", "x", "X", -1, ""},
	{"radar", "r", "<r>", -1, "<r>ada<r>"},
	{"", "", "<>", -1, "<>"},
	{"banana", "a", "<>", -1, "b<>n<>n<>"},
	{"banana", "a", "<>", 1, "b<>nana"},
	{"banana", "a", "<>", 1000, "b<>n<>n<>"},
	{"banana", "an", "<>", -1, "b<><>a"},
	{"banana", "ana", "<>", -1, "b<>na"},
	{"banana", "", "<>", -1, "<>b<>a<>n<>a<>n<>a<>"},
	{"banana", "", "<>", 10, "<>b<>a<>n<>a<>n<>a<>"},
	{"banana", "", "<>", 6, "<>b<>a<>n<>a<>n<>a"},
	{"banana", "", "<>", 5, "<>b<>a<>n<>a<>na"},
	{"banana", "", "<>", 1, "<>banana"},
	{"banana", "a", "a", -1, "banana"},
	{"banana", "a", "a", 1, "banana"},
	{"☺☻☹", "", "<>", -1, "<>☺<>☻<>☹<>"},
}

func TestReplace(t *testing.T) {
	for _, tt := range ReplaceTests {
		in := append([]byte(tt.in), "<spare>"...)
		in = in[:len(tt.in)]
		out := Replace(in, []byte(tt.old), []byte(tt.new), tt.n)
		if s := string(out); s != tt.out {
			t.Errorf("Replace(%q, %q, %q, %d) = %q, want %q", tt.in, tt.old, tt.new, tt.n, s, tt.out)
		}
		if cap(in) == cap(out) && &in[:1][0] == &out[:1][0] {
			t.Errorf("Replace(%q, %q, %q, %d) didn't copy", tt.in, tt.old, tt.new, tt.n)
		}
	}
}

type TitleTest struct {
	in, out string
}

var TitleTests = []TitleTest{
	{"", ""},
	{"a", "A"},
	{" aaa aaa aaa ", " Aaa Aaa Aaa "},
	{" Aaa Aaa Aaa ", " Aaa Aaa Aaa "},
	{"123a456", "123a456"},
	{"double-blind", "Double-Blind"},
	{"ÿøû", "Ÿøû"},
	{"with_underscore", "With_underscore"},
	{"unicode \xe2\x80\xa8 line separator", "Unicode \xe2\x80\xa8 Line Separator"},
}

func TestTitle(t *testing.T) {
	for _, tt := range TitleTests {
		if s := string(Title([]byte(tt.in))); s != tt.out {
			t.Errorf("Title(%q) = %q, want %q", tt.in, s, tt.out)
		}
	}
}

var ToTitleTests = []TitleTest{
	{"", ""},
	{"a", "A"},
	{" aaa aaa aaa ", " AAA AAA AAA "},
	{" Aaa Aaa Aaa ", " AAA AAA AAA "},
	{"123a456", "123A456"},
	{"double-blind", "DOUBLE-BLIND"},
	{"ÿøû", "ŸØÛ"},
}

func TestToTitle(t *testing.T) {
	for _, tt := range ToTitleTests {
		if s := string(ToTitle([]byte(tt.in))); s != tt.out {
			t.Errorf("ToTitle(%q) = %q, want %q", tt.in, s, tt.out)
		}
	}
}

var EqualFoldTests = []struct {
	s, t string
	out  bool
}{
	{"abc", "abc", true},
	{"ABcd", "ABcd", true},
	{"123abc", "123ABC", true},
	{"αβδ", "ΑΒΔ", true},
	{"abc", "xyz", false},
	{"abc", "XYZ", false},
	{"abcdefghijk", "abcdefghijX", false},
	{"abcdefghijk", "abcdefghij\u212A", true},
	{"abcdefghijK", "abcdefghij\u212A", true},
	{"abcdefghijkz", "abcdefghij\u212Ay", false},
	{"abcdefghijKz", "abcdefghij\u212Ay", false},
}

func TestEqualFold(t *testing.T) {
	for _, tt := range EqualFoldTests {
		if out := EqualFold([]byte(tt.s), []byte(tt.t)); out != tt.out {
			t.Errorf("EqualFold(%#q, %#q) = %v, want %v", tt.s, tt.t, out, tt.out)
		}
		if out := EqualFold([]byte(tt.t), []byte(tt.s)); out != tt.out {
			t.Errorf("EqualFold(%#q, %#q) = %v, want %v", tt.t, tt.s, out, tt.out)
		}
	}
}

func TestBufferGrowNegative(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatal("Grow(-1) should have panicked")
		}
	}()
	var b Buffer
	b.Grow(-1)
}

func TestBufferTruncateNegative(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatal("Truncate(-1) should have panicked")
		}
	}()
	var b Buffer
	b.Truncate(-1)
}

func TestBufferTruncateOutOfRange(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatal("Truncate(20) should have panicked")
		}
	}()
	var b Buffer
	b.Write(make([]byte, 10))
	b.Truncate(20)
}

var containsTests = []struct {
	b, subslice []byte
	want        bool
}{
	{[]byte("hello"), []byte("hel"), true},
	{[]byte("日本語"), []byte("日本"), true},
	{[]byte("hello"), []byte("Hello, world"), false},
	{[]byte("東京"), []byte("京東"), false},
}

func TestContains(t *testing.T) {
	for _, tt := range containsTests {
		if got := Contains(tt.b, tt.subslice); got != tt.want {
			t.Errorf("Contains(%q, %q) = %v, want %v", tt.b, tt.subslice, got, tt.want)
		}
	}
}

var ContainsAnyTests = []struct {
	b        []byte
	substr   string
	expected bool
}{
	{[]byte(""), "", false},
	{[]byte(""), "a", false},
	{[]byte(""), "abc", false},
	{[]byte("a"), "", false},
	{[]byte("a"), "a", true},
	{[]byte("aaa"), "a", true},
	{[]byte("abc"), "xyz", false},
	{[]byte("abc"), "xcz", true},
	{[]byte("a☺b☻c☹d"), "uvw☻xyz", true},
	{[]byte("aRegExp*"), ".(|)*+?^$[]", true},
	{[]byte(dots + dots + dots), " ", false},
}

func TestContainsAny(t *testing.T) {
	for _, ct := range ContainsAnyTests {
		if ContainsAny(ct.b, ct.substr) != ct.expected {
			t.Errorf("ContainsAny(%s, %s) = %v, want %v",
				ct.b, ct.substr, !ct.expected, ct.expected)
		}
	}
}

var ContainsRuneTests = []struct {
	b        []byte
	r        rune
	expected bool
}{
	{[]byte(""), 'a', false},
	{[]byte("a"), 'a', true},
	{[]byte("aaa"), 'a', true},
	{[]byte("abc"), 'y', false},
	{[]byte("abc"), 'c', true},
	{[]byte("a☺b☻c☹d"), 'x', false},
	{[]byte("a☺b☻c☹d"), '☻', true},
	{[]byte("aRegExp*"), '*', true},
}

func TestContainsRune(t *testing.T) {
	for _, ct := range ContainsRuneTests {
		if ContainsRune(ct.b, ct.r) != ct.expected {
			t.Errorf("ContainsRune(%q, %q) = %v, want %v",
				ct.b, ct.r, !ct.expected, ct.expected)
		}
	}
}

var makeFieldsInput = func() []byte {
	x := make([]byte, 1<<20)
	// Input is ~10% space, ~10% 2-byte UTF-8, rest ASCII non-space.
	for i := range x {
		switch rand.Intn(10) {
		case 0:
			x[i] = ' '
		case 1:
			if i > 0 && x[i-1] == 'x' {
				copy(x[i-1:], "χ")
				break
			}
			fallthrough
		default:
			x[i] = 'x'
		}
	}
	return x
}

var makeFieldsInputASCII = func() []byte {
	x := make([]byte, 1<<20)
	// Input is ~10% space, rest ASCII non-space.
	for i := range x {
		if rand.Intn(10) == 0 {
			x[i] = ' '
		} else {
			x[i] = 'x'
		}
	}
	return x
}

var bytesdata = []struct {
	name string
	data []byte
}{
	{"ASCII", makeFieldsInputASCII()},
	{"Mixed", makeFieldsInput()},
}

func BenchmarkFields(b *testing.B) {
	for _, sd := range bytesdata {
		b.Run(sd.name, func(b *testing.B) {
			for j := 1 << 4; j <= 1<<20; j <<= 4 {
				b.Run(fmt.Sprintf("%d", j), func(b *testing.B) {
					b.ReportAllocs()
					b.SetBytes(int64(j))
					data := sd.data[:j]
					for i := 0; i < b.N; i++ {
						Fields(data)
					}
				})
			}
		})
	}
}

func BenchmarkFieldsFunc(b *testing.B) {
	for _, sd := range bytesdata {
		b.Run(sd.name, func(b *testing.B) {
			for j := 1 << 4; j <= 1<<20; j <<= 4 {
				b.Run(fmt.Sprintf("%d", j), func(b *testing.B) {
					b.ReportAllocs()
					b.SetBytes(int64(j))
					data := sd.data[:j]
					for i := 0; i < b.N; i++ {
						FieldsFunc(data, unicode.IsSpace)
					}
				})
			}
		})
	}
}

func BenchmarkTrimSpace(b *testing.B) {
	s := []byte("  Some text.  \n")
	for i := 0; i < b.N; i++ {
		TrimSpace(s)
	}
}

func makeBenchInputHard() []byte {
	tokens := [...]string{
		"<a>", "<p>", "<b>", "<strong>",
		"</a>", "</p>", "</b>", "</strong>",
		"hello", "world",
	}
	x := make([]byte, 0, 1<<20)
	for {
		i := rand.Intn(len(tokens))
		if len(x)+len(tokens[i]) >= 1<<20 {
			break
		}
		x = append(x, tokens[i]...)
	}
	return x
}

var benchInputHard = makeBenchInputHard()

func BenchmarkSplitEmptySeparator(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Split(benchInputHard, nil)
	}
}

func BenchmarkSplitSingleByteSeparator(b *testing.B) {
	sep := []byte("/")
	for i := 0; i < b.N; i++ {
		Split(benchInputHard, sep)
	}
}

func BenchmarkSplitMultiByteSeparator(b *testing.B) {
	sep := []byte("hello")
	for i := 0; i < b.N; i++ {
		Split(benchInputHard, sep)
	}
}

func BenchmarkSplitNSingleByteSeparator(b *testing.B) {
	sep := []byte("/")
	for i := 0; i < b.N; i++ {
		SplitN(benchInputHard, sep, 10)
	}
}

func BenchmarkSplitNMultiByteSeparator(b *testing.B) {
	sep := []byte("hello")
	for i := 0; i < b.N; i++ {
		SplitN(benchInputHard, sep, 10)
	}
}

func BenchmarkRepeat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Repeat([]byte("-"), 80)
	}
}

func BenchmarkBytesCompare(b *testing.B) {
	for n := 1; n <= 2048; n <<= 1 {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			var x = make([]byte, n)
			var y = make([]byte, n)

			for i := 0; i < n; i++ {
				x[i] = 'a'
			}

			for i := 0; i < n; i++ {
				y[i] = 'a'
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Compare(x, y)
			}
		})
	}
}

func BenchmarkIndexAnyASCII(b *testing.B) {
	x := Repeat([]byte{'#'}, 4096) // Never matches set
	cs := "0123456789abcdef"
	for k := 1; k <= 4096; k <<= 4 {
		for j := 1; j <= 16; j <<= 1 {
			b.Run(fmt.Sprintf("%d:%d", k, j), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					IndexAny(x[:k], cs[:j])
				}
			})
		}
	}
}

func BenchmarkTrimASCII(b *testing.B) {
	cs := "0123456789abcdef"
	for k := 1; k <= 4096; k <<= 4 {
		for j := 1; j <= 16; j <<= 1 {
			b.Run(fmt.Sprintf("%d:%d", k, j), func(b *testing.B) {
				x := Repeat([]byte(cs[:j]), k) // Always matches set
				for i := 0; i < b.N; i++ {
					Trim(x[:k], cs[:j])
				}
			})
		}
	}
}

func BenchmarkIndexPeriodic(b *testing.B) {
	key := []byte{1, 1}
	for _, skip := range [...]int{2, 4, 8, 16, 32, 64} {
		b.Run(fmt.Sprintf("IndexPeriodic%d", skip), func(b *testing.B) {
			buf := make([]byte, 1<<16)
			for i := 0; i < len(buf); i += skip {
				buf[i] = 1
			}
			for i := 0; i < b.N; i++ {
				Index(buf, key)
			}
		})
	}
}
