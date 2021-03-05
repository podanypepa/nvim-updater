package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

var sources = map[string]string{
	"darwin": "https://github.com/neovim/neovim/releases/download/nightly/nvim-macos.tar.gz",
	"linux":  "https://github.com/neovim/neovim/releases/download/nightly/nvim-linux64.tar.gz",
}

var binDir = ""

func init() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	binDir = usr.HomeDir + "/bin"
}

func main() {
	versionOnMyMac, err := getInstalledNvimVersion()
	if err != nil {
		log.Fatal(err)
	}

	latestNightBuild, err := getLatestNightBuildVer()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("on my pc:           ", versionOnMyMac)
	fmt.Println("latest night build: ", latestNightBuild)

	if versionOnMyMac == latestNightBuild {
		fmt.Println(colorGreen + "OK" + colorReset + " you are up to date")
		return
	}

	fmt.Println(colorYellow + "yes! " + colorReset + " new nvim version is here!")

	newNvimtar, err := wgetNewNvim()
	if err != nil {
		log.Fatal(err)
	}
	installNewVersion(newNvimtar)
}

func wgetNewNvim() (downloadedFile string, err error) {
	fmt.Println("downloading new nvim version")

	src := sources[runtime.GOOS]

	cmd := exec.Command("wget", src, "-P", binDir)
	if err = cmd.Run(); err != nil {
		return "", err
	}

	downloadedFile = binDir + "/" + string(src[strings.LastIndex(src, "/")+1:])

	return downloadedFile, nil
}

func installNewVersion(tarFile string) (err error) {
	fmt.Println("installing new nvim version")

	cmd := exec.Command("tar", "xzvf", tarFile, "-C", binDir)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return os.Remove(tarFile)
}

func getLatestNightBuildVer() (version string, err error) {
	verURL := "https://github.com/neovim/neovim/releases"
	res, err := http.Get(verURL)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	version = doc.
		Find(".release-header .f1").
		First().
		Find("a").
		Text()

	return version, nil
}

func getInstalledNvimVersion() (version string, err error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("nvim", "--version")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		return "", err
	}

	lines, err := stringToLines(stdout.String())
	if err != nil {
		return "", err
	}

	return lines[0], nil
}

func stringToLines(s string) (lines []string, err error) {
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	err = scanner.Err()
	return lines, err
}
