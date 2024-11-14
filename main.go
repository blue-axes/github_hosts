package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	builtinHostUrlList = []string{
		"https://raw.githubusercontent.com/521xueweihan/GitHub520/refs/heads/main/hosts.json",
		"https://gitee.com/a-little-teaser/GitHub520/raw/main/hosts.json",
		"https://gitee.com/chujin_w/GitHub520/raw/main/hosts.json",
		"https://gitee.com/xnjy/github520/raw/main/hosts.json",
	}
	startMark = "########## start github_hosts ##########"
	endMark   = "########## end github_hosts ##########"
)

func main() {
	var (
		hostsUrl = "https://raw.githubusercontent.com/521xueweihan/GitHub520/refs/heads/main/hosts.json"
		mode     = "append"
		timeout  = time.Second * 30
	)
	flag.StringVar(&hostsUrl, "url", hostsUrl, "the hosts url")
	flag.StringVar(&mode, "mode", mode, "the rewrite mode. append or rewrite")
	flag.Parse()
	builtinHostUrlList = append([]string{hostsUrl}, builtinHostUrlList...)
	var (
		httpClient = http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			CheckRedirect: http.DefaultClient.CheckRedirect,
			Jar:           http.DefaultClient.Jar,
			Timeout:       timeout,
		}

		resList = make([][]string, 0)
	)

	for _, hostsUrl := range builtinHostUrlList {
		resp, err := httpClient.Get(hostsUrl)
		if err != nil {
			fmt.Printf("[%s] get content err:%s\n", hostsUrl, err)
			continue
		}
		dec := json.NewDecoder(resp.Body)
		dec.UseNumber()
		err = dec.Decode(&resList)
		_ = resp.Body.Close()
		if err != nil {
			fmt.Printf("[%s] parse content err:%s\n", hostsUrl, err)
			continue
		}
		if len(resList) > 0 {
			break
		}
	}

	fmt.Printf("hosts file path:%s\n", hostsFilePath)

	switch strings.ToLower(mode) {
	case "append":
		oldList := parseLocalHostsFile()
		newList := append(oldList, []string{startMark})
		newList = append(newList, resList...)
		newList = append(newList, []string{endMark})
		err := writeLocalHostsFile(newList)
		if err != nil {
			fmt.Printf("write hosts file err:%s\n", err)
			time.Sleep(time.Second * 10)
			return
		}
	case "rewrite":
		fallthrough
	default:
		newList := [][]string{
			{startMark},
		}
		newList = append(newList, resList...)
		newList = append(newList, []string{endMark})
		err := writeLocalHostsFile(newList)
		if err != nil {
			fmt.Printf("write hosts file err:%s\n", err)
			time.Sleep(time.Second * 10)
			return
		}
	}
}

func parseLocalHostsFile() [][]string {
	frd, err := os.OpenFile(hostsFilePath, os.O_RDONLY, 0666)
	if err != nil {
		return nil
	}
	defer frd.Close()
	reader := bufio.NewReader(frd)
	var (
		line      = ""
		commentRe = regexp.MustCompile(`^\s*#\s*`)
		splitRe   = regexp.MustCompile(`\s+`)
		result    = make([][]string, 0)
		ignore    = false
	)
	for {
		currentLine, hasPrefix, err := reader.ReadLine()
		if err != nil {
			break
		}
		if hasPrefix {
			line += string(currentLine)
			continue
		} else {
			line += string(currentLine)
		}
		if line == "# "+startMark {
			// 忽略
			ignore = true
		} else if line == "# "+endMark {
			ignore = false
			continue
		}
		if ignore {
			line = ""
			continue
		}

		if commentRe.MatchString(line) {
			result = append(result, []string{commentRe.ReplaceAllString(line, "")})
			line = ""
			continue
		}
		row := splitRe.Split(line, 2)
		if len(row) < 2 {
			line = ""
			continue
		}
		result = append(result, []string{row[0], row[1]})
		line = ""
	}
	return result
}

func writeLocalHostsFile(items [][]string) error {
	fwd, err := os.OpenFile(hostsFilePath, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fwd.Close()
	var (
		content = ""
	)
	for _, item := range items {
		if len(item) == 0 {
			continue
		}
		if len(item) != 2 {
			content += fmt.Sprintf("# %s"+eol, strings.Join(item, ","))
		} else {
			content += fmt.Sprintf("%s\t%s"+eol, item[0], item[1])
		}
	}
	_, err = fwd.WriteString(content)
	return err
}
