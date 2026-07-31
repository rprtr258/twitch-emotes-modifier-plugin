package internal

import (
	"image"
	"iter"
)

type mergedTimestamp struct {
	Timestamp int
	Frames    []int
}

func mergeTimeSeries(first, second []int) []mergedTimestamp {
	// if second is static
	if len(second) == 1 && second[0] == 0 {
		res := make([]mergedTimestamp, 0, len(first))
		for i, ts := range first {
			res = append(res, mergedTimestamp{
				Timestamp: ts,
				Frames:    []int{i, 0},
			})
		}
		return res
	}

	res := make([]mergedTimestamp, 0, len(first)+len(second))
	i, j := 0, 0
	secondOffset := 0
	for i < len(first) {
		var m mergedTimestamp
		switch {
		case first[i] < second[j]+secondOffset:
			m = mergedTimestamp{
				Timestamp: first[i],
				Frames:    []int{i, j},
			}
			i++
		case first[i] > second[j]+secondOffset:
			m = mergedTimestamp{
				Timestamp: second[j] + secondOffset,
				Frames:    []int{i, j},
			}
			j++
			if j == len(second) {
				j = 0
				secondOffset += second[len(second)-1]
			}
		case first[i] == second[j]+secondOffset:
			m = mergedTimestamp{
				Timestamp: first[i],
				Frames:    []int{i, j},
			}
			i++
			j++
			if j == len(second) {
				j = 0
				secondOffset += second[len(second)-1]
			}
		}
		res = append(res, m)
	}
	return res
}

func MergeTimeSeries(first, second []int) []mergedTimestamp {
	if len(first) == 0 || len(second) == 0 {
		panic("time series must not be empty")
	}

	if first[len(first)-1] < second[len(second)-1] {
		res := mergeTimeSeries(second, first)
		for i := range res {
			res[i].Frames = []int{
				res[i].Frames[1],
				res[i].Frames[0],
			}
		}
		return res
	}

	return mergeTimeSeries(first, second)
}

func ReverseTimestamps(timestamps []int) []int {
	if len(timestamps) < 2 {
		return timestamps
	}

	n := len(timestamps)

	res := make([]int, 0, n)
	res = append(
		res,
		timestamps[n-1]-timestamps[n-2],
	)
	for i := 1; i < n-1; i++ {
		res = append(
			res,
			res[i-1]+timestamps[n-i-1]-timestamps[n-i-2],
		)
	}
	res = append(
		res,
		res[n-2]+timestamps[0],
	)
	return res
}

func DelaysToTimestamps(delays []int) []int {
	timestamps := make([]int, len(delays))
	sum := 0
	for i, d := range delays {
		sum += d
		timestamps[i] = sum
	}
	return timestamps
}

// RGBA returns the frame as an RGBA image.
// webp.DecodeAll decodes every frame as *image.RGBA,
// so this assertion always holds for emotes loaded via the repository.
func RGBA(img image.Image) *image.RGBA {
	return img.(*image.RGBA)
}

func ToRGBAs(frames []image.Image) iter.Seq2[int, *image.RGBA] {
	return func(yield func(int, *image.RGBA) bool) {
		for i, frame := range frames {
			if !yield(i, RGBA(frame)) {
				return
			}
		}
	}
}
