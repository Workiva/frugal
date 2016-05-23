package test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Workiva/frugal/compiler"
	"github.com/Workiva/frugal/compiler/globals"
)

func TestValidJava(t *testing.T) {
	defer globals.Reset()
	nowBefore := globals.Now
	defer func() {
		globals.Now = nowBefore
	}()
	globals.Now = time.Date(2015, 11, 24, 0, 0, 0, 0, time.UTC)

	options := compiler.Options{
		File:  validFile,
		Gen:   "java",
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	pubPath := filepath.Join(outputDir, "foo", "FooPublisher.java")
	compareFiles(t, "expected/java/FooPublisher.java", pubPath)
	subPath := filepath.Join(outputDir, "foo", "FooSubscriber.java")
	compareFiles(t, "expected/java/FooSubscriber.java", subPath)
	pubPath = filepath.Join(outputDir, "foo", "BlahPublisher.java")
	compareFiles(t, "expected/java/BlahPublisher.java", pubPath)
	subPath = filepath.Join(outputDir, "foo", "BlahSubscriber.java")
	compareFiles(t, "expected/java/BlahSubscriber.java", subPath)
	servicePath := filepath.Join(outputDir, "foo", "FBlah.java")
	compareFiles(t, "expected/java/FBlah.java", servicePath)
}

func TestValidJavaWithAsync(t *testing.T) {
	defer globals.Reset()
	nowBefore := globals.Now
	defer func() {
		globals.Now = nowBefore
	}()
	globals.Now = time.Date(2015, 11, 24, 0, 0, 0, 0, time.UTC)

	options := compiler.Options{
		File:  validFile,
		Gen:   "java:async",
		Out:   outputDir,
		Delim: delim,
	}
	if err := compiler.Compile(options); err != nil {
		t.Fatal("Unexpected error", err)
	}

	servicePath := filepath.Join(outputDir, "foo", "FBlah.java")
	compareFiles(t, "expected/java/FBlah_async.java", servicePath)
}
