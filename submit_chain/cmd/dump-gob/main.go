package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/google/go-github/v28/github"
	"github.com/kr/pretty"
)

func main() {
	var (
		login = flag.String("login", "", "Login of the user to submit for.")
	)
	flag.Parse()
	if *login == "" {
		glog.Fatal("You need to specify a non-empty value for the flags `-login`")
	}

	r := bufio.NewReader(os.Stdin)
	enc := gob.NewDecoder(r)
	var prs []*github.PullRequest
	if err := enc.Decode(&prs); err != nil {
		glog.Fatalf("failed to decode prs: %v", err)
	}

	for _, pr := range prs {
		fmt.Printf("%# v\n", pretty.Formatter(*pr))
	}
}
