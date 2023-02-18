package cache

import "fmt"

const DLMT = "\t"

func EntryToText(title string, body []byte) string {
	return fmt.Sprintf("%s%s%s\n", title, DLMT, body)
}
