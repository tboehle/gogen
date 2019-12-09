package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/tboehle/gogen/unmarshalmap"
)

var (
	// gogen.exe --o "unmarshal.go" --o-test "unmarshal_test.go" --pkg "github.com/tboehle/gogentest" GoGenTester
	out  = flag.String("o", "", "what file to write")
	tOut = flag.String("o-test", "", "what file to write the test to")
	pkg  = flag.String("pkg", "", "what package to get the interface from")
)

func main() {
	flag.Parse()

	st := flag.Arg(0)

	if st == "" {
		log.Fatal("need to specify a struct name")
	}

	if count := len(*pkg); count == 0 {
		module, err := workspaceDir()
		if err != nil {
			log.WithField("error", err).Fatal("Fatal error while getting workspace dir because of no given GOPATH package")
		}
		*pkg = module
	}

	log.WithFields(log.Fields{
		"outFile":        out,
		"testOutputFile": tOut,
		"pkgName":        pkg,
		"StructName":     st,
	}).Info()

	gen, err := unmarshalmap.NewGenerator(*pkg, st)
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer

	log.Printf("Generating func (*%s) UnmarshalMap(map[string]interface{}) error", st)
	err = gen.Write(&buf)
	if err != nil {
		log.Fatal(err)
	}

	if *out != "" {
		err := ioutil.WriteFile(*out, buf.Bytes(), 0666)
		if err != nil {
			log.Fatal(err)
		}
		if *tOut == "" {
			*tOut = fmt.Sprintf("%s_test.go", strings.TrimRight(*out, ".go"))
		}
	} else {
		fmt.Println(buf.String())
	}

	buf.Reset()

	err = gen.WriteTest(&buf)
	if err != nil {
		log.Fatal(err)
	}

	if *tOut != "" {
		err := ioutil.WriteFile(*tOut, buf.Bytes(), 0666)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println(buf.String())
	}
}

func workspaceDir() (string, error) {
	workspace, dirErr := os.Getwd()
	if dirErr != nil {
		return "", fmt.Errorf("Error while getting workspace")
	}
	if dir, err := os.Stat(workspace); err == nil && dir.IsDir() {
		// Prrove if Backslash or Slash
		return workspace, nil
	}
	return "", fmt.Errorf("Workspace is no directory or is not existing")
}
