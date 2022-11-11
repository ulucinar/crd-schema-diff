// HO.

package main

import (
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/tufin/oasdiff/diff"
	"github.com/tufin/oasdiff/report"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
		app             = kingpin.New("crddiff", "A tool for checking breaking API changes between two CRD OpenAPI v3 schemas").DefaultEnvars()
		baseCRDPath     = app.Arg("base", "The manifest file path of the CRD to be used as the base").Required().ExistingFile()
		revisionCRDPath = app.Arg("revision", "The manifest file path of the CRD to be used as a revision to the base").Required().ExistingFile()
	)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	baseCRD, err := getCRD(*baseCRDPath)
	if err != nil {
		panic(err)
	}
	baseDoc, err := getOpenAPIv3Document(baseCRD)
	if err != nil {
		panic(err)
	}
	revisionCRD, err := getCRD(*revisionCRDPath)
	if err != nil {
		panic(err)
	}
	revisionDoc, err := getOpenAPIv3Document(revisionCRD)
	if err != nil {
		panic(err)
	}

	d, err := getBreakingChanges(baseDoc, revisionDoc)
	if err != nil {
		panic(err)
	}
	if d.Empty() {
		return
	}
	l := log.New(os.Stderr, "", 0)
	l.Println(getDiffReport(d))
	syscall.Exit(1)
}

func getDiffReport(d *diff.Diff) string {
	l := strings.Split(report.GetTextReportAsString(d), "\n")
	l = l[12:]
	for i, s := range l {
		l[i] = strings.TrimPrefix(s, "      ")
	}
	return strings.Join(l, "\n")
}
