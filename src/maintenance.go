package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

// maintenance runs
func (ctx *Instance) Tidy() {
	// Do they want to reset everything?
	ctx.MaybeInitialize()

	// Skip if we've been disabled
	if _, err := os.Stat(ctx.StatePath("disabled")); err == nil {
		log.Print("disabled file found, suspending maintenance")
		return
	}

	// Skip if we've expired
	untilspec, err := ioutil.ReadFile(ctx.StatePath("until"))
	if err == nil {
		until, err := time.Parse(time.RFC3339, string(untilspec))
		if err != nil {
			log.Printf("Unparseable date in until file: %v", until)
		} else {
			if until.Before(time.Now()) {
				log.Print("until file time reached, suspending maintenance")
				return
			}
		}
	}

	// Any new categories?
	files, err := ioutil.ReadDir(ctx.MothballPath())
	if err != nil {
		log.Printf("Error listing mothballs: %s", err)
	}
	for _, f := range files {
		filename := f.Name()
		filepath := ctx.MothballPath(filename)
		if !strings.HasSuffix(filename, ".mb") {
			continue
		}
		categoryName := strings.TrimSuffix(filename, ".mb")

		if _, ok := ctx.Categories[categoryName]; !ok {
			mb, err := OpenMothball(filepath)
			if err != nil {
				log.Printf("Error opening %s: %s", filepath, err)
				continue
			}
			log.Printf("New category: %s", filename)
			ctx.Categories[categoryName] = mb
		}
	}

	// Any old categories?
	log.Print("XXX: Check for and reap old categories")

	ctx.CollectPoints()
}

// collectPoints gathers up files in points.new/ and appends their contents to points.log,
// removing each points.new/ file as it goes.
func (ctx *Instance) CollectPoints() {
	logf, err := os.OpenFile(ctx.StatePath("points.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Can't append to points log: %s", err)
		return
	}
	defer logf.Close()

	files, err := ioutil.ReadDir(ctx.StatePath("points.new"))
	if err != nil {
		log.Printf("Error reading packages: %s", err)
	}
	for _, f := range files {
		filename := ctx.StatePath("points.new", f.Name())
		s, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Printf("Can't read points file %s: %s", filename, err)
			continue
		}
		award, err := ParseAward(string(s))
		if err != nil {
			log.Printf("Can't parse award file %s: %s", filename, err)
			continue
		}
		fmt.Fprintf(logf, "%s\n", award.String())
		log.Print("XXX: check for duplicates", award.String())
		logf.Sync()
		if err := os.Remove(filename); err != nil {
			log.Printf("Unable to remove %s: %s", filename, err)
		}
	}
}

// maintenance is the goroutine that runs a periodic maintenance task
func (ctx *Instance) Maintenance(maintenanceInterval time.Duration) {
	for ; ; time.Sleep(maintenanceInterval) {
		ctx.Tidy()
	}
}
