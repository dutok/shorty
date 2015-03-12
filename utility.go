package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"os/exec"
	"strconv"
)

func incrementCount(alias string, db *bolt.DB) error {
	var totalcount []byte
	var url *URLJson
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("urls"))
		urljson := b.Get([]byte(alias))
		err := json.Unmarshal(urljson, &url)
		if err != nil {
			fmt.Println("error:", err)
		}
		count := *url.Count + 1
		*url.Count = count

		urljson, err = json.Marshal(&url)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		err = b.Put([]byte(alias), urljson)

		b = tx.Bucket([]byte("stats"))
		totalcount = b.Get([]byte("totalcount"))
		if totalcount == nil {
			err = b.Put([]byte("totalcount"), []byte("0"))
			if err != nil {
				fmt.Println(err)
			}
		} else {
			newcount, err := incStr((string(totalcount)))
			if err != nil {
				fmt.Println(err)
			}
			err = b.Put([]byte("totalcount"), []byte(newcount))
			if err != nil {
				fmt.Println(err)
			}
		}

		return nil
	})
	return nil
}

func incStr(count string) (string, error) {
	i, err := strconv.Atoi(count)
	if err != nil {
		return count, err
	}
	t := i + 1
	return strconv.Itoa(t), nil
}

func loadVersion() string {
	out, err := exec.Command("sh", "-c", "git describe --long --tags ").Output()
	if err != nil {
		fmt.Println(err)
		return ""
	}
  return string(out)
}
