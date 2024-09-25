package builder

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	gotestUtil "github.com/OpenTestSolar/testtool-golang-gotest/pkg/util"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/conc/pool"
)

const (
	MaxBuildConcurrency  = 2
	MaxExecCmdRetry      = 3
	ExecCmdRetryInterval = 2 * time.Second
)

func Build(projPath string) error {
	var packageList []string
	var err error
	filepath.Walk(projPath, func(path string, fi os.FileInfo, _ error) error {
		if strings.HasSuffix(path, "_test.go") {
			packagePath := filepath.Dir(path)
			if packagePath == projPath {
				packagePath = ""
			} else {
				packagePath, err = filepath.Rel(projPath, packagePath)
				if err != nil {
					return err
				}
			}
			if !gotestUtil.ElementIsInSlice(packagePath, packageList) {
				packageList = append(packageList, packagePath)
			}
		}
		return nil
	})

	concBuild, _ := strconv.ParseBool(os.Getenv("TESTSOlAR_TTP_CONCURRENTBUILD"))
	concurrencyLevel := 1
	if concBuild == true {
		log.Printf("[PLUGIN]enable concurrent building package")
		if runtime.GOMAXPROCS(0)+2 > MaxBuildConcurrency {
			concurrencyLevel = MaxBuildConcurrency
		} else {
			concurrencyLevel = runtime.GOMAXPROCS(0) + 2
		}
		log.Printf("[PLUGIN]Build package concurrency level %d", concurrencyLevel)
	}
	p := pool.New().WithMaxGoroutines(concurrencyLevel).WithErrors().WithFirstError()
	compress, _ := strconv.ParseBool(os.Getenv("TESTSOLAR_TTP_COMPRESSBINARY"))
	log.Printf("[PLUGIN]compress binaries %v", compress)
	for _, packagePath := range packageList {
		log.Printf("[PLUGIN]Build package %s", packagePath)
		p.Go(func() error {
			return buildAndCompressTestBin(projPath, packagePath, compress)
		})
	}
	err = p.Wait()
	if err != nil {
		return errors.Wrapf(err, "build package failed")
	}
	return nil
}

func compressBinFile(projPath, pkgBin string) error {
	if _, err := exec.LookPath("upx"); err != nil {
		log.Println("in order to compress binaries, upx need to be installed")
		return err
	}
	cmdline := fmt.Sprintf("upx -1 -f %s", pkgBin)
	log.Printf("[PLUGIN]compress cmdline: %s", cmdline)
	startTime := time.Now()
	_, stderr, err := gotestUtil.RunCommandWithOutput(cmdline, projPath)
	if err != nil {
		log.Printf("[PLUGIN]compress %s failed: %s, err: %s", pkgBin, stderr, err.Error())
		return err
	}
	endTime := time.Now()
	delta := endTime.Sub(startTime)
	_, err = os.Stat(pkgBin)
	if err != nil {
		log.Printf("[PLUGIN]can't find %s failed: %s, err: %s", pkgBin, stderr, err.Error())
		return err
	}
	log.Printf("[PLUGIN]Run compress command cost %.2fs", delta.Seconds())
	return nil
}

func buildAndCompressTestBin(projPath string, packagePath string, compress bool) error {
	startTime := time.Now()
	pkgBin, err := BuildTestPackage(projPath, packagePath, compress)
	if err != nil {
		log.Printf("[PLUGIN]Build package %s failed, err: %s", packagePath, err.Error())
		return err
	}
	endTime := time.Now()
	log.Printf("[PLUGIN]Run compile command cost %.2fs", endTime.Sub(startTime).Seconds())
	if compress {
		err := compressBinFile(projPath, pkgBin)
		if err != nil {
			log.Printf("[PLUGIN]Compress bin file %s failed, err: %s", pkgBin, err.Error())
		}
	}
	return nil
}

func BuildTestPackage(projPath string, packagePath string, compress bool) (string, error) {
	pkgBin := filepath.Join(projPath, packagePath+".test")
	cmdline := ""
	if compress {
		cmdline = fmt.Sprintf("go test -ldflags=\"-s -w\" -c ./%s -o %s", packagePath, pkgBin)
	} else {
		cmdline = fmt.Sprintf("go test -c ./%s -o %s", packagePath, pkgBin)
	}
	log.Printf("[PLUGIN]Build package %s by cmd: %s", packagePath, cmdline)
	err := retry.Do(
		func() error {
			_, stderr, err := gotestUtil.RunCommandWithOutput(cmdline, projPath)
			if err != nil {
				log.Printf("[PLUGIN]Build package %s failed: %s, err: %s", packagePath, stderr, err.Error())
				return err
			}
			_, err = os.Stat(pkgBin)
			if err != nil {
				log.Printf("[PLUGIN]Can't find bin file: %s, stderr: %s, err: %s", pkgBin, stderr, err.Error())
				return err
			}
			return nil
		},
		retry.Attempts(MaxExecCmdRetry),
		retry.Delay(ExecCmdRetryInterval),
	)
	if err != nil {
		log.Printf("[PLUGIN]Build package %s failed, err: %s", packagePath, err.Error())
		return "", err
	}
	_, err = os.Stat(pkgBin)
	if err != nil {
		log.Printf("[PLUGIN]Stat build bin file %s during running failed, err: %v", pkgBin, err)
		return "", err
	}
	err = os.Chmod(pkgBin, 0777)
	if err != nil {
		log.Printf("[PLUGIN]Change bin file %s mode failed, err: %v", pkgBin, err)
		return "", err
	}
	return pkgBin, nil
}
