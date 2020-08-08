package base64

import (
	"errors"
)

var base64 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

func Encode(input []byte) ([]byte, error) {
	var output []byte
	len := len(input)

	idx := 0
	for ; len > 0; {
		output = append(output, base64[(input[idx + 0] & 0xFC) >> 2])

		if len == 1 {
			output = append(output, base64[(input[idx + 0] & 0x03) << 4])
			output = append(output, '=')
			output = append(output, '=')
			break
		}

		output = append(output, base64[((input[idx + 0] & 0x03) << 4) | ((input[idx + 1] & 0xF0) >> 4)])

		if len == 2 {
			output = append(output, base64[(input[idx + 1] & 0x0F) << 2])
			output = append(output, '=')
			break;
		}

		output = append(output, base64[((input[idx + 1] & 0x0F) << 2) | ((input[idx + 2] & 0xC0) >> 6)])
		output = append(output, base64[(input[idx + 2] & 0x3F)])

		len -= 3
		idx += 3
	}
	return output, nil
}

var unbase64 = [...]int {
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 62, -1, -1, -1, 63, 52,
  53, 54, 55, 56, 57, 58, 59, 60, 61, -1, -1, -1, 0, -1, -1, -1,
  0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
  16, 17, 18, 19, 20, 21, 22, 23, 24, 25, -1, -1, -1, -1, -1, -1,
  26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41,
  42, 43, 44, 45, 46, 47, 48, 49, 50, 51, -1, -1, -1, -1, -1, -1,
}

func Decode(input []byte) ([]byte, error) {
	var output []byte
	len := len(input)
	idx := 0
	
	for {
		for i := 0; i <= 3; i++ {
			if input[i] > 128 || unbase64[input[i]] == -1 {
				return nil, errors.New("invalid character for base64 encoding: " + string(input[i]))
			}
		}
		output = append(output, byte((unbase64[input[idx + 0]] << 2) | ((unbase64[idx + 1] & 0x30) >> 4)))

		if input[idx + 2] != '=' {
			output = append(output, byte(((unbase64[input[idx + 1]] & 0x0F) << 4) | ((unbase64[input[idx + 3]] & 0x3C) >> 2)))
		}

		if input[idx + 3] != '=' {
			output = append(output, byte(((unbase64[input[idx + 2]] & 0x03) << 6) | unbase64[input[idx + 3]]))
		}
		idx += 4
		len -= 4
	}
	return output, nil
}